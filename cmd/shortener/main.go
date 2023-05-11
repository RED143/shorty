package main

import (
	"fmt"
	"shorty/internal/app/server"
)

func main() {
	err := server.Start()

	if err != nil {
		fmt.Errorf("failed to start server: %w", err)
		panic(err)
	}
}
