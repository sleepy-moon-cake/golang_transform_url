package main

import (
	"fmt"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/config"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/handler"
)

func main() {
	cfg := config.GetConfig()

	if err := handler.ListenAndServe(cfg); err != nil {
		fmt.Println("Start server error")
	}
}
