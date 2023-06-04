package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Storage struct {
	mu       *sync.Mutex
	links    map[string]string
	filePath string
}

type fileLine struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (s *Storage) Put(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links[key] = value

	if s.filePath != "" {
		if err := s.saveURLToFile(key, value); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val := s.links[key]
	return val, nil
}

func (s *Storage) mapURLsFromFileToStorage() error {
	file, err := os.OpenFile(s.filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("failed to open the file \"%s\": %w", s.filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := fileLine{}
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			return fmt.Errorf("failed to decode json: %w", err)
		}
		s.links[line.ShortURL] = line.OriginalURL
	}

	return nil
}

func (s *Storage) saveURLToFile(key, value string) error {
	file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open the file \"%s\": %w", s.filePath, err)
	}
	defer file.Close()

	line := fileLine{ShortURL: key, OriginalURL: value}
	data, err := json.Marshal(&line)
	if err != nil {
		return fmt.Errorf("failed to encode json: %w", err)
	}
	data = append(data, '\n')

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to save data to file: %w", err)
	}

	return nil
}

func NewStorage(filePath string) (*Storage, error) {
	storage := Storage{
		mu:       &sync.Mutex{},
		links:    map[string]string{},
		filePath: filePath,
	}

	if filePath != "" {
		if err := storage.mapURLsFromFileToStorage(); err != nil {
			return nil, err
		}
	}
	return &storage, nil
}
