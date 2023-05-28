package app

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/store"
	"github.com/rawen554/shortener/internal/utils"
)

type (
	App struct {
		config *config.ServerConfig
		store  *store.Storage
	}

	ShortenReq struct {
		URL string `json:"url"`
	}

	ShortenRes struct {
		Result string `json:"result"`
	}
)

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

	if originalURL == "" {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (a *App) ShortenURL(c *gin.Context) {
	req := c.Request
	res := c.Writer

	var originalURL string

	contentType := req.Header.Get("Content-Type")
	woCharset := strings.Split(contentType, ";")
	switch woCharset[0] {
	case "application/json":
		var shorten ShortenReq
		if err := json.NewDecoder(req.Body).Decode(&shorten); err != nil {
			log.Printf("Body cannot be decoded: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		originalURL = shorten.URL
	case "text/plain":
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Printf("Body cannot be read: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		originalURL = string(body)
	default:
		log.Printf("Invalid Content-Type: %v", req.Header.Get("Content-Type"))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := utils.GenerateRandomString(8)
	if err != nil {
		log.Printf("Random string generator error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	resultURL, err := url.JoinPath(a.config.RedirectBaseURL, id)
	if err != nil {
		log.Printf("URL cannot be joined: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	a.store.Put(id, originalURL)

	switch woCharset[0] {
	case "application/json":
		respURL := ShortenRes{
			Result: resultURL,
		}
		resp, err := json.Marshal(respURL)
		if err != nil {
			log.Printf("URL cannot be encoded: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(resp)

	case "text/plain":
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		if _, err := res.Write([]byte(resultURL)); err != nil {
			log.Printf("Error writing body: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
