package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/repository"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/service"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/shared/contextkeys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const domainURL = "http://localhost:8080"
const longURL = "https://practicum.yandex.ru/"
const userIDKey = "userId"

func TestUrlHandle_CreateshortURL(t *testing.T) {
	h := newTestHandle(t)

	t.Run("Create url", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(longURL))
		request.Header.Set("Content-Type", "text/plain")

		w := httptest.NewRecorder()

		ctx := context.WithValue(request.Context(), contextkeys.UserId, userIDKey)

		request = request.WithContext(ctx)

		h.CreateShortURL(w, request)

		result := w.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusCreated, result.StatusCode)
		resBody, err := io.ReadAll(result.Body)
		require.NoError(t, err)
		require.NotEmpty(t, resBody)
		assert.Contains(t, string(resBody), domainURL)
		assert.Contains(t, result.Header.Get("Content-Type"), "text/plain")
	})
}

func TestUrlHandle_GetshortURL(t *testing.T) {
	h := newTestHandle(t)

	tests := []struct {
		name       string
		setup      func() (string, error)
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "path is empty - negative",
			path:       "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "storage is empty - negative",
			path:       "empty",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "get short url - positive",
			setup: func() (string, error) {
				ctx := context.WithValue(t.Context(), contextkeys.UserId, userIDKey)

				return h.service.CreateShortURL(ctx, longURL)
			},
			wantStatus: http.StatusTemporaryRedirect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortID := tt.path
			if tt.setup != nil {
				if value, err := tt.setup(); err == nil {
					shortID = value
				}
			}

			req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
			w := httptest.NewRecorder()

			ctx := context.WithValue(req.Context(), contextkeys.UserId, userIDKey)

			req = req.WithContext(ctx)

			h.GetShortURL(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)

			if tt.wantStatus == http.StatusTemporaryRedirect {
				location := res.Header.Get("Location")
				assert.NotEmpty(t, location)
				assert.Contains(t, location, longURL)
			}
		})
	}
}

func TestUrlHandle_ShortenURL(t *testing.T) {
	h := newTestHandle(t)

	t.Run("API shorten url - positive", func(t *testing.T) {
		jsonBody := `{"url":"` + longURL + `"}`

		request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(jsonBody))
		request.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		ctx := context.WithValue(request.Context(), contextkeys.UserId, userIDKey)

		request = request.WithContext(ctx)

		h.ShortenURL(w, request)

		result := w.Result()
		defer result.Body.Close()

		// Проверяем статус и заголовки
		assert.Equal(t, http.StatusCreated, result.StatusCode)
		assert.Contains(t, result.Header.Get("Content-Type"), "application/json")

		// Проверяем тело ответа
		resBody, err := io.ReadAll(result.Body)
		require.NoError(t, err)

		// Ожидаем JSON формата {"result": "..."}
		assert.Contains(t, string(resBody), `"result"`)
		require.NotEmpty(t, resBody)
	})

	t.Run("API shorten url - invalid json", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{invalid json}`))
		w := httptest.NewRecorder()

		ctx := context.WithValue(request.Context(), contextkeys.UserId, userIDKey)

		request = request.WithContext(ctx)

		h.ShortenURL(w, request)

		result := w.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
	})
}

func newTestHandle(t *testing.T) *URLHandle {
	tmpFile, err := os.CreateTemp("", "storage_*.json")
	require.NoError(t, err)

	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	repo := repository.NewRepository(tmpFile.Name(), nil)
	svc := service.NewService(repo)

	return &URLHandle{baseURL: domainURL, service: svc}
}

func TestUrlHandle_Batch(t *testing.T) {
	h := newTestHandle(t)

	t.Run("Batch shorten - success", func(t *testing.T) {
		// Подготавливаем JSON-запрос с двумя ссылками
		jsonBody := `[
			{"correlation_id": "first-id", "original_url": "https://google.com"},
			{"correlation_id": "second-id", "original_url": "https://yandex.ru"}
		]`

		request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(jsonBody))
		request.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx := context.WithValue(request.Context(), contextkeys.UserId, userIDKey)

		request = request.WithContext(ctx)

		h.Batch(w, request)

		result := w.Result()
		defer result.Body.Close()

		// 1. Проверяем статус 201 Created
		assert.Equal(t, http.StatusCreated, result.StatusCode)

		// 2. Проверяем заголовок Content-Type
		assert.Contains(t, result.Header.Get("Content-Type"), "application/json")

		// 3. Декодируем тело ответа для детальной проверки
		var response []model.ShortenURLBatchResponse
		err := json.NewDecoder(result.Body).Decode(&response)
		require.NoError(t, err)

		// 4. Проверяем длину и содержимое
		assert.Len(t, response, 2)

		// Проверяем, что correlation_id вернулись правильно
		assert.Equal(t, "first-id", response[0].CorrelationID)
		assert.Equal(t, "second-id", response[1].CorrelationID)

		// Проверяем, что сформированы короткие ссылки с твоим доменом
		assert.Contains(t, response[0].ShortURL, domainURL)
		assert.Contains(t, response[1].ShortURL, domainURL)
	})

	t.Run("Batch shorten - empty array", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(`[]`))
		request.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx := context.WithValue(request.Context(), contextkeys.UserId, userIDKey)

		request = request.WithContext(ctx)

		h.Batch(w, request)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "[]\n", w.Body.String()) // Encode добавляет перенос строки
	})

	t.Run("Batch shorten - invalid JSON", func(t *testing.T) {
		// Присылаем объект вместо массива
		request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(`{"id": "not a batch"}`))
		w := httptest.NewRecorder()

		ctx := context.WithValue(request.Context(), contextkeys.UserId, userIDKey)

		request = request.WithContext(ctx)

		h.Batch(w, request)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
