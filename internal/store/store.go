package store

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
	"sync"
)

type Storage struct {
	path      string
	mux       *sync.Mutex
	urls      map[string]string
	urlsCount int
	sr        *StorageReader
	sw        *StorageWriter
}

type Record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewStorage(filename string) (*Storage, error) {
	sr, err := NewStorageReader(filename)
	if err != nil {
		return nil, err
	}

	records, err := sr.ReadFromFile()
	if err != nil {
		return nil, err
	}

	sw, err := NewStorageWriter(filename)
	if err != nil {
		return nil, err
	}

	return &Storage{
		path:      filename,
		mux:       &sync.Mutex{},
		urls:      records,
		urlsCount: len(records),
		sr:        sr,
		sw:        sw,
	}, nil
}

func (s *Storage) DeleteStorageFile() error {
	return os.Remove(s.path)
}

type StorageReader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewStorageReader(filename string) (*StorageReader, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &StorageReader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (sr *StorageReader) ReadFromFile() (map[string]string, error) {
	records := make(map[string]string)
	for {
		r, err := sr.ReadLine()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		records[r.ShortURL] = r.OriginalURL
	}

	return records, nil
}

func (sr *StorageReader) ReadLine() (*Record, error) {
	r := Record{}
	if err := sr.decoder.Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}

type StorageWriter struct {
	file    *os.File
	encoder *json.Encoder
}

func NewStorageWriter(filename string) (*StorageWriter, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &StorageWriter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (sw *StorageWriter) AppendToFile(r *Record) error {
	return sw.encoder.Encode(&r)
}

func (s *Storage) Put(id string, url string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.urls[id] = url
	s.urlsCount += 1
	return s.sw.AppendToFile(&Record{UUID: strconv.Itoa(s.urlsCount), OriginalURL: url, ShortURL: id})
}

func (s *Storage) Get(id string) (string, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	originalURL := s.urls[id]
	return originalURL, nil
}
