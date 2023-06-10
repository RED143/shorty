package mapstorage

import (
	"errors"
	"sync"
)

type mapStorage struct {
	mu    *sync.Mutex
	links map[string]string
}

func (s *mapStorage) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val := s.links[key]
	return val, nil
}

func (s *mapStorage) Put(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links[key] = value
	return nil
}

func (s *mapStorage) Ping() error {
	return errors.New("there is not a ping method for map storage")
}

func CreateMapStorage() (*mapStorage, error) {
	s := &mapStorage{
		mu:    &sync.Mutex{},
		links: map[string]string{},
	}

	return s, nil
}
