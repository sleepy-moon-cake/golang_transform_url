package repository

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
)

type SQLRepository struct {
	mutex sync.Mutex
	db    *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Save(ctx context.Context, record model.ShortenURLRecord) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO urls (original, short) VALUES ($1, $2)",
		&record.OriginalURL, &record.ShortURL)

	return err
}

func (r *SQLRepository) Ping(ctx context.Context) error {
	return nil
}

func (r *SQLRepository) FindByCode(ctx context.Context, code string) (model.ShortenURLRecord, error) {
	return model.ShortenURLRecord{}, errors.New("")
}
