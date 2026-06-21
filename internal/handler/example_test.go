package handler_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/handler"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
)

// --- Заглушка (Mock) сервиса для демонстрации примеров без запуска реальной БД ---

type exampleMockService struct{}

func (m *exampleMockService) CreateShortURL(ctx context.Context, str string) (string, error) {
	return "h9O2kL8nP1", nil
}

func (m *exampleMockService) GetURLByCode(ctx context.Context, code string) (string, error) {
	return "https://yandex.ru", nil
}

func (m *exampleMockService) Ping(ctx context.Context) error {
	return nil
}

func (m *exampleMockService) Batch(ctx context.Context, batch []model.ShortenURLBatchRequest) ([]model.ShortenURLBatchResponse, error) {
	return []model.ShortenURLBatchResponse{
		{CorrelationID: "req-1", ShortURL: "h9O2kL8nP1"},
	}, nil
}

func (m *exampleMockService) GetUserShortUrls(ctx context.Context) ([]model.ShortenURLRecord, error) {
	return []model.ShortenURLRecord{
		{ShortURL: "h9O2kL8nP1", OriginalURL: "https://yandex.ru"},
	}, nil
}

func (m *exampleMockService) DeleteBatchUrl(ctx context.Context, shotUrls []string) error {
	return nil
}

func (e *exampleMockService) GetStats(ctx context.Context) (model.URLStats, error) {
	return model.URLStats{Urls: 1, Users: 1}, nil
}

// ExampleURLHandle_CreateShortURL демонстрирует практический пример работы
// эндпоинта POST /, принимающего оригинальный URL в формате обычного текста.
func ExampleURLHandle_CreateShortURL() {
	svc := &exampleMockService{}
	h := handler.NewURLHandler("http://localhost:8080", svc)

	// Симулируем отправку длинной ссылки plain text
	body := bytes.NewBufferString("https://yandex.ru")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	h.CreateShortURL(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Body:", w.Body.String())

	// Секция Output проверяется Go автоматически при запуске `go test`
	// Output:
	// Status: 201
	// Body: http://localhost:8080/h9O2kL8nP1
}

// ExampleURLHandle_ShortenURL демонстрирует пример работы REST API эндпоинта
// POST /api/shorten, принимающего и возвращающего JSON-данные.
func ExampleURLHandle_ShortenURL() {
	svc := &exampleMockService{}
	h := handler.NewURLHandler("http://localhost:8080", svc)

	// Формируем JSON-запрос
	jsonReq := `{"url":"https://yandex.ru"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(jsonReq))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.ShortenURL(w, req)

	fmt.Println("Status:", w.Code)
	// Убираем лишние пробелы и переносы для стабильного сравнения в тесте
	fmt.Println("JSON:", strings.TrimSpace(w.Body.String()))

	// Output:
	// Status: 201
	// JSON: {"result":"http://localhost:8080/h9O2kL8nP1"}
}

// ExampleURLHandle_Batch демонстрирует пример пакетного (массового) сокращения
// набора ссылок через эндпоинт POST /api/shorten/batch.
func ExampleURLHandle_Batch() {
	svc := &exampleMockService{}
	h := handler.NewURLHandler("http://localhost:8080", svc)

	jsonReq := `[{"correlation_id": "req-1", "original_url": "https://yandex.ru"}]`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString(jsonReq))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.Batch(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("JSON:", strings.TrimSpace(w.Body.String()))

	// Output:
	// Status: 201
	// JSON: [{"correlation_id":"req-1","short_url":"http://localhost:8080/h9O2kL8nP1"}]
}
