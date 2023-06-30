package filestorage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"os"
	"shorty/internal/app/models"
	"shorty/internal/app/storage/mapstorage"
)

type fileStorage struct {
	mapStorage *mapstorage.MapStorage
	filePath   string
}

type fileLine struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	IsDeleted   bool   `json:"is_deleted"`
}

func (s *fileStorage) Put(ctx context.Context, shortURL, originalURL, userID string) error {
	file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open the file \"%s\": %w", s.filePath, err)
	}
	defer file.Close()

	line := fileLine{UUID: uuid.NewString(), ShortURL: shortURL, OriginalURL: originalURL, UserID: userID}
	data, err := json.Marshal(&line)
	if err != nil {
		return fmt.Errorf("failed to encode json: %w", err)
	}
	data = append(data, '\n')

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to save data to file: %w", err)
	}

	s.mapStorage.Put(ctx, shortURL, originalURL, userID)

	return nil
}

func (s *fileStorage) Get(ctx context.Context, shortURL string) (models.UserURLs, error) {
	return s.mapStorage.Get(ctx, shortURL)
}

func (s *fileStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *fileStorage) Batch(ctx context.Context, urls []models.UserURLs, userID string) error {
	for _, url := range urls {
		if err := s.Put(ctx, url.ShortURL, url.OriginalURL, userID); err != nil {
			return err
		}
	}
	return nil
}

func (s *fileStorage) UserURLs(ctx context.Context, userID string) ([]models.UserURLs, error) {
	return s.mapStorage.UserURLs(ctx, userID)
}

func (s *fileStorage) DeleteUserURls(ctx context.Context, shortURLs []string, userID string) error {
	return s.mapStorage.DeleteUserURls(ctx, shortURLs, userID)
}

func (s *fileStorage) Close() error {
	return nil
}

func CreateFileStorage(filePath string, mapStorage *mapstorage.MapStorage) (*fileStorage, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file \"%s\": %w", filePath, err)
	}
	defer file.Close()

	s := &fileStorage{filePath: filePath, mapStorage: mapStorage}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := fileLine{}
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			return nil, fmt.Errorf("failed to decode json: %w", err)
		}
		err := s.mapStorage.Put(context.Background(), line.ShortURL, line.OriginalURL, line.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to save line in map storafe: %w", err)
		}
	}

	return s, nil
}
