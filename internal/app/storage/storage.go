package storage

import (
	"sync"
)

type storage struct {
	mu    sync.Mutex
	links map[string]string
}

var Storage = storage{
	links: map[string]string{},
}

func (s *storage) Put(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links[key] = value
}

func (s *storage) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.links[key]
	return val, ok
}
