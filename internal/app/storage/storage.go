package storage

import (
	"context"
	"shorty/internal/app/config"
	"shorty/internal/app/models"
	"shorty/internal/app/storage/dbstorage"
	"shorty/internal/app/storage/filestorage"
	"shorty/internal/app/storage/mapstorage"
)

type Storage interface {
	Put(ctx context.Context, key, value, userID string) error
	Get(ctx context.Context, key string) (string, error)
	Ping(ctx context.Context) error
	Batch(ctx context.Context, urls []models.UserURLs, userID string) error
	UserURLs(ctx context.Context, userID string) ([]models.UserURLs, error)
	Close() error
}

func NewStorage(config config.Config) (Storage, error) {
	if config.DatabaseDSN != "" {
		s, err := dbstorage.CreateDBStorage(context.Background(), config.DatabaseDSN)
		return s, err
	} else if config.FileStoragePath != "" {
		s, err := filestorage.CreateFileStorage(config.FileStoragePath)
		return s, err
	} else {
		s, err := mapstorage.CreateMapStorage()
		return s, err
	}
}
