package config

import (
	"flag"
	"strconv"
	"strings"
)

var Port = 8080
var Host = "localhost"
var BaseAddress = "localhost:8080"

func GetServerAddress() string {
	return Host + ":" + strconv.Itoa(Port)
}

func InitFlags() {
	flag.Func("a", "set a server address", func(serverAddress string) error {
		hp := strings.Split(serverAddress, ":")

		port, err := strconv.Atoi(hp[1])
		if err != nil {
			return err
		}

		Host = hp[0]
		Port = port

		return nil
	})

	flag.StringVar(&BaseAddress, "b", "localhost:8080", "set a base address")

	flag.Parse()
}
