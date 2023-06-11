package mapstorage

import (
	"errors"
	"fmt"
	"shorty/internal/app/models"
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

func (s *mapStorage) Batch(urls models.ShortenBatchRequest) error {
	fmt.Println("map storage batching", urls)
	return nil
}

func CreateMapStorage() (*mapStorage, error) {
	s := &mapStorage{
		mu:    &sync.Mutex{},
		links: map[string]string{},
	}

	return s, nil
}
