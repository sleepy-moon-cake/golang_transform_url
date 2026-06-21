package subnet

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrustedSubnetMiddleware(t *testing.T) {
	// Создаем финальный хендлер-заглушку, до которого запрос дойдет только при успехе
	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("Positive - IP inside subnet allowed", func(t *testing.T) {
		// Инициализируем middleware с маской подсети
		mw := NewTrustedSubnetMiddleware("192.168.1.0/24")
		handler := mw(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/any-path", nil)
		req.Header.Set("X-Real-IP", "192.168.1.50") // IP входит в подсеть
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Negative - IP outside subnet forbidden", func(t *testing.T) {
		mw := NewTrustedSubnetMiddleware("192.168.1.0/24")
		handler := mw(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/any-path", nil)
		req.Header.Set("X-Real-IP", "10.0.0.1") // IP НЕ входит в подсеть
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Negative - Empty X-Real-IP header forbidden", func(t *testing.T) {
		mw := NewTrustedSubnetMiddleware("192.168.1.0/24")
		handler := mw(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/any-path", nil)
		// Заголовок X-Real-IP не передаем вообще
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Negative - Invalid X-Real-IP format forbidden", func(t *testing.T) {
		mw := NewTrustedSubnetMiddleware("192.168.1.0/24")
		handler := mw(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/any-path", nil)
		req.Header.Set("X-Real-IP", "not-an-ip-address") // Сломанный IP
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Negative - Empty mask config forbidden", func(t *testing.T) {
		mw := NewTrustedSubnetMiddleware("") // Пустая маска в конфиге
		handler := mw(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/any-path", nil)
		req.Header.Set("X-Real-IP", "192.168.1.50")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Negative - Invalid CIDR mask forbidden", func(t *testing.T) {
		mw := NewTrustedSubnetMiddleware("invalid-cidr") // Некорректная маска
		handler := mw(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/any-path", nil)
		req.Header.Set("X-Real-IP", "192.168.1.50")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
