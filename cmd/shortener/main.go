package main

import (
	"log"
	"shorty/internal/app/server"
)

func main() {
	if err := server.Start(); err != nil {
		log.Fatalf("closing with error: %v", err)
	}
}
