package app

import (
	"fmt"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/rawen554/shortener/internal/middleware/compress"
	ginLogger "github.com/rawen554/shortener/internal/middleware/logger"

	"github.com/rawen554/shortener/internal/middleware/auth"
)

func (a *App) SetupRouter() (*gin.Engine, error) {
	r := gin.New()
	if a.config.ProfileMode {
		pprof.Register(r)
	}

	authMiddleware, err := auth.NewAuthMiddleware(a.config.Secret, a.logger.Named("auth_middleware"))
	if err != nil {
		return nil, fmt.Errorf("error initializing auth middleware: %w", err)
	}

	subnetAuthMiddleware := auth.NewSubnetChecker(a.config.TrustedSubnet, a.logger.Named("subnet_middleware"))

	r.Use(ginLogger.Logger(a.logger.Named("middleware")))
	r.Use(authMiddleware)
	r.Use(compress.Compress())

	r.GET("/:id", a.RedirectToOriginal)
	r.POST(rootPath, a.ShortenURL)
	r.GET(pingPath, a.Ping)

	api := r.Group("/api")
	{
		internalAPI := api.Group("/internal")
		internalAPI.Use(subnetAuthMiddleware)
		{
			internalAPI.GET("/stats", a.GetStats)
		}

		shortenerAPI := api.Group("/shorten")
		{
			shortenerAPI.POST("", a.ShortenURL)
			shortenerAPI.POST("/batch", a.ShortenBatch)
		}

		userAPI := api.Group("/user/urls")
		{
			userAPI.GET("", a.GetUserRecords)
			userAPI.DELETE("", a.DeleteUserRecords)
		}
	}

	return r, nil
}
