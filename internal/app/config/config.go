package config

import (
	"flag"
	"os"
)

type config struct {
	ServerAddress string
	BaseAddress   string
}

var cfg = config{}

func ParseFlags() {
	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "set a server address")
	flag.StringVar(&cfg.BaseAddress, "b", "http://localhost:8080", "set a base address")
	flag.Parse()
}

func GetConfig() config {
	envServerAddress := os.Getenv("SERVER_ADDRESS")
	if envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	}

	envBaseAddress := os.Getenv("BASE_URL")
	if envBaseAddress != "" {
		cfg.BaseAddress = envBaseAddress
	}

	return cfg
}
