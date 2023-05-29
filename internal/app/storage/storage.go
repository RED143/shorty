package storage

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"sync"
)

type Storage struct {
	mu       *sync.Mutex
	filePath string
}

type Line struct {
	UUID        string `json:"uuid"`
	ShortUrl    string `json:"short_url"`
	OriginalUrl string `json:"original_url"`
}

func (s *Storage) Put(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("failed to open storage file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	countLines := 0
	isAlreadySaved := false

	for scanner.Scan() {
		line := Line{}
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			log.Fatalf("failed to read json line: %v", err)
		}
		if line.ShortUrl == key {
			isAlreadySaved = true
		}
		countLines++
	}

	if isAlreadySaved {
		return
	}

	line := Line{UUID: strconv.Itoa(countLines + 1), ShortUrl: key, OriginalUrl: value}
	data, err := json.Marshal(&line)
	if err != nil {
		log.Fatalf("failed format data to json format: %v", err)
	}
	data = append(data, '\n')

	_, err = file.Write(data)
	if err != nil {
		log.Fatalf("failed to write in file: %v", err)
	}
}

func (s *Storage) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := os.OpenFile(s.filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("failed to open storage file: %v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := Line{}
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			log.Fatalf("failed to read json line: %v", err)
		}
		if line.ShortUrl == key {
			return line.OriginalUrl, true
			break
		}
	}
	return "", false
}

func NewStorage(fileStoragePath string) *Storage {
	return &Storage{
		filePath: fileStoragePath,
		mu:       &sync.Mutex{},
	}
}
