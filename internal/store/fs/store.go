package fs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/rawen554/shortener/internal/models"
	"github.com/rawen554/shortener/internal/store/memory"
)

const FileStorageFilePerm = 0600

type FSStorage struct {
	*memory.MemoryStorage
	sr   *StorageReader
	sw   *StorageWriter
	path string
}

func NewFileStorage(filename string) (*FSStorage, error) {
	sr, err := NewStorageReader(filename)
	if err != nil {
		return nil, err
	}

	records, err := sr.ReadFromFile()
	if err != nil {
		return nil, err
	}

	storage, err := memory.NewMemoryStorage(records)
	if err != nil {
		return nil, fmt.Errorf("error initialising memory storage with records: %w", err)
	}

	sw, err := NewStorageWriter(filename)
	if err != nil {
		return nil, err
	}

	return &FSStorage{
		path:          filename,
		MemoryStorage: storage,
		sr:            sr,
		sw:            sw,
	}, nil
}

func (s *FSStorage) PutBatch(urls []models.URLBatchReq, userID string) ([]models.URLBatchRes, error) {
	result := make([]models.URLBatchRes, 0)

	for _, url := range urls {
		id, err := s.Put(url.CorrelationID, url.OriginalURL, userID)
		if err != nil {
			return nil, err
		}
		result = append(result, models.URLBatchRes{
			CorrelationID: url.CorrelationID,
			ShortURL:      id,
		})
	}

	return result, nil
}

func (s *FSStorage) Ping() error {
	return nil
}

func (s *FSStorage) Close() {
	if err := s.sw.file.Close(); err != nil {
		log.Printf("error closing file: %v", err)
	}
}

func (s *FSStorage) DeleteStorageFile() error {
	if err := os.Remove(s.path); err != nil {
		return fmt.Errorf("error delete file: %w", err)
	}
	return nil
}

type StorageReader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewStorageReader(filename string) (*StorageReader, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, FileStorageFilePerm)
	if err != nil {
		return nil, fmt.Errorf("error open file: %w", err)
	}

	return &StorageReader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (sr *StorageReader) ReadFromFile() (map[string]models.URLRecordMemory, error) {
	records := make(map[string]models.URLRecordMemory)
	for {
		r, err := sr.ReadLine()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		records[r.ShortURL] = models.URLRecordMemory{OriginalURL: r.OriginalURL, UserID: r.UserID}
	}

	return records, nil
}

func (sr *StorageReader) ReadLine() (*models.URLRecordFS, error) {
	r := models.URLRecordFS{}
	if err := sr.decoder.Decode(&r); err != nil {
		return nil, fmt.Errorf("error decode records: %w", err)
	}

	return &r, nil
}

type StorageWriter struct {
	file    *os.File
	encoder *json.Encoder
}

func NewStorageWriter(filename string) (*StorageWriter, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, FileStorageFilePerm)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	return &StorageWriter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (sw *StorageWriter) AppendToFile(r *models.URLRecordFS) error {
	if err := sw.encoder.Encode(&r); err != nil {
		return fmt.Errorf("error encode records: %w", err)
	}
	return nil
}

func (s *FSStorage) Put(id string, url string, userID string) (string, error) {
	id, err := s.MemoryStorage.Put(id, url, userID)
	if err != nil {
		return "", fmt.Errorf("error put file: %w", err)
	}
	return id,
		s.sw.AppendToFile(&models.URLRecordFS{UUID: strconv.Itoa(s.UrlsCount), UserID: userID, URLRecord: models.URLRecord{
			OriginalURL: url, ShortURL: id,
		}})
}

func (s *FSStorage) GetStats() (stats *models.Stats, err error) {
	stats, err = s.MemoryStorage.GetStats()
	if err != nil {
		return nil, fmt.Errorf("fs storage error: %w", err)
	}

	return stats, nil
}
