package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	ServerAddress  string
	BaseURLAddress string
}

func GetConfig() *Config {
	var config Config

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "server address")
	flag.StringVar(&config.BaseURLAddress, "b", "http://localhost:8080", "base shorted URL")

	flag.Parse()

	if evnServerAddres := os.Getenv("SERVER_ADDRESS"); evnServerAddres != "" {
		config.ServerAddress = evnServerAddres
	}

	if evnBaseURLAddress := os.Getenv("BASE_URL"); evnBaseURLAddress != "" {
		config.BaseURLAddress = evnBaseURLAddress
	}

	fmt.Println("Server address:", config.ServerAddress)
	fmt.Println("Base short URL:", config.BaseURLAddress)

	return &config
}
