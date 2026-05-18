package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
const longURL = "https://yandex.ru"
const userIDKey = "userId"

// --- MOCK URL SERVICE ---
type mockURLService struct {
	CreateShortURLFunc   func(ctx context.Context, str string) (string, error)
	GetURLByCodeFunc     func(ctx context.Context, code string) (string, error)
	PingFunc             func(ctx context.Context) error
	BatchFunc            func(ctx context.Context, shorURLBatch []model.ShortenURLBatchRequest) ([]model.ShortenURLBatchResponse, error)
	GetUserShortUrlsFunc func(ctx context.Context) ([]model.ShortenURLRecord, error)
	DeleteBatchUrlFunc   func(ctx context.Context, shotUrls []string) error
}

func (m *mockURLService) CreateShortURL(ctx context.Context, str string) (string, error) {
	return m.CreateShortURLFunc(ctx, str)
}
func (m *mockURLService) GetURLByCode(ctx context.Context, code string) (string, error) {
	return m.GetURLByCodeFunc(ctx, code)
}
func (m *mockURLService) Ping(ctx context.Context) error {
	return m.PingFunc(ctx)
}
func (m *mockURLService) Batch(ctx context.Context, shorURLBatch []model.ShortenURLBatchRequest) ([]model.ShortenURLBatchResponse, error) {
	return m.BatchFunc(ctx, shorURLBatch)
}
func (m *mockURLService) GetUserShortUrls(ctx context.Context) ([]model.ShortenURLRecord, error) {
	return m.GetUserShortUrlsFunc(ctx)
}
func (m *mockURLService) DeleteBatchUrl(ctx context.Context, shotUrls []string) error {
	return m.DeleteBatchUrlFunc(ctx, shotUrls)
}

// --- TESTS FOR CreateShortURL ---

func TestUrlHandle_CreateShortURL_Scenarios(t *testing.T) {
	t.Run("Positive - Success 201", func(t *testing.T) {
		svc := &mockURLService{
			CreateShortURLFunc: func(ctx context.Context, str string) (string, error) {
				return "short123", nil
			},
		}
		h := NewURLHandler(domainURL, svc)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(longURL))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		h.CreateShortURL(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, domainURL+"/short123", w.Body.String())
	})

	t.Run("Negative - Wrong Content-Type 400", func(t *testing.T) {
		h := NewURLHandler(domainURL, &mockURLService{})
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(longURL))
		req.Header.Set("Content-Type", "application/json") // Неверный тип
		w := httptest.NewRecorder()

		h.CreateShortURL(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Positive - URL Conflict 409", func(t *testing.T) {
		svc := &mockURLService{
			CreateShortURLFunc: func(ctx context.Context, str string) (string, error) {
				return "existingID", repository.ErrURLConflict
			},
		}
		h := NewURLHandler(domainURL, svc)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(longURL))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		h.CreateShortURL(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Equal(t, domainURL+"/existingID", w.Body.String())
	})
}

// --- TESTS FOR GetShortURL ---

func TestUrlHandle_GetShortURL_Scenarios(t *testing.T) {
	t.Run("Negative - Empty Path 400", func(t *testing.T) {
		h := NewURLHandler(domainURL, &mockURLService{})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		h.GetShortURL(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Positive - Temporary Redirect 307", func(t *testing.T) {
		svc := &mockURLService{
			GetURLByCodeFunc: func(ctx context.Context, code string) (string, error) {
				return longURL, nil
			},
		}
		h := NewURLHandler(domainURL, svc)

		req := httptest.NewRequest(http.MethodGet, "/short123", nil)
		w := httptest.NewRecorder()

		h.GetShortURL(w, req)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, longURL, w.Header().Get("Location"))
	})

	t.Run("Negative - URL Been Deleted 410 Gone", func(t *testing.T) {
		svc := &mockURLService{
			GetURLByCodeFunc: func(ctx context.Context, code string) (string, error) {
				return "", service.ErrURLBeenDeleted
			},
		}
		h := NewURLHandler(domainURL, svc)

		req := httptest.NewRequest(http.MethodGet, "/deleted123", nil)
		w := httptest.NewRecorder()

		h.GetShortURL(w, req)

		assert.Equal(t, http.StatusGone, w.Code)
	})
}

