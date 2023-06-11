package filestorage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"shorty/internal/app/hash"
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

func (s *fileStorage) Put(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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

	s.links[key] = value

	return nil
}

func (s *fileStorage) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val := s.links[key]
	return val, nil
}

func (s *fileStorage) Ping() error {
	return errors.New("there is no ping method for file storage")
}

func (s *fileStorage) Batch(urls models.ShortenBatchRequest) error {
	for _, url := range urls {
		if err := s.Put(hash.Generate([]byte(url.OriginalURL)), url.OriginalURL); err != nil {
			return err
		}
	}
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
