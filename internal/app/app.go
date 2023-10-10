package app

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/logic"
	"github.com/rawen554/shortener/internal/middleware/auth"
	"github.com/rawen554/shortener/internal/models"
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
	GetStats() (*models.Stats, error)
	GetAllByUserID(userID string) ([]models.URLRecord, error)
	DeleteMany(ids models.DeleteUserURLsReq, userID string) error
	Put(id string, shortURL string, userID string) (string, error)
	PutBatch(data []models.URLBatchReq, userID string) ([]models.URLBatchRes, error)
	Ping() error
}

type App struct {
	config    *config.ServerConfig
	logger    *zap.SugaredLogger
	coreLogic *logic.CoreLogic
}

func NewApp(config *config.ServerConfig, coreLogic *logic.CoreLogic, logger *zap.SugaredLogger) *App {
	return &App{
		config:    config,
		coreLogic: coreLogic,
		logger:    logger,
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
		if err := a.coreLogic.DeleteUserRecords(c, userID, batch); err != nil {
			a.logger.Errorf("error deleting: %v", err)
		}
	}()

	res.WriteHeader(http.StatusAccepted)
}

func (a *App) GetUserRecords(c *gin.Context) {
	res := c.Writer
	userID := c.GetString(auth.UserIDKey)

	records, err := a.coreLogic.GetUserRecords(c, userID)
	if err != nil {
		if errors.Is(err, logic.ErrNoContent) {
			res.WriteHeader(http.StatusNoContent)
			return
		}

		a.logger.Errorf("Error getting all user urls: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, records)
}

func (a *App) RedirectToOriginal(c *gin.Context) {
	res := c.Writer
	id := c.Param("id")

	originalURL, err := a.coreLogic.GetOriginalURL(c, id)
	if err != nil {
		if errors.Is(err, logic.ErrIsDeleted) {
			res.WriteHeader(http.StatusGone)
			return
		}

		if errors.Is(err, logic.ErrNotFound) {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		a.logger.Errorf("Error getting original URL: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
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

	result, err := a.coreLogic.ShortenBatch(c, userID, batch)
	if err != nil {
		a.logger.Errorf("Cant put batch: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, result)
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

	resultURL, err := a.coreLogic.ShortenURL(c, userID, originalURL)
	if err != nil {
		if errors.Is(err, logic.ErrConflict) {
			res.WriteHeader(http.StatusConflict)
			return
		}

		a.logger.Errorf("Error saving data: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusCreated)

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
	if err := a.coreLogic.Ping(c); err != nil {
		a.logger.Errorf("Error opening connection to DB: %v", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	c.Writer.WriteHeader(http.StatusOK)
}

func (a *App) GetStats(c *gin.Context) {
	stats, err := a.coreLogic.GetStats(c)
	if err != nil {
		a.logger.Errorf("error getting service stats: %v", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, stats)
}
