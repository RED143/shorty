package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type Storage struct {
	mu       *sync.Mutex
	links    map[string]string
	filePath string
}

type Line struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (s *Storage) Put(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.filePath == "" {
		s.links[key] = value
		return nil
	}
	file, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	countLines := 0
	isAlreadySaved := false

	for scanner.Scan() {
		line := Line{}
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			return err
		}
		if line.ShortURL == key {
			isAlreadySaved = true
		}
		countLines++
	}

	if isAlreadySaved {
		return nil
	}

	line := Line{UUID: strconv.Itoa(countLines + 1), ShortURL: key, OriginalURL: value}
	data, err := json.Marshal(&line)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.filePath == "" {
		val := s.links[key]
		return val, nil
	}
	file, err := os.OpenFile(s.filePath, os.O_RDONLY, 0666)
	if err != nil {
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := Line{}
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			return "", err
		}
		if line.ShortURL == key {
			return line.OriginalURL, nil
		}
	}
	return "", nil
}

func NewStorage(fileStoragePath string) *Storage {
	if _, err := os.Stat(fileStoragePath); os.IsNotExist(err) && fileStoragePath != "" {
		os.MkdirAll(filepath.Dir(fileStoragePath), 0666)
		os.Create(fileStoragePath)
	}
	return &Storage{
		filePath: fileStoragePath,
		links:    map[string]string{},
		mu:       &sync.Mutex{},
	}
}
