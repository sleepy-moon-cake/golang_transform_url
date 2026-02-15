package main

import (
	"log"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/config"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/handler"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/logger"
)

func main() {
	cfg := config.GetConfig()

	logger.Init(cfg.LoggerLevel)

	if err := handler.ListenAndServe(cfg); err != nil {
		log.Fatal(err)
	}
}
