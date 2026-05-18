package main

import (
	"context"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/config"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/handler"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/repository"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/service"
)

func BenchmarkCreateShortURL(b *testing.B) {
	// 1. Создаем временный файл один раз для всего бенчмарка
	tmpFile, err := os.CreateTemp("", "bench_storage_*.json")
	if err != nil {
		b.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Удаляем файл после завершения бенчмарка

	cfg := &config.Config{
		ServerAddress:   "localhost:8080",
		BaseURLAddress:  "http://localhost:8080",
		FileStoragePath: tmpFile.Name(),
	}

	// 2. Инициализируем зависимости с файловым путем
	repo, err := repository.NewRepository(cfg.FileStoragePath, nil)
	if err != nil {
		b.Fatalf("failed to create repo in benchmark: %v", err)
	}
	svc := service.NewService(repo)
	h := handler.NewURLHandler(cfg.BaseURLAddress, svc)

	router := createRouter(context.Background(), cfg, h)

	b.ResetTimer() // Сбрасываем таймер, чтобы создание файла не влияло на метрики
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", strings.NewReader("https://yandex.ru"))

		req.Header.Set("Content-Type", "text/plain")

		router.ServeHTTP(w, req)
	}
}
