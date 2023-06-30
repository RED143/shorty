package mapstorage

import (
	"context"
	"shorty/internal/app/models"
	"sync"
)

type storageItem struct {
	OriginalURL string
	UserID      string
}

type MapStorage struct {
	mu    *sync.Mutex
	links map[string]storageItem
}

func (s *MapStorage) Get(ctx context.Context, key string) (models.UserURLs, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val := s.links[key]
	return models.UserURLs{
		OriginalURL: val.OriginalURL,
		ShortURL:    key,
	}, nil
}

func (s *MapStorage) Put(ctx context.Context, key, value, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links[key] = storageItem{OriginalURL: value, UserID: userID}
	return nil
}

func (s *MapStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *MapStorage) Batch(ctx context.Context, urls []models.UserURLs, userID string) error {
	for _, url := range urls {
		if err := s.Put(ctx, url.ShortURL, url.OriginalURL, userID); err != nil {
			return err
		}
	}
	return nil
}

func (s *MapStorage) UserURLs(ctx context.Context, userID string) ([]models.UserURLs, error) {
	var UserUrls []models.UserURLs
	for shortURL, storageItem := range s.links {
		if storageItem.UserID == userID {
			UserUrls = append(UserUrls, models.UserURLs{OriginalURL: storageItem.OriginalURL, ShortURL: shortURL})
		}
	}
	return UserUrls, nil
}

func (s *MapStorage) DeleteUserURls(ctx context.Context, urls []string, userID string) error {
	return nil
}

func (s *MapStorage) Close() error {
	return nil
}

func CreateMapStorage() (*MapStorage, error) {
	s := &MapStorage{
		mu:    &sync.Mutex{},
		links: map[string]storageItem{},
	}

	return s, nil
}
