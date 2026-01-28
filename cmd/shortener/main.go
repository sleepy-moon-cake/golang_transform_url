package main

import (
	"log"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/config"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/handler"
)

func main() {
	cfg := config.GetConfig()

	if err := handler.ListenAndServe(cfg); err != nil {
		log.Fatal(err)
	}
}
