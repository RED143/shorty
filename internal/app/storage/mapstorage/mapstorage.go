package mapstorage

import (
	"context"
	"shorty/internal/app/models"
	"sync"
)

type storageItem struct {
	OriginalURL string
	UserID      string
	IsDeleted   bool
}

type MapStorage struct {
	mu    *sync.Mutex
	Links map[string]storageItem
}

func (s *MapStorage) Get(ctx context.Context, key string) (models.UserURLs, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val := s.Links[key]
	return models.UserURLs{
		OriginalURL: val.OriginalURL,
		ShortURL:    key,
		IsDeleted:   val.IsDeleted,
	}, nil
}

func (s *MapStorage) Put(ctx context.Context, key, value, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Links[key] = storageItem{OriginalURL: value, UserID: userID, IsDeleted: false}
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
	var userUrls []models.UserURLs
	for shortURL, storageItem := range s.Links {
		if storageItem.UserID == userID {
			userUrls = append(userUrls, models.UserURLs{OriginalURL: storageItem.OriginalURL, ShortURL: shortURL})
		}
	}
	return userUrls, nil
}

func (s *MapStorage) DeleteUserURls(ctx context.Context, shortURLs []string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, shortURL := range shortURLs {
		item, ok := s.Links[shortURL]
		if ok && item.UserID == userID {
			s.Links[shortURL] = storageItem{OriginalURL: item.OriginalURL, UserID: userID, IsDeleted: true}
		}
	}
	return nil
}

func (s *MapStorage) Close() error {
	return nil
}

func CreateMapStorage() (*MapStorage, error) {
	s := &MapStorage{
		mu:    &sync.Mutex{},
		Links: map[string]storageItem{},
	}

	return s, nil
}
