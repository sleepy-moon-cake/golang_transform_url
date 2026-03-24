package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

func (r *SQLRepository) Save(ctx context.Context, record model.ShortenURLRecord) (model.ShortenURLRecord, error) {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO urls (original_url, short_url) VALUES ($1, $2)",
		record.OriginalURL,
		record.ShortURL,
	)

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		var existingRecord model.ShortenURLRecord

		err = r.db.QueryRowContext(
			ctx,
			`SELECT original_url, short_url FROM urls WHERE original_url = $1`,
			record.OriginalURL,
		).Scan(&existingRecord.OriginalURL, &existingRecord.ShortURL)

		if err != nil {
			return model.ShortenURLRecord{}, fmt.Errorf("database save: %w", err)
		}

		return existingRecord, ErrURLConflict
	}

	if err != nil {
		return model.ShortenURLRecord{}, fmt.Errorf("database save: %w", err)
	}

	return record, nil
}

func (r *SQLRepository) Ping(ctx context.Context) error {
	if err := r.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database pind: %w", err)
	}

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
		return model.ShortenURLRecord{}, fmt.Errorf("database findByCode: %w", err)
	}

	return record, nil
}

func (r *SQLRepository) Batch(ctx context.Context, shortenURLRecords []model.ShortenURLRecord) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	values := make([]string, 0, len(shortenURLRecords))
	args := make([]any, 0, len(shortenURLRecords)*2)

	for i, record := range shortenURLRecords {
		values = append(values, fmt.Sprintf("($%d,$%d)", i*2+1, i*2+2))
		args = append(args, record.ShortURL, record.OriginalURL)
	}

	query := fmt.Sprintf(
		"INSERT INTO urls (short_url, original_url) VALUES %s",
		strings.Join(values, ","),
	)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("database batch: %w", err)
	}

	return tx.Commit()
}
