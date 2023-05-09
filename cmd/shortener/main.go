package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/rawen554/shortener/internal/flags"
	"github.com/rawen554/shortener/internal/handlers"
)

func setupRouter(urls map[string][]byte) *gin.Engine {
	r := gin.Default()

	r.GET("/:id", handlers.RedirectToOriginal(urls))
	r.POST("/", handlers.ShortenURL(urls))

	return r
}

func main() {
	err := flags.ParseFlags()
	if err != nil {
		panic(err)
	}
	var urls = make(map[string][]byte)

	r := setupRouter(urls)
	log.Fatal(r.Run(flags.Config.FlagRunAddr))
}
