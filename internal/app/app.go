package app

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/store"
	"github.com/rawen554/shortener/internal/utils"
)

type GenericStore struct {
	Get      func(id string) (string, error)
	GetBatch func(urls []store.BatchReq) ([]store.BatchRes, error)
	Put      func(id string, url string) error
	PutBatch func(urls []store.BatchReq) error
}

type (
	App struct {
		config *config.ServerConfig
		store  *GenericStore
	}

	ShortenReq struct {
		URL string `json:"url"`
	}

	ShortenRes struct {
		Result string `json:"result"`
	}
)

func NewApp(config *config.ServerConfig, storage *GenericStore) *App {
	return &App{
		config: config,
		store:  storage,
	}
}

func (a *App) RedirectToOriginal(c *gin.Context) {
	res := c.Writer
	id := c.Param("id")

	originalURL, err := a.store.Get(id)
	if err != nil {
		log.Printf("Error getting original URL: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if originalURL == "" {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (a *App) ShortenBatch(c *gin.Context) {
	req := c.Request
	res := c.Writer

	batch := make([]store.BatchReq, 0)
	if err := json.NewDecoder(req.Body).Decode(&batch); err != nil {
		log.Printf("Body cannot be decoded: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err := a.store.PutBatch(batch)
	if err != nil {
		log.Printf("Cant put batch: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, err := a.store.GetBatch(batch)
	if err != nil {
		log.Printf("Cant get batch: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	for idx, urlObj := range result {
		resultURL, err := url.JoinPath(a.config.RedirectBaseURL, urlObj.CorrelationID)
		if err != nil {
			log.Printf("URL cannot be joined: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		result[idx].ShortURL = resultURL
	}

	res.WriteHeader(http.StatusCreated)
	res.Header().Add("Content-Type", "application/json")
	json.NewEncoder(res).Encode(result)
}

func (a *App) ShortenURL(c *gin.Context) {
	req := c.Request
	res := c.Writer

	var originalURL string

	switch req.RequestURI {
	case "/api/shorten":
		var shorten ShortenReq
		if err := json.NewDecoder(req.Body).Decode(&shorten); err != nil {
			log.Printf("Body cannot be decoded: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		originalURL = shorten.URL
	case "/":
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Printf("Body cannot be read: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		originalURL = string(body)
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

	switch req.RequestURI {
	case "/api/shorten":
		respURL := ShortenRes{
			Result: resultURL,
		}
		resp, err := json.Marshal(respURL)
		if err != nil {
			log.Printf("URL cannot be encoded: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(resp)

	case "/":
		res.Header().Add("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		if _, err := res.Write([]byte(resultURL)); err != nil {
			log.Printf("Error writing body: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (a *App) DBHealthCheck(c *gin.Context) {
	db, err := sql.Open("pgx", a.config.DatabaseDSN)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()
	c.Writer.WriteHeader(http.StatusOK)
}
