package session

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/shared/contextkeys"
)

var ErrParseToken = errors.New("Parse error")

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type SessionConfig struct {
	SecretKey string
	Name      string
	ExpiresAt time.Duration
}

func JWTSession(cfg *SessionConfig) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cfg.Name)

			if err != nil || cookie == nil || cookie.Value == "" {
				newUserID := uuid.NewString()
				jwtString, err := buildJWTString(newUserID, cfg.SecretKey, cfg.ExpiresAt)

				if err != nil {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}

				http.SetCookie(w, createSessionCookie(jwtString, cfg.Name))

				ctx := context.WithValue(r.Context(), contextkeys.UserId, newUserID)
				r = r.WithContext(ctx)
			} else {
				userID, err := parseJWTString(cookie.Value, cfg.SecretKey)

				if err != nil {
					slog.Error("Parse JWT err", "err", err)
					http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
					return
				}

				ctx := context.WithValue(r.Context(), contextkeys.UserId, userID)
				r = r.WithContext(ctx)
			}

			h.ServeHTTP(w, r)
		})
	}
}

func buildJWTString(userId string, secretKey string, expire time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
		},
		UserID: userId,
	})

	signedString, err := token.SignedString([]byte(secretKey))

	if err != nil {
		return "", fmt.Errorf("buildJWTString: %w", err)
	}

	return signedString, nil
}

func parseJWTString(jwtString string, secretKey string) (string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(jwtString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrParseToken, err)
	}

	if !token.Valid {
		return "", ErrParseToken
	}

	return claims.UserID, nil
}

func createSessionCookie(value string, name string) *http.Cookie {
	return &http.Cookie{
		Name:  name,
		Value: value,
	}
}
