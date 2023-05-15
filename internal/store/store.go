package store

import "sync"

type Storage struct {
	mux  *sync.Mutex
	urls map[string][]byte
}

func NewStorage(urls map[string][]byte) *Storage {
	return &Storage{
		mux:  &sync.Mutex{},
		urls: urls,
	}
}

func (s *Storage) Put(id string, url []byte) {
	s.mux.Lock()
	s.urls[id] = url
	s.mux.Unlock()
}

func (s *Storage) Get(id string) []byte {
	s.mux.Lock()
	originalURL := s.urls[id]
	s.mux.Unlock()
	return originalURL
}
