// Модуль работает как единая точка входа для создания хранилища сервиса.

package store

import (
	"context"
	"fmt"

	_ "github.com/golang/mock/mockgen/model"
	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/models"
	"github.com/rawen554/shortener/internal/store/fs"
	"github.com/rawen554/shortener/internal/store/memory"
	"github.com/rawen554/shortener/internal/store/postgres"
)

// Store Интерфейс содержит все необходимые методы для работы сервиса.
type Store interface {
	Get(id string) (string, error)
	GetStats() (*models.Stats, error)
	GetAllByUserID(userID string) ([]models.URLRecord, error)
	DeleteMany(ids models.DeleteUserURLsReq, userID string) error
	Put(id string, shortURL string, userID string) (string, error)
	PutBatch(data []models.URLBatchReq, userID string) ([]models.URLBatchRes, error)
	Ping() error
	Close()
}

// NewStore Функция получения конкретной реализации интерфейса.
// Приоритет выбора: база данных, сохранение в файл, внутрення память.
func NewStore(ctx context.Context, conf *config.ServerConfig) (Store, error) {
	if conf.DatabaseDSN != "" {
		store, err := postgres.NewPostgresStore(ctx, conf.DatabaseDSN)
		if err != nil {
			return nil, fmt.Errorf("error creating postgres store: %w", err)
		}
		return store, nil
	}
	if conf.FileStoragePath != "" {
		store, err := fs.NewFileStorage(conf.FileStoragePath)
		if err != nil {
			return nil, fmt.Errorf("error creating file store: %w", err)
		}
		return store, nil
	}

	store, err := memory.NewMemoryStorage(make(map[string]models.URLRecordMemory))
	if err != nil {
		return nil, fmt.Errorf("error creating memory store: %w", err)
	}
	return store, nil
}
