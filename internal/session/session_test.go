package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/shared/contextkeys"
)

// Тест проверяет, что при отсутствии куки создается новый пользователь,
// выставляется кука сессии, а UserID прокидывается в контекст
func TestJWTSession_NewUser(t *testing.T) {
	cfg := &SessionConfig{
		SecretKey: "super-secret-key",
		Name:      "TestSession",
		ExpiresAt: 1 * time.Hour,
	}

	// Хэндлер проверяет, что в контекст дошел сгенерированный UserID
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(contextkeys.UserId).(string)
		if !ok || userID == "" {
			t.Error("expected missing user ID in context to be generated and set")
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := JWTSession(cfg)(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Проверяем, что вернулся статус 200
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Проверяем, что кука сессии была добавлена в ответ
	cookies := resp.Cookies()
	var found bool
	for _, cookie := range cookies {
		if cookie.Name == cfg.Name {
			found = true
			if cookie.Value == "" {
				t.Error("expected non-empty session cookie value")
			}
		}
	}
	if !found {
		t.Errorf("expected session cookie '%s' to be set in response", cfg.Name)
	}
}

// Тест проверяет, что при наличии валидной куки сессия успешно парсится,
// а исходный UserID прокидывается в контекст без изменения куки
func TestJWTSession_ValidExistingUser(t *testing.T) {
	cfg := &SessionConfig{
		SecretKey: "super-secret-key",
		Name:      "TestSession",
		ExpiresAt: 1 * time.Hour,
	}

	expectedUserID := "permanent-user-id-123"
	tokenStr, err := buildJWTString(expectedUserID, cfg.SecretKey, cfg.ExpiresAt)
	if err != nil {
		t.Fatalf("failed to prepare JWT string for test: %v", err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(contextkeys.UserId).(string)
		if !ok {
			t.Fatal("expected user ID in context")
		}
		if userID != expectedUserID {
			t.Errorf("expected user ID '%s', got '%s'", expectedUserID, userID)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := JWTSession(cfg)(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: cfg.Name, Value: tokenStr})
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// Тест проверяет, что при передаче битой или невалидной куки
// middleware прерывает запрос и возвращает статус 401 Unauthorized
func TestJWTSession_InvalidToken(t *testing.T) {
	cfg := &SessionConfig{
		SecretKey: "super-secret-key",
		Name:      "TestSession",
		ExpiresAt: 1 * time.Hour,
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with an invalid token")
	})

	middleware := JWTSession(cfg)(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Передаем строку, которая не является валидным JWT
	req.AddCookie(&http.Cookie{Name: cfg.Name, Value: "invalid-token-string"})
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}
