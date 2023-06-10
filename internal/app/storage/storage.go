package storage

import (
	"shorty/internal/app/config"
	"shorty/internal/app/storage/dbStorage"
	"shorty/internal/app/storage/fileStorage"
	"shorty/internal/app/storage/mapStorage"
)

type Storage interface {
	Put(key, value string) error
	Get(key string) (string, error)
	Ping() error
}

func NewStorage(config config.Config) (Storage, error) {
	if config.DatabaseDSN != "" {
		s, err := dbStorage.CreateDBStorage(config.DatabaseDSN)
		return s, err
	} else if config.FileStoragePath != "" {
		s, err := fileStorage.CreateFileStorage(config.FileStoragePath)
		return s, err
	} else {
		s, err := mapStorage.CreateMapStorage()
		return s, err
	}
}
