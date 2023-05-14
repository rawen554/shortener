package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/rawen554/shortener/internal/flags"
	"github.com/rawen554/shortener/internal/handlers"
)

func setupRouter(m handlers.SyncURLMap) *gin.Engine {
	r := gin.Default()

	r.GET("/:id", handlers.RedirectToOriginal(m))
	r.POST("/", handlers.ShortenURL(m))

	return r
}

func main() {
	if err := flags.ParseFlags(); err != nil {
		log.Fatal(err)
	}
	m := handlers.NewSyncURLMap(make(map[string][]byte))

	r := setupRouter(m)
	log.Fatal(r.Run(flags.Config.FlagRunAddr))
}
