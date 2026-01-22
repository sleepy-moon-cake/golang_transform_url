package config

import (
	"flag"
	"fmt"
)

type Config struct {
	ServerAddress  string
	BaseURLAddress string
}

func GetConfig() *Config {
	var config Config

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "server address")
	flag.StringVar(&config.BaseURLAddress, "b", "http://localhost:8000", "base shorted URL")

	flag.Parse()

	fmt.Println("Server address:", config.BaseURLAddress)
	fmt.Println("Base short URL:", config.ServerAddress)

	return &config
}
