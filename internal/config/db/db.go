package db

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DbSQL struct {
	*sql.DB
}

func NewSQLDB(ctx context.Context, databaseDSN string) (*DbSQL, error) {
	db, err := sql.Open("pgx", databaseDSN)

	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return &DbSQL{db}, nil
}
