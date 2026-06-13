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

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()

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

	if err := run(ctx, cfg, database); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, cng *config.Config, db *db.DBSQL) error {
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

	srv := &http.Server{
		Addr:         cng.ServerAddress,
		Handler:      router,
		ReadTimeout:  5 * time.Second,   // время на чтение запроса
		WriteTimeout: 10 * time.Second,  // время на отправку ответа
		IdleTimeout:  120 * time.Second, // время удержания соединения (Keep-Alive)
	}

	if cng.Secure {
		slog.Info("Start listen server in secure mode")
		return srv.ListenAndServeTLS("cert.pem", "key.pem")
	}
	slog.Info("Start listen server")
	return srv.ListenAndServe()
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

func printBuildInfo() {
	version := "N/A"
	if buildVersion != "" {
		version = buildVersion
	}

	date := "N/A"
	if buildDate != "" {
		date = buildDate
	}

	commit := "N/A"
	if buildCommit != "" {
		commit = buildCommit
	}

	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)
}
