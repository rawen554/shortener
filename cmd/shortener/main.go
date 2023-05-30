package main

import (
	"log"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/rawen554/shortener/internal/app"
	"github.com/rawen554/shortener/internal/config"
	ginLogger "github.com/rawen554/shortener/internal/logger"
	"github.com/rawen554/shortener/internal/store"
)

func setupRouter(a *app.App) *gin.Engine {
	r := gin.New()
	r.Use(ginLogger.Logger())
	r.Use(gzip.Gzip(gzip.BestCompression, gzip.WithDecompressFn(gzip.DefaultDecompressHandle)))

	r.GET("/:id", a.RedirectToOriginal)
	r.POST("/", a.ShortenURL)

	api := r.Group("/api")
	{
		api.POST("/shorten", a.ShortenURL)
	}

	return r
}

func main() {
	config, err := config.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}
	storage, err := store.NewStorage(config.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	app := app.NewApp(config, storage)

	r := setupRouter(app)
	log.Fatal(r.Run(config.FlagRunAddr))
}
