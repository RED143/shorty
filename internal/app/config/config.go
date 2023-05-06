package config

import (
	"errors"
	"flag"
	"net"
	"os"
	"strconv"
	"strings"
)

var Port = 8080
var Host = "localhost"
var BaseAddress = "localhost:8080"
var Scheme = "http://"

func GetServerAddress() string {
	return Host + ":" + strconv.Itoa(Port)
}

func InitConfig() {
	getServerAddress()
	getBaseAddress()
	flag.Parse()
}

func getServerAddress() error {

	flag.Func("a", "set a server address", func(serverAddress string) error {
		defaultAddress := serverAddress
		envAddress := os.Getenv("SERVER_ADDRESS")

		if envAddress != "" {
			defaultAddress = envAddress
		}

		host, port, err := parseHostAndPort(defaultAddress)
		if err != nil {
			return err
		}

		Host = host
		Port = port

		return nil
	})

	return nil
}

func getBaseAddress() error {
	defaultAddress := "localhost:8080"
	envBaseURL := os.Getenv("BASE_URL")

	if envBaseURL != "" {
		defaultAddress = envBaseURL
	}

	flag.StringVar(&BaseAddress, "b", defaultAddress, "set a base address")

	return nil
}

func parseHostAndPort(s string) (string, int, error) {
	s = strings.Replace(s, Scheme, "", 1)
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return "", 0, errors.New("invalid parsing url")
	}

	parsedPort, err := strconv.Atoi(strings.ReplaceAll(port, "/", ""))
	if err != nil {
		return "", 0, errors.New("invalid port")
	}

	return host, parsedPort, err
}
