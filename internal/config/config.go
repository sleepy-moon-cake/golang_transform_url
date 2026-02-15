package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	ServerAddress   string
	BaseURLAddress  string
	LoggerLevel     string
	FileStoragePath string
}

func GetConfig() *Config {
	var config Config

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "server address")
	flag.StringVar(&config.BaseURLAddress, "b", "http://localhost:8080", "base shorted URL")
	flag.StringVar(&config.LoggerLevel, "l", "info", "log level")
	flag.StringVar(&config.FileStoragePath, "f", "/storage.json", "path to storage file")

	flag.Parse()

	if evnServerAddres := os.Getenv("SERVER_ADDRESS"); evnServerAddres != "" {
		config.ServerAddress = evnServerAddres
	}

	if evnBaseURLAddress := os.Getenv("BASE_URL"); evnBaseURLAddress != "" {
		config.BaseURLAddress = evnBaseURLAddress
	}

	if evnLoggerLevel := os.Getenv("LOG_LEVEL"); evnLoggerLevel != "" {
		config.LoggerLevel = evnLoggerLevel
	}

	if fileStoragePath := os.Getenv("FILE_STORAGE_PATH"); fileStoragePath != "" {
		config.FileStoragePath = fileStoragePath
	}

	fmt.Println("Server address:", config.ServerAddress)
	fmt.Println("Base short URL:", config.BaseURLAddress)
	fmt.Println("Log level:", config.LoggerLevel)

	return &config
}
