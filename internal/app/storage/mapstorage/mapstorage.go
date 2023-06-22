package mapstorage

import (
	"context"
	"shorty/internal/app/models"
	"sync"
)

type mapStorage struct {
	mu    *sync.Mutex
	links map[string]string
}

func (s *mapStorage) Get(ctx context.Context, key string) (models.UserURLs, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val := s.links[key]
	return models.UserURLs{
		OriginalURL: val,
		ShortURL:    key,
	}, nil
}

func (s *mapStorage) Put(ctx context.Context, key, value, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links[key] = value
	return nil
}

func (s *mapStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *mapStorage) Batch(ctx context.Context, urls []models.UserURLs, userID string) error {
	for _, url := range urls {
		if err := s.Put(ctx, url.ShortURL, url.OriginalURL, userID); err != nil {
			return err
		}
	}
	return nil
}

func (s *mapStorage) UserURLs(ctx context.Context, userID string) ([]models.UserURLs, error) {
	return nil, nil
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
