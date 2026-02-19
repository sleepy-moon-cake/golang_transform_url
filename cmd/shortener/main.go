package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
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
)

func main() {
	cfg := config.GetConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	db, err := db.NewSqlDB(ctx, cfg.DatabaseDSN)

	if err != nil {
		slog.Error("Database err", slog.String("err", err.Error()))
		panic(err)
	}

	logger.Init(cfg.LoggerLevel)

	if err := listenAndServe(cfg, db); err != nil {
		log.Fatal(err)
	}
}

func listenAndServe(cng *config.Config, db *db.DbSql) error {
	repository := repository.NewRepository(cng.FileStoragePath)
	service := service.NewService(repository)
	handler := handler.NewURLHandler(cng.BaseURLAddress, service)

	router := createRouter(handler)

	return http.ListenAndServe(cng.ServerAddress, logger.Logger(compressor.Compressor(router)))
}

func createRouter(handler *handler.URLHandle) http.Handler {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/{shortURL}", handler.GetShortURL)
		r.Post("/", handler.CreateShortURL)
	})
	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", handler.ShortenURL)
	})

	r.Route("/ping", func(r chi.Router) {
		r.Get("/", handler.Ping)
	})

	return r
}
