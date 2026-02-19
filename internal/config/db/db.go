package db

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBSQL struct {
	*sql.DB
}

func NewSQLDB(ctx context.Context, databaseDSN string) (*DBSQL, error) {
	db, err := sql.Open("pgx", databaseDSN)

	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return &DBSQL{db}, nil
}
