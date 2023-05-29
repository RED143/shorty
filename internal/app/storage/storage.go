package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type Storage interface {
	Put(string, string) error
	Get(string) (string, error)
}

type fileStorage struct {
	filePath string
}

type fileLine struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (s fileStorage) Put(key, value string) error {
	file, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	countLines := 0
	isAlreadySaved := false

	for scanner.Scan() {
		line := fileLine{}
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

	line := fileLine{UUID: strconv.Itoa(countLines + 1), ShortURL: key, OriginalURL: value}
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

func (s fileStorage) Get(key string) (string, error) {
	file, err := os.OpenFile(s.filePath, os.O_RDONLY, 0666)
	if err != nil {
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := fileLine{}
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			return "", err
		}
		if line.ShortURL == key {
			return line.OriginalURL, nil
		}
	}
	return "", nil
}

type mapStorage struct {
	mu    *sync.Mutex
	links map[string]string
}

func (s mapStorage) Put(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links[key] = value
	return nil
}

func (s mapStorage) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val := s.links[key]
	return val, nil
}

func NewStorage(fileStoragePath string) Storage {
	switch fileStoragePath {
	case "":
		return &mapStorage{
			links: map[string]string{},
			mu:    &sync.Mutex{},
		}
	default:
		if _, err := os.Stat(fileStoragePath); os.IsNotExist(err) && fileStoragePath != "" {
			os.MkdirAll(filepath.Dir(fileStoragePath), 0666)
			os.Create(fileStoragePath)
		}
		return fileStorage{
			filePath: fileStoragePath,
		}
	}
}
