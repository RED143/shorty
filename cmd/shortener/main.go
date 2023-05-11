package main

import (
	"shorty/internal/app/server"
)

func main() {
	err := server.Start()

	if err != nil {
		panic(err)
	}
}
