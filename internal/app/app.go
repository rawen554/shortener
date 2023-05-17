package app

import (
	"fmt"
	"io"
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
		fmt.Printf("body cannot be read: %s\n", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := utils.GenerateRandomString(8)
	if err != nil {
		fmt.Printf("random string generator error: %s\n", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, err := url.JoinPath(a.config.RedirectBaseURL, id)
	if err != nil {
		fmt.Printf("URL cannot be joined: %s\n", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	a.store.Put(id, body)

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(result))
}
