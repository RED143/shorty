package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress        string
	BaseAddress          string
	FileStoragePath      string
	DatabaseDSN          string
	JWTSecret            string
	MaxDBConnections     int
	MaxIdleDBConnections int
}

func GetConfig() Config {
	var cfg = Config{MaxDBConnections: 100, MaxIdleDBConnections: 100}

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "server address")
	flag.StringVar(&cfg.BaseAddress, "b", "http://localhost:8080", "base address")
	flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/short-url-db.json", "file storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database DSN")
	flag.StringVar(&cfg.JWTSecret, "s", "jwt_secret", "JWT secret")
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

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		cfg.DatabaseDSN = envDatabaseDSN
	}

	if envJWTSecret := os.Getenv("JWT_SECRET"); envJWTSecret != "" {
		cfg.JWTSecret = envJWTSecret
	}

	return cfg
}
