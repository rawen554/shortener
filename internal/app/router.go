package app

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/rawen554/shortener/internal/middleware/compress"
	ginLogger "github.com/rawen554/shortener/internal/middleware/logger"

	"github.com/rawen554/shortener/internal/middleware/auth"
)

func (a *App) SetupRouter() (*gin.Engine, error) {
	r := gin.New()
	pprof.Register(r)

	r.Use(ginLogger.Logger(a.logger.Named("middleware")))
	r.Use(auth.AuthMiddleware(a.config.Secret, a.logger.Named("auth_middleware")))
	r.Use(compress.Compress())

	r.GET("/:id", a.RedirectToOriginal)
	r.POST(rootPath, a.ShortenURL)
	r.GET(pingPath, a.Ping)

	api := r.Group("/api")
	{
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
