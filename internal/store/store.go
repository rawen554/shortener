package store

import "sync"

type Storage struct {
	mux  *sync.Mutex
	urls map[string][]byte
}

func NewStorage() *Storage {
	return &Storage{
		mux:  &sync.Mutex{},
		urls: make(map[string][]byte),
	}
}

func (s *Storage) Put(id string, url []byte) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.urls[id] = url
}

func (s *Storage) Get(id string) []byte {
	s.mux.Lock()
	defer s.mux.Unlock()
	originalURL := s.urls[id]
	return originalURL
}
