package main

import (
	"context"
	"fmt"
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

	db, err := db.NewSQLDB(ctx, cfg.DatabaseDSN)

	fmt.Println("START__INFFF")
	fmt.Println(db)
	fmt.Println(err)
	fmt.Println("END___INFFF")

	if err != nil {
		if cfg.DatabaseDSN != "" {
			slog.Error("Database err", slog.String("err", err.Error()))
			panic(err)
		}
		db = nil
	}

	logger.Init(cfg.LoggerLevel)

	if err := listenAndServe(cfg, db); err != nil {
		log.Fatal(err)
	}
}

func listenAndServe(cng *config.Config, db *db.DBSQL) error {
	repository := repository.NewRepository(cng.FileStoragePath, db)
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
