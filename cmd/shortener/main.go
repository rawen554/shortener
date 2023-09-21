package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/rawen554/shortener/internal/app"
	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/logger"
	"github.com/rawen554/shortener/internal/store"
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
	ctx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)

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

	a := app.NewApp(config, storage, logger.Named("app"))

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
			if err := app.CreateCertificates(); err != nil {
				errs <- fmt.Errorf("error creating tls certs: %w", err)
			}

			if err := srv.ListenAndServeTLS("./certs/cert.pem", "./certs/private.pem"); err != nil {
				if errors.Is(err, http.ErrServerClosed) {
					return
				}
				errs <- fmt.Errorf("run tls server has failed: %w", err)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil {
				if errors.Is(err, http.ErrServerClosed) {
					return
				}
				errs <- fmt.Errorf("run server has failed: %w", err)
			}
		}
	}(componentsErrs)

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
		log.Print(err)
		cancelCtx()
	}

	go func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		logger.Fatal("failed to gracefully shutdown the service")
	}()
}
