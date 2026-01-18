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

const domainUrl = "http://localhost:8080"
const longUrl = "https://practicum.yandex.ru/"

func TestCreateshortURL(t *testing.T) {
	t.Run("Create url", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(longUrl))
		request.Header.Set("Content-Type", "text/plain")

		w := httptest.NewRecorder()

		createshortURL(w, request)

		result := w.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusCreated, result.StatusCode)
		resBody, err := io.ReadAll(result.Body)
		require.NoError(t, err)
		require.NotEmpty(t, resBody)
		assert.Contains(t, string(resBody), domainUrl)
		assert.Contains(t, "text/plain", result.Header.Get("Content-Type"))
	})
}

func TestGetshortURL(t *testing.T) {
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
				return service.CreateshortURL(longUrl)
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
			getshortURL(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)

			if tt.wantStatus == http.StatusTemporaryRedirect {
				location := res.Header.Get("Location")
				assert.NotEmpty(t, location)
				assert.Contains(t, location, longUrl)
			}
		})
	}
}
