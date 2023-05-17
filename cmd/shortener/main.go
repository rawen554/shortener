package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/rawen554/shortener/internal/app"
	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/store"
)

func setupRouter(a *app.App) *gin.Engine {
	r := gin.Default()

	r.GET("/:id", a.RedirectToOriginal)
	r.POST("/", a.ShortenURL)

	return r
}

func main() {
	config, err := config.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}
	app := app.NewApp(config, store.NewStorage())

	r := setupRouter(app)
	log.Fatal(r.Run(config.FlagRunAddr))
}
