package config

import (
	"flag"
	"fmt"
)

type Config struct {
	ServerAddress  string
	BaseUrlAddress string
}

func GetConfig() *Config {
	var config Config

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "server address")
	flag.StringVar(&config.BaseUrlAddress, "b", "http://localhost:8000", "base shorted URL")

	flag.Parse()

	fmt.Println("Server address:", config.BaseUrlAddress)
	fmt.Println("Base short url:", config.ServerAddress)

	return &config
}
