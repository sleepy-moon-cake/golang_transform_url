package db

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DbSql struct {
	db *sql.DB
}

func NewSqlDB(ctx context.Context, databaseDSN string) (*DbSql, error) {
	db, err := sql.Open("pgx", databaseDSN)

	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return &DbSql{db: db}, nil
}
