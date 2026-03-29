package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
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
			os.Exit(1)
		}

		database = dbSQL
	}

	logger.Init(cfg.LoggerLevel)

	if err := listenAndServe(cfg, database); err != nil {
		log.Fatal(err)
	}
}

func listenAndServe(cng *config.Config, db *db.DBSQL) error {
	repository := repository.NewRepository(cng.FileStoragePath, db)
	service := service.NewService(repository)
	handler := handler.NewURLHandler(cng.BaseURLAddress, service)

	router := createRouter(handler)

	return http.ListenAndServe(cng.ServerAddress, router)
}

func createRouter(handler *handler.URLHandle) http.Handler {
	r := chi.NewRouter()
	r.Use(session.JWTSession(&session.SessionConfig{
		Name:      "Session",
		SecretKey: "SecretKey",
		ExpiresAt: 3 * time.Hour,
	}))
	r.Use(logger.Logger)
	r.Use(compressor.Compressor)

	r.Route("/", func(r chi.Router) {
		r.Get("/{shortURL}", handler.GetShortURL)
		r.Post("/", handler.CreateShortURL)
	})
	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", handler.ShortenURL)
		r.Post("/shorten/batch", handler.Batch)
	})

	r.Get("/ping", handler.Ping)

	r.Get("/api/user/urls", handler.GetUserShortUrls)

	return r
}
