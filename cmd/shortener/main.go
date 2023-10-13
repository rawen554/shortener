package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rawen554/shortener/internal/app"
	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/handlers"
	pb "github.com/rawen554/shortener/internal/handlers/proto"
	"github.com/rawen554/shortener/internal/logger"
	"github.com/rawen554/shortener/internal/logic"
	"github.com/rawen554/shortener/internal/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

const (
	timeoutServerShutdown = time.Second * 5
	timeoutShutdown       = time.Second * 10
)

func main() {
	ctx, cancelCtx := signal.NotifyContext(context.Background(), syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	logger, err := logger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}

	logger.Infof("Build version: %s", buildVersion)
	logger.Infof("Build date: %s", buildDate)
	logger.Infof("Build commit: %s", buildCommit)

	defer cancelCtx()

	config, err := config.ParseFlags()
	if err != nil {
		logger.Fatal(err)
	}

	storage, err := store.NewStore(ctx, config)
	if err != nil {
		logger.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	defer func() {
		wg.Wait()
	}()

	wg.Add(1)
	go func() {
		defer logger.Info("closed DB")
		defer wg.Done()
		<-ctx.Done()

		storage.Close()
	}()

	componentsErrs := make(chan error, 1)

	coreLogic := logic.NewCoreLogic(config, storage, logger.Named("logic"))
	a := app.NewApp(config, coreLogic, logger.Named("app"))

	r, err := a.SetupRouter()
	if err != nil {
		logger.Fatal(err)
	}
	srv := http.Server{
		Addr:    config.RunAddr,
		Handler: r,
	}

	go func(errs chan<- error) {
		if config.EnableHTTPS {
			_, errCert := os.ReadFile(config.TLSCertPath)
			_, errKey := os.ReadFile(config.TLSKeyPath)

			if errors.Is(errCert, os.ErrNotExist) || errors.Is(errKey, os.ErrNotExist) {
				privateKey, certBytes, err := app.CreateCertificates(logger.Named("certs-builder"))
				if err != nil {
					errs <- fmt.Errorf("error creating tls certs: %w", err)
					return
				}

				if err := app.WriteCertificates(certBytes, config.TLSCertPath, privateKey, config.TLSKeyPath, logger); err != nil {
					errs <- fmt.Errorf("error writing tls certs: %w", err)
					return
				}
			}

			if err := srv.ListenAndServeTLS(config.TLSCertPath, config.TLSKeyPath); err != nil {
				if errors.Is(err, http.ErrServerClosed) {
					return
				}
				errs <- fmt.Errorf("run tls server has failed: %w", err)
				return
			}
		}

		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			errs <- fmt.Errorf("run server has failed: %w", err)
		}
	}(componentsErrs)

	if config.GRPCPort != "" {
		wg.Add(1)
		go func(errs chan<- error) {
			defer wg.Done()
			lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GRPCPort))
			if err != nil {
				logger.Errorf("failed to listen: %w", err)
				errs <- err
				return
			}
			grpcServer := grpc.NewServer()
			reflection.Register(grpcServer)

			pb.RegisterShortenerServer(grpcServer, handlers.NewService(logger, coreLogic))

			logger.Infof("running gRPC service on %s", config.GRPCPort)

			if err = grpcServer.Serve(lis); err != nil {
				if errors.Is(err, grpc.ErrServerStopped) {
					return
				}
				errs <- err
			}
		}(componentsErrs)
	}

	wg.Add(1)
	go func() {
		defer logger.Info("server has been shutdown")
		defer wg.Done()
		<-ctx.Done()

		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), timeoutServerShutdown)
		defer cancelShutdownTimeoutCtx()
		if err := srv.Shutdown(shutdownTimeoutCtx); err != nil {
			logger.Errorf("an error occurred during server shutdown: %v", err)
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-componentsErrs:
		logger.Error(err)
		cancelCtx()
	}

	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		logger.Fatal("failed to gracefully shutdown the service")
	}()
}
