package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/rawen554/shortener/internal/app"
	"github.com/rawen554/shortener/internal/compress"
	"github.com/rawen554/shortener/internal/config"
	ginLogger "github.com/rawen554/shortener/internal/logger"
	"github.com/rawen554/shortener/internal/store"
)

func setupRouter(a *app.App) *gin.Engine {
	r := gin.New()
	ginLoggerMiddleware, err := ginLogger.Logger()
	if err != nil {
		log.Fatal(err)
	}
	r.Use(ginLoggerMiddleware)
	r.Use(compress.Compress())

	r.GET("/:id", a.RedirectToOriginal)
	r.POST("/", a.ShortenURL)
	r.GET("/ping", a.DBHealthCheck)

	api := r.Group("/api")
	{
		api.POST("/shorten", a.ShortenURL)
		api.POST("/shorten/batch", a.ShortenBatch)
	}

	return r
}

func main() {
	config, err := config.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	storage, err := store.NewStore(config)
	if err != nil {
		log.Fatal(err)
	}

	app := app.NewApp(config, storage)

	r := setupRouter(app)
	log.Fatal(r.Run(config.FlagRunAddr))
}
