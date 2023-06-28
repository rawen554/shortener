package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rawen554/shortener/internal/auth"
	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/models"
	"github.com/rawen554/shortener/internal/store/postgres"
)

var ErrDecodeBody = errors.New("body cannot be decoded")
var ErrJoinURL = errors.New("URL cannot be joined")
var ErrWriteBody = errors.New("error writing body")

const (
	contentType     = "Content-Type"
	applicationJSON = "application/json"
	apiShorten      = "/api/shorten"
	root            = "/"
	randBytesSize   = 4
)

type Store interface {
	Get(id string) (string, error)
	GetAllByUserID(userID string) ([]models.URLRecord, error)
	DeleteMany(ids models.DeleteUserURLsReq, userID string) error
	Put(id string, shortURL string, userID string) (string, error)
	PutBatch(data []models.URLBatchReq, userID string) ([]models.URLBatchRes, error)
	Ping() error
}

type App struct {
	config *config.ServerConfig
	store  Store
}

func NewApp(config *config.ServerConfig, store Store) *App {
	return &App{
		config: config,
		store:  store,
	}
}

func (a *App) DeleteUserRecords(c *gin.Context) {
	req := c.Request
	res := c.Writer
	userID := c.GetString(auth.UserIDKey)

	batch := make(models.DeleteUserURLsReq, 0)
	if err := json.NewDecoder(req.Body).Decode(&batch); err != nil {
		fmt.Errorf("body cannot be decoded: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	go func() {
		err := a.store.DeleteMany(batch, userID)
		if err != nil {
			log.Printf("error deleting: %v", err)
		}
	}()

	res.WriteHeader(http.StatusAccepted)
}

func (a *App) GetUserRecors(c *gin.Context) {
	res := c.Writer
	userID := c.GetString(auth.UserIDKey)

	records, err := a.store.GetAllByUserID(userID)
	if err != nil {
		log.Printf("Error getting all user urls: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(records) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	for idx, urlObj := range records {
		resultURL, err := url.JoinPath(a.config.RedirectBaseURL, urlObj.ShortURL)
		if err != nil {
			log.Printf("%v: %v", ErrJoinURL, err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		records[idx].ShortURL = resultURL
	}

	res.Header().Add(contentType, applicationJSON)
	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(records); err != nil {
		log.Printf("Error writing response in JSON: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (a *App) RedirectToOriginal(c *gin.Context) {
	res := c.Writer
	id := c.Param("id")

	originalURL, err := a.store.Get(id)
	if err != nil {
		if errors.Is(err, postgres.ErrURLDeleted) {
			res.WriteHeader(http.StatusGone)
			return
		} else {
			log.Printf("Error getting original URL: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
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
	userID := c.GetString(auth.UserIDKey)

	batch := make([]models.URLBatchReq, 0)
	if err := json.NewDecoder(req.Body).Decode(&batch); err != nil {
		log.Printf("Body cannot be decoded: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, err := a.store.PutBatch(batch, userID)
	if err != nil {
		log.Printf("Cant put batch: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	for idx, urlObj := range result {
		resultURL, err := url.JoinPath(a.config.RedirectBaseURL, urlObj.CorrelationID)
		if err != nil {
			log.Printf("%v: %v", ErrJoinURL, err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		result[idx].ShortURL = resultURL
	}

	res.WriteHeader(http.StatusCreated)
	res.Header().Add(contentType, applicationJSON)
	if err := json.NewEncoder(res).Encode(result); err != nil {
		log.Printf("Error writing response in JSON: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (a *App) ShortenURL(c *gin.Context) {
	req := c.Request
	res := c.Writer
	userID := c.GetString(auth.UserIDKey)

	var originalURL string

	switch req.RequestURI {
	case apiShorten:
		var shorten models.ShortenReq
		if err := json.NewDecoder(req.Body).Decode(&shorten); err != nil {
			log.Printf("Body cannot be decoded: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		originalURL = shorten.URL
	case root:
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Printf("Body cannot be read: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		originalURL = string(body)
	}

	b := make([]byte, randBytesSize)
	_, err := rand.Read(b)
	if err != nil {
		log.Printf("Random string generator error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	id := hex.EncodeToString(b)

	id, err = a.store.Put(id, originalURL, userID)
	if err != nil {
		if errors.Is(err, postgres.ErrDBInsertConflict) {
			res.WriteHeader(http.StatusConflict)
		} else {
			log.Printf("Error saving data: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		res.WriteHeader(http.StatusCreated)
	}

	resultURL, err := url.JoinPath(a.config.RedirectBaseURL, id)
	if err != nil {
		log.Printf("%v: %v", ErrJoinURL, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch req.RequestURI {
	case apiShorten:
		respURL := models.ShortenRes{
			Result: resultURL,
		}
		resp, err := json.Marshal(respURL)
		if err != nil {
			log.Printf("URL cannot be encoded: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.Header().Set(contentType, applicationJSON)
		if _, err := res.Write(resp); err != nil {
			log.Printf("%v: %v", ErrWriteBody, err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

	case root:
		res.Header().Set(contentType, "text/plain")
		if _, err := res.Write([]byte(resultURL)); err != nil {
			log.Printf("Error writing body: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (a *App) Ping(c *gin.Context) {
	if err := a.store.Ping(); err != nil {
		log.Printf("Error opening connection to DB: %v", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.Writer.WriteHeader(http.StatusOK)
}
