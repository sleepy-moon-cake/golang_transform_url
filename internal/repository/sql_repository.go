package repository

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
		"INSERT INTO urls (original_url, short_url) VALUES ($1, $2)",
		record.OriginalURL,
		record.ShortURL,
	)

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		return ErrURLConflict
	}

	return err
}

func (r *SQLRepository) Ping(ctx context.Context) error {
	return nil
}

func (r *SQLRepository) FindByCode(ctx context.Context, code string) (model.ShortenURLRecord, error) {
	var record model.ShortenURLRecord

	err := r.db.QueryRowContext(
		ctx,
		"SELECT original_url, short_url FROM urls WHERE short_url = $1",
		code,
	).Scan(&record.OriginalURL, &record.ShortURL)

	if err != nil {
		return model.ShortenURLRecord{}, err
	}

	return record, nil
}

func (r *SQLRepository) Batch(ctx context.Context, shortenURLRecords []model.ShortenURLRecord) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, record := range shortenURLRecords {
		_, err := tx.ExecContext(ctx, "INSERT INTO urls (short_url,original_url) VALUES ($1, $2)",
			record.ShortURL, record.OriginalURL)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
