package storage

import (
	"shorty/internal/app/storage/fileStorage"
	"shorty/internal/app/storage/mapStorage"
)

type Storage interface {
	Put(key, value string) error
	Get(key string) (string, error)
}

func NewStorage(filePath string) (Storage, error) {
	if filePath != "" {
		s, err := fileStorage.CreateFileStorage(filePath)
		return s, err
	} else {
		s, err := mapStorage.CreateMapStorage()
		return s, err
	}
}
