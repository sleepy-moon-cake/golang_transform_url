package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const domainURL = "http://localhost:8080"
const longURL = "https://practicum.yandex.ru/"

func TestUrlHandle_CreateshortURL(t *testing.T) {
	h := URLHandle{baseURL: domainURL}

	t.Run("Create url", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(longURL))
		request.Header.Set("Content-Type", "text/plain")

		w := httptest.NewRecorder()

		h.createShortURL(w, request)

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
	h := URLHandle{baseURL: domainURL}

	tests := []struct {
		name       string
		setup      func() string
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
			setup: func() string {
				return service.CreateShortURL(longURL)
			},
			wantStatus: http.StatusTemporaryRedirect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortID := tt.path
			if tt.setup != nil {
				shortID = tt.setup()
			}

			req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
			w := httptest.NewRecorder()

			h.getShortURL(w, req)

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
	h := URLHandle{baseURL: domainURL}

	t.Run("API shorten url - positive", func(t *testing.T) {
		jsonBody := `{"url":"` + longURL + `"}`

		request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(jsonBody))
		request.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		h.shortenURL(w, request)

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

		h.shortenURL(w, request)

		result := w.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
	})
}
