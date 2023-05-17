package storage

import (
	"sync"
)

type Storage struct {
	mu    sync.Mutex
	links map[string]string
}

func (s *Storage) Put(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links[key] = value
}

func (s *Storage) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.links[key]
	return val, ok
}

func NewStorage() Storage {
	return Storage{links: map[string]string{}}
}
