package main

import (
	"log"
	"shorty/internal/app/server"
)

func main() {
	err := server.Start()

	if err != nil {
		log.Fatal("Closing with errorL", err)
	}
}
