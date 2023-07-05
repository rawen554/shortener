package store

import (
	"context"

	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/models"
	"github.com/rawen554/shortener/internal/store/fs"
	"github.com/rawen554/shortener/internal/store/memory"
	"github.com/rawen554/shortener/internal/store/postgres"
)

type Store interface {
	Get(id string) (string, error)
	GetAllByUserID(userID string) ([]models.URLRecord, error)
	DeleteMany(ids models.DeleteUserURLsReq, userID string) error
	Put(id string, shortURL string, userID string) (string, error)
	PutBatch(data []models.URLBatchReq, userID string) ([]models.URLBatchRes, error)
	Ping() error
	Close()
}

func NewStore(ctx context.Context, conf *config.ServerConfig) (Store, error) {
	if conf.DatabaseDSN != "" {
		return postgres.NewPostgresStore(ctx, conf.DatabaseDSN)
	}
	if conf.FileStoragePath != "" {
		return fs.NewFileStorage(conf.FileStoragePath)
	}
	return memory.NewMemoryStorage(make(map[string]models.URLRecordMemory))
}
