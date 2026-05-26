package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/config"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/handler"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/repository"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/service"
)

// Тест проверяет успешное создание короткого URL через POST-запрос к API
func TestCreateShortURL_Success(t *testing.T) {
	dir := t.TempDir()
	tmpFile := filepath.Join(dir, "test_storage.json")

	cfg := &config.Config{
		ServerAddress:   "localhost:8080",
		BaseURLAddress:  "http://localhost:8080",
		FileStoragePath: tmpFile,
	}

	repo, err := repository.NewRepository(cfg.FileStoragePath, nil)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	svc := service.NewService(repo)
	h := handler.NewURLHandler(cfg.BaseURLAddress, svc)
	router, err := createRouter(t.Context(), cfg, h)

	if err != nil {
		t.Fatalf("failed to create router: %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", strings.NewReader("https://yandex.ru"))
	req.Header.Set("Content-Type", "text/plain")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	if w.Body.Len() == 0 {
		t.Error("expected non-empty response body with short URL")
	}
}

// Тест проверяет, что эндпоинт /ping корректно возвращает статус 200 OK
func TestPing_WithoutDB(t *testing.T) {
	cfg := &config.Config{
		ServerAddress:  "localhost:8080",
		BaseURLAddress: "http://localhost:8080",
	}

	repo, _ := repository.NewRepository("", nil)
	svc := service.NewService(repo)
	h := handler.NewURLHandler(cfg.BaseURLAddress, svc)

	router, err := createRouter(t.Context(), cfg, h)
	if err != nil {
		t.Fatalf("failed to create router: %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// Ваш бенчмарк для замера скорости и снятия профилей памяти pprof
func BenchmarkCreateShortURL(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "bench_storage_*.json")
	if err != nil {
		b.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	cfg := &config.Config{
		ServerAddress:   "localhost:8080",
		BaseURLAddress:  "http://localhost:8080",
		FileStoragePath: tmpFile.Name(),
	}

	repo, err := repository.NewRepository(cfg.FileStoragePath, nil)
	if err != nil {
		b.Fatalf("failed to create repo in benchmark: %v", err)
	}
	svc := service.NewService(repo)
	h := handler.NewURLHandler(cfg.BaseURLAddress, svc)

	router, err := createRouter(b.Context(), cfg, h)

	if err != nil {
		b.Fatalf("failed to create router: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", strings.NewReader("https://yandex.ru"))
		req.Header.Set("Content-Type", "text/plain")

		router.ServeHTTP(w, req)
	}
}
