package filestorage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"shorty/internal/app/models"
	"strconv"
	"sync"
)

type fileStorage struct {
	mu         *sync.Mutex
	filePath   string
	linesCount int
	links      map[string]string
}

type fileLine struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (s *fileStorage) Put(ctx context.Context, key, value, userID string) error {
	file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open the file \"%s\": %v", s.filePath, err)
	}
	defer file.Close()

	s.linesCount += 1
	line := fileLine{UUID: strconv.Itoa(s.linesCount), ShortURL: key, OriginalURL: value}
	data, err := json.Marshal(&line)
	if err != nil {
		return fmt.Errorf("failed to encode json: %v", err)
	}
	data = append(data, '\n')

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to save data to file: %v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.links[key] = value

	return nil
}

func (s *fileStorage) Get(ctx context.Context, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val := s.links[key]
	return val, nil
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
	return nil, nil
}

func (s *fileStorage) Close() error {
	return nil
}

func CreateFileStorage(filePath string) (*fileStorage, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file \"%s\": %v", filePath, err)
	}
	defer file.Close()

	storage := &fileStorage{filePath: filePath, linesCount: 0, mu: &sync.Mutex{},
		links: map[string]string{}}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := fileLine{}
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			return nil, fmt.Errorf("failed to decode json: %v", err)
		}
		storage.links[line.ShortURL] = line.OriginalURL
		storage.linesCount += 1
	}

	return storage, nil
}