// --- TESTS FOR ShortenURL (JSON API) ---

func TestUrlHandle_ShortenURL_Scenarios(t *testing.T) {
	t.Run("Positive - JSON Shorten Success 201", func(t *testing.T) {
		svc := &mockURLService{
			CreateShortURLFunc: func(ctx context.Context, str string) (string, error) {
				return "json123", nil
			},
		}
		h := NewURLHandler(domainURL, svc)

		reqBody := `{"url":"https://yandex.ru"}`
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		h.ShortenURL(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

		var res model.ShortenURLResponse
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.Equal(t, domainURL+"/json123", res.Result)
	})

	t.Run("Positive - JSON Shorten Conflict 409", func(t *testing.T) {
		svc := &mockURLService{
			CreateShortURLFunc: func(ctx context.Context, str string) (string, error) {
				return "conflictedJSON", repository.ErrURLConflict
			},
		}
		h := NewURLHandler(domainURL, svc)

		reqBody := `{"url":"https://yandex.ru"}`
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		h.ShortenURL(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

// --- TESTS FOR Ping ---

func TestUrlHandle_Ping(t *testing.T) {
	t.Run("Success 200", func(t *testing.T) {
		svc := &mockURLService{
			PingFunc: func(ctx context.Context) error { return nil },
		}
		h := NewURLHandler(domainURL, svc)
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		h.Ping(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Failure 500", func(t *testing.T) {
		svc := &mockURLService{
			PingFunc: func(ctx context.Context) error { return errors.New("db disconnect") },
		}
		h := NewURLHandler(domainURL, svc)
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		h.Ping(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// --- TESTS FOR GetUserShortUrls ---

func TestUrlHandle_GetUserShortUrls_Scenarios(t *testing.T) {
	t.Run("Negative - Unauthorized 401", func(t *testing.T) {
		h := NewURLHandler(domainURL, &mockURLService{})
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil) // Нет UserId в контексте
		w := httptest.NewRecorder()

		h.GetUserShortUrls(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Positive - No Content 204", func(t *testing.T) {
		svc := &mockURLService{
			GetUserShortUrlsFunc: func(ctx context.Context) ([]model.ShortenURLRecord, error) {
				return []model.ShortenURLRecord{}, nil // Ссылок нет
			},
		}
		h := NewURLHandler(domainURL, svc)
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserId, userIDKey)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		h.GetUserShortUrls(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Positive - Success 200 with Records", func(t *testing.T) {
		svc := &mockURLService{
			GetUserShortUrlsFunc: func(ctx context.Context) ([]model.ShortenURLRecord, error) {
				return []model.ShortenURLRecord{
					{ShortURL: "id1", OriginalURL: "http://site1.com"},
				}, nil
			},
		}
		h := NewURLHandler(domainURL, svc)
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserId, userIDKey)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		h.GetUserShortUrls(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

		var records []model.ShortenURLRecord
		err := json.Unmarshal(w.Body.Bytes(), &records)
		require.NoError(t, err)
		assert.Len(t, records, 1)
		assert.Equal(t, domainURL+"/id1", records[0].ShortURL)
	})
}

// --- TESTS FOR DeleteBatch ---

func TestUrlHandle_DeleteBatch(t *testing.T) {
	t.Run("Success 202 Accepted", func(t *testing.T) {
		svc := &mockURLService{
			DeleteBatchUrlFunc: func(ctx context.Context, shotUrls []string) error {
				assert.Len(t, shotUrls, 2)
				assert.Equal(t, "code1", shotUrls[0])
				return nil
			},
		}
		h := NewURLHandler(domainURL, svc)

		reqBody := `["code1", "code2"]`
		req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		h.DeleteBatch(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)
	})
}
