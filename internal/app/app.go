package app

import (
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/store"
	"github.com/rawen554/shortener/internal/utils"
)

type App struct {
	config *config.ServerConfig
	store  *store.Storage
}

func NewApp(config *config.ServerConfig, storage *store.Storage) *App {
	return &App{
		config: config,
		store:  storage,
	}
}

func (a *App) RedirectToOriginal(c *gin.Context) {
	res := c.Writer
	id := c.Param("id")

	originalURL := a.store.Get(id)

	if originalURL == nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	res.Header().Set("Location", string(originalURL))
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (a *App) ShortenURL(c *gin.Context) {
	req := c.Request
	res := c.Writer
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Body cannot be read: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := utils.GenerateRandomString(8)
	if err != nil {
		log.Printf("Random string generator error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, err := url.JoinPath(a.config.RedirectBaseURL, id)
	if err != nil {
		log.Printf("URL cannot be joined: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	a.store.Put(id, body)

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	if _, err := res.Write([]byte(result)); err != nil {
		log.Printf("Error writing body: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}
