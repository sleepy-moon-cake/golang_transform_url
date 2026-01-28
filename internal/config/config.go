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
	flag.StringVar(&config.BaseURLAddress, "b", "http://localhost:8080", "base shorted URL")

	flag.Parse()

	fmt.Println("Server address:", config.ServerAddress)
	fmt.Println("Base short URL:", config.BaseURLAddress)

	return &config
}
