package storage

import (
	"shorty/internal/app/config"
	"shorty/internal/app/storage/dbstorage"
	"shorty/internal/app/storage/filestorage"
	"shorty/internal/app/storage/mapstorage"
)

type Storage interface {
	Put(key, value string) error
	Get(key string) (string, error)
	Ping() error
	Batch() error
}

func NewStorage(config config.Config) (Storage, error) {
	if config.DatabaseDSN != "" {
		s, err := dbstorage.CreateDBStorage(config.DatabaseDSN)
		return s, err
	} else if config.FileStoragePath != "" {
		s, err := filestorage.CreateFileStorage(config.FileStoragePath)
		return s, err
	} else {
		s, err := mapstorage.CreateMapStorage()
		return s, err
	}
}
