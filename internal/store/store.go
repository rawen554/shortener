// Модуль работает как единая точка входа для создания хранилища сервиса.

package store

import (
	"context"

	_ "github.com/golang/mock/mockgen/model"
	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/models"
	"github.com/rawen554/shortener/internal/store/fs"
	"github.com/rawen554/shortener/internal/store/memory"
	"github.com/rawen554/shortener/internal/store/postgres"
)

// Интерфейс содержит все необходимые методы для работы сервиса.
type Store interface {
	Get(id string) (string, error)
	GetAllByUserID(userID string) ([]models.URLRecord, error)
	DeleteMany(ids models.DeleteUserURLsReq, userID string) error
	Put(id string, shortURL string, userID string) (string, error)
	PutBatch(data []models.URLBatchReq, userID string) ([]models.URLBatchRes, error)
	Ping() error
	Close()
}

// Функция получения конкретной реализации интерфейса.
// Приоритет выбора: база данных, сохранение в файл, внутрення память.
func NewStore(ctx context.Context, conf *config.ServerConfig) (Store, error) {
	if conf.DatabaseDSN != "" {
		return postgres.NewPostgresStore(ctx, conf.DatabaseDSN)
	}
	if conf.FileStoragePath != "" {
		return fs.NewFileStorage(conf.FileStoragePath)
	}
	return memory.NewMemoryStorage(make(map[string]models.URLRecordMemory))
}
