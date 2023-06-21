package memory

import (
	"sync"

	"github.com/rawen554/shortener/internal/models"
)

type MemoryStorage struct {
	mux       *sync.Mutex
	urls      map[string]string
	UrlsCount int
}

func NewMemoryStorage(records map[string]string) (*MemoryStorage, error) {
	return &MemoryStorage{
		mux:       &sync.Mutex{},
		urls:      records,
		UrlsCount: len(records),
	}, nil
}

func (s *MemoryStorage) Put(id string, url string) (string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.urls[id] = url
	s.UrlsCount += 1
	return id, nil
}

func (s *MemoryStorage) Get(id string) (string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	originalURL := s.urls[id]
	return originalURL, nil
}

func (s *MemoryStorage) PutBatch(urls []models.URLBatchReq) ([]models.URLBatchRes, error) {
	result := make([]models.URLBatchRes, 0)

	for _, url := range urls {
		id, err := s.Put(url.CorrelationID, url.OriginalURL)
		if err != nil {
			return nil, err
		}
		result = append(result, models.URLBatchRes{
			CorrelationID: id,
			ShortURL:      id,
		})
	}

	return result, nil
}

func (s *MemoryStorage) Ping() error {
	return nil
}
