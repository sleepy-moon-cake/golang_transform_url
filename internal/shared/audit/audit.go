package audit

import (
	"context"
	"net/http"
	"time"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/shared/contextkeys"
)

type AuditConfig struct {
	URL  string
	Path string
}

func NewAuditMiddleware(ctx context.Context, cfg *AuditConfig) func(http.Handler) http.Handler {
	server := NewAuditService(ctx)

	if cfg.Path != "" {
		fileStorage := NewFileStorage(cfg.Path)
		subFile := NewSubscriber(100)
		server.Subscribe(subFile)

		go StartWorker(ctx, fileStorage, subFile, server)
	}

	if cfg.URL != "" {
		remoteStorage := NewRemoteStorage(cfg.URL)
		subRemote := NewSubscriber(100)
		server.Subscribe(subRemote)

		go StartWorker(ctx, remoteStorage, subRemote, server)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrapper := ResponseWrapper{ResponseWriter: w}

			next.ServeHTTP(&wrapper, r)

			if wrapper.StatusCode >= 200 && wrapper.StatusCode < 300 {
				userID, ok := r.Context().Value(contextkeys.UserId).(string)
				if !ok {
					userID = ""
				}

				action := "shorten"
				if r.Method == http.MethodGet {
					action = "follow"
				}

				server.Emit(Event{
					Ts:     int(time.Now().Unix()),
					Action: action,
					UserId: userID,
					URL:    r.URL.String(),
				})
			}
		})
	}
}

type ResponseWrapper struct {
	http.ResponseWriter
	StatusCode int
}

func (r *ResponseWrapper) Write(b []byte) (int, error) {
	if r.StatusCode == 0 {
		r.StatusCode = 200
	}

	return r.ResponseWriter.Write(b)
}

func (r *ResponseWrapper) WriteHeader(statusCode int) {
	r.StatusCode = statusCode

	r.ResponseWriter.WriteHeader(statusCode)
}
