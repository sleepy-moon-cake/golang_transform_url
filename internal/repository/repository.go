package repository

import (
	"context"
	"errors"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/config/db"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
)

type Repository interface {
	Ping(ctx context.Context) error
	FindByCode(ctx context.Context, code string) (model.ShortenURLRecord, error)
	Save(ctx context.Context, record model.ShortenURLRecord) (model.ShortenURLRecord, error)
	Batch(ctx context.Context, shortenURLRecords []model.ShortenURLRecord) error
}

func NewRepository(filePath string, db *db.DBSQL) Repository {
	if db != nil {
		return NewSQLRepository(db.DB)
	}

	return NewFileRepository(filePath)
}

var ErrURLConflict = errors.New("DB already has that URL")

var ErrNotFound = errors.New("not found")
