package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/compressor"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/config"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/config/db"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/handler"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/logger"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/repository"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/service"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/session"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/shared/audit"
)

func main() {
	cfg := config.GetConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var database *db.DBSQL

	if cfg.DatabaseDSN != "" {
		dbSQL, err := db.NewSQLDB(ctx, cfg.DatabaseDSN)

		if err != nil {
			slog.Error("Database err", slog.String("err", err.Error()))

			log.Fatalf("Database err: %v", err)
		}

		database = dbSQL
	}

	logger.Init(cfg.LoggerLevel)

	if err := listenAndServe(ctx, cfg, database); err != nil {
		log.Fatal(err)
	}
}

func listenAndServe(ctx context.Context, cng *config.Config, db *db.DBSQL) error {
	repository, err := repository.NewRepository(cng.FileStoragePath, db)
	if err != nil {
		return fmt.Errorf("repository init failed: %w", err)
	}

	service := service.NewService(repository)
	handler := handler.NewURLHandler(cng.BaseURLAddress, service)

	router, err := createRouter(ctx, cng, handler)

	if err != nil {
		return err
	}

	return http.ListenAndServe(cng.ServerAddress, router)
}

func createRouter(ctx context.Context, cfg *config.Config, handler *handler.URLHandle) (http.Handler, error) {
	auditMW, err := audit.NewAuditMiddleware(ctx, &audit.AuditConfig{URL: cfg.AuditURL, Path: cfg.AuditFile})

	if err != nil {
		return nil, fmt.Errorf("failed to create audit middleware: %w", err)
	}

	r := chi.NewRouter()
	r.Use(session.JWTSession(&session.SessionConfig{
		Name:      "Session",
		SecretKey: "SecretKey",
		ExpiresAt: 3 * time.Hour,
	}))
	r.Use(logger.Logger)
	r.Use(compressor.Compressor)

	r.Route("/", func(r chi.Router) {
		r.With(auditMW).Get("/{shortURL}", handler.GetShortURL)
		r.With(auditMW).Post("/", handler.CreateShortURL)
	})

	r.Route("/api", func(r chi.Router) {
		r.With(auditMW).Post("/shorten", handler.ShortenURL)
		r.Post("/shorten/batch", handler.Batch)
	})

	r.Get("/ping", handler.Ping)

	r.Get("/api/user/urls", handler.GetUserShortUrls)

	r.Delete("/api/user/urls", handler.DeleteBatch)

	r.Mount("/debug", http.DefaultServeMux)

	return r, nil
}
