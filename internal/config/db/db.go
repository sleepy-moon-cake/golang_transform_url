package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sleepy-moon-cake/golang_transform_url/migrations"
)

type DBSQL struct {
	*sql.DB
}

func NewSQLDB(ctx context.Context, databaseDSN string) (*DBSQL, error) {
	db, err := sql.Open("pgx", databaseDSN)

	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	slog.Info("DATABASE CONNECTED", slog.String("DNS", databaseDSN))

	if err := migrations.RunMigrations(db); err != nil {
		return nil, fmt.Errorf("migration db: %w", err)
	}

	slog.Info("DATABASE MIGRATION COMPLETE")

	return &DBSQL{db}, nil
}
