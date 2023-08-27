package app

import (
	"fmt"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	ginLogger "github.com/rawen554/shortener/internal/middleware/logger"

	"github.com/rawen554/shortener/internal/middleware/auth"
	"github.com/rawen554/shortener/internal/middleware/compress"
)

func (a *App) SetupRouter() (*gin.Engine, error) {
	r := gin.New()
	pprof.Register(r)

	ginLoggerMiddleware, err := ginLogger.Logger(a.logger.Named("middleware"))
	if err != nil {
		return nil, fmt.Errorf("error creating middleware logger func: %w", err)
	}
	r.Use(ginLoggerMiddleware)
	r.Use(auth.AuthMiddleware(a.config.Secret))
	r.Use(compress.Compress())

	r.GET("/:id", a.RedirectToOriginal)
	r.POST("/", a.ShortenURL)
	r.GET("/ping", a.Ping)

	api := r.Group("/api")
	{
		api.POST("/shorten", a.ShortenURL)
		api.POST("/shorten/batch", a.ShortenBatch)

		api.GET("/user/urls", a.GetUserRecords)
		api.DELETE("/user/urls", a.DeleteUserRecords)
	}

	return r, nil
}
