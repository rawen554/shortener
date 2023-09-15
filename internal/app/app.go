package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/middleware/auth"
	"github.com/rawen554/shortener/internal/models"
	"github.com/rawen554/shortener/internal/store/postgres"
	"go.uber.org/zap"
)

const (
	slugLength      = 4
	applicationJSON = "application/json"
	textPlain       = "text/plain"
	contentType     = "Content-Type"
	location        = "Location"

	rootPath       = "/"
	pingPath       = "/ping"
	apiShortenPath = "/api/shorten"

	ErrorJoinURL     = "URL cannot be joined: %v"
	ErrorDecodeBody  = "Body cannot be decoded: %v"
	ErrorEncodeBody  = "Error writing response in JSON: %v"
	ErrorWritingBody = "Error writing body: %v"
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
	logger *zap.SugaredLogger
}

func NewApp(config *config.ServerConfig, store Store, logger *zap.SugaredLogger) *App {
	return &App{
		config: config,
		store:  store,
		logger: logger,
	}
}

func (a *App) DeleteUserRecords(c *gin.Context) {
	req := c.Request
	res := c.Writer
	userID := c.GetString(auth.UserIDKey)

	batch := make(models.DeleteUserURLsReq, 0)
	if err := json.NewDecoder(req.Body).Decode(&batch); err != nil {
		a.logger.Errorf(ErrorDecodeBody, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	go func() {
		err := a.store.DeleteMany(batch, userID)
		if err != nil {
			a.logger.Errorf("error deleting: %v", err)
		}
	}()

	res.WriteHeader(http.StatusAccepted)
}

func (a *App) GetUserRecords(c *gin.Context) {
	res := c.Writer
	userID := c.GetString(auth.UserIDKey)

	records, err := a.store.GetAllByUserID(userID)
	if err != nil {
		a.logger.Errorf("Error getting all user urls: %v", err)
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
			a.logger.Errorf(ErrorJoinURL, err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		records[idx].ShortURL = resultURL
	}

	c.JSON(http.StatusOK, records)
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
			a.logger.Errorf("Error getting original URL: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if originalURL == "" {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, originalURL)
}

func (a *App) ShortenBatch(c *gin.Context) {
	req := c.Request
	res := c.Writer
	userID := c.GetString(auth.UserIDKey)

	batch := make([]models.URLBatchReq, 0)
	if err := json.NewDecoder(req.Body).Decode(&batch); err != nil {
		a.logger.Errorf(ErrorDecodeBody, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	result, err := a.store.PutBatch(batch, userID)
	if err != nil {
		a.logger.Errorf("Cant put batch: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	for idx, urlObj := range result {
		resultURL, err := url.JoinPath(a.config.RedirectBaseURL, urlObj.CorrelationID)
		if err != nil {
			a.logger.Errorf(ErrorJoinURL, err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		result[idx].ShortURL = resultURL
	}

	res.WriteHeader(http.StatusCreated)
	res.Header().Add(contentType, applicationJSON)
	if err := json.NewEncoder(res).Encode(result); err != nil {
		a.logger.Errorf(ErrorEncodeBody, err)
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
	case apiShortenPath:
		var shorten models.ShortenReq
		if err := json.NewDecoder(req.Body).Decode(&shorten); err != nil {
			a.logger.Errorf(ErrorDecodeBody, err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		originalURL = shorten.URL
	case rootPath:
		body, err := io.ReadAll(req.Body)
		if err != nil {
			a.logger.Errorf("Body cannot be read: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		originalURL = string(body)
	}

	b := make([]byte, slugLength)
	_, err := rand.Read(b)
	if err != nil {
		a.logger.Errorf("Random string generator error: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	id := hex.EncodeToString(b)

	id, err = a.store.Put(id, originalURL, userID)
	if err != nil {
		if errors.Is(err, postgres.ErrDBInsertConflict) {
			res.WriteHeader(http.StatusConflict)
		} else {
			a.logger.Errorf("Error saving data: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		res.WriteHeader(http.StatusCreated)
	}

	resultURL, err := url.JoinPath(a.config.RedirectBaseURL, id)
	if err != nil {
		a.logger.Errorf(ErrorJoinURL, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch req.RequestURI {
	case apiShortenPath:
		respURL := models.ShortenRes{
			Result: resultURL,
		}
		resp, err := json.Marshal(respURL)
		if err != nil {
			a.logger.Errorf("URL cannot be encoded: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.Header().Set(contentType, applicationJSON)
		if _, err := res.Write(resp); err != nil {
			a.logger.Errorf(ErrorWritingBody, err)
		}

	case rootPath:
		res.Header().Set(contentType, textPlain)
		if _, err := res.Write([]byte(resultURL)); err != nil {
			a.logger.Errorf("Error writing body: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (a *App) Ping(c *gin.Context) {
	if err := a.store.Ping(); err != nil {
		a.logger.Errorf("Error opening connection to DB: %v", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.Writer.WriteHeader(http.StatusOK)
}
