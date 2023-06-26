package store

import (
	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/models"
	"github.com/rawen554/shortener/internal/store/fs"
	"github.com/rawen554/shortener/internal/store/memory"
	"github.com/rawen554/shortener/internal/store/postgres"
)

type Store interface {
	Get(id string) (string, error)
	GetAllByUserID(userID string) ([]models.URLRecord, error)
	Put(id string, shortURL string, userID string) (string, error)
	PutBatch(data []models.URLBatchReq, userID string) ([]models.URLBatchRes, error)
	Ping() error
}

func NewStore(conf *config.ServerConfig) (Store, error) {
	if conf.DatabaseDSN != "" {
		return postgres.NewPostgresStore(conf.DatabaseDSN)
	}
	if conf.FileStoragePath != "" {
		return fs.NewFileStorage(conf.FileStoragePath)
	}
	return memory.NewMemoryStorage(make(map[string]string))
}
