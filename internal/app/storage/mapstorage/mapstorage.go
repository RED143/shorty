package mapstorage

import (
	"context"
	"shorty/internal/app/hash"
	"shorty/internal/app/models"
	"sync"
)

type mapStorage struct {
	mu    *sync.Mutex
	links map[string]string
}

func (s *mapStorage) Get(ctx context.Context, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val := s.links[key]
	return val, nil
}

func (s *mapStorage) Put(ctx context.Context, key, value string, userId int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links[key] = value
	return nil
}

func (s *mapStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *mapStorage) Batch(ctx context.Context, urls models.ShortenBatchRequest) error {
	for _, url := range urls {
		if err := s.Put(ctx, hash.Generate([]byte(url.OriginalURL)), url.OriginalURL, 0); err != nil {
			return err
		}
	}
	return nil
}

func (s *mapStorage) Close() error {
	return nil
}

func CreateMapStorage() (*mapStorage, error) {
	s := &mapStorage{
		mu:    &sync.Mutex{},
		links: map[string]string{},
	}

	return s, nil
}
