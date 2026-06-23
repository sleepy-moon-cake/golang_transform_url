package subnet

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrustedSubnetMiddleware(t *testing.T) {
	successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("Positive - IP inside subnet allowed", func(t *testing.T) {
		mw, err := NewTrustedSubnetMiddleware("192.168.1.0/24")
		require.NoError(t, err)
		require.NotNil(t, mw)

		handler := mw(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/any-path", nil)
		req.Header.Set("X-Real-IP", "192.168.1.50")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Negative - IP outside subnet forbidden", func(t *testing.T) {
		mw, err := NewTrustedSubnetMiddleware("192.168.1.0/24")
		require.NoError(t, err)

		handler := mw(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/any-path", nil)
		req.Header.Set("X-Real-IP", "10.0.0.1")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Negative - Empty X-Real-IP header forbidden", func(t *testing.T) {
		mw, err := NewTrustedSubnetMiddleware("192.168.1.0/24")
		require.NoError(t, err)

		handler := mw(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/any-path", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Negative - Invalid X-Real-IP format forbidden", func(t *testing.T) {
		mw, err := NewTrustedSubnetMiddleware("192.168.1.0/24")
		require.NoError(t, err)

		handler := mw(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/any-path", nil)
		req.Header.Set("X-Real-IP", "not-an-ip-address")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Config - Empty mask returns nil middleware and no error", func(t *testing.T) {
		mw, err := NewTrustedSubnetMiddleware("")

		assert.NoError(t, err)
		assert.Nil(t, mw)
	})

	t.Run("Config - Invalid CIDR mask returns error at startup", func(t *testing.T) {
		mw, err := NewTrustedSubnetMiddleware("invalid-cidr")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parseCIDR")
		assert.Nil(t, mw)
	})
}
