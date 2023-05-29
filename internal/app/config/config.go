package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress   string
	BaseAddress     string
	FileStoragePath string
}

func GetConfig() Config {
	var cfg = Config{}

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "server address")
	flag.StringVar(&cfg.BaseAddress, "b", "http://localhost:8080", "base address")
	flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/short-url-db.json", "file storage path")
	flag.Parse()

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	}

	if envBaseAddress := os.Getenv("BASE_URL"); envBaseAddress != "" {
		cfg.BaseAddress = envBaseAddress
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	}

	return cfg
}
