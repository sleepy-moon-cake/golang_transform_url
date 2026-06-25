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
		"INSERT INTO urls (original_url, short_url, user_id) VALUES ($1, $2, $3)",
		record.OriginalURL,
		record.ShortURL,
		record.UserUUID,
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
		"SELECT original_url, short_url, is_deleted  FROM urls WHERE short_url = $1",
		code,
	).Scan(&record.OriginalURL, &record.ShortURL, &record.DeletedFlag)

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
		values = append(values, fmt.Sprintf("($%d,$%d,$%d)", i*3+1, i*3+2, i*3+3))
		args = append(args, record.ShortURL, record.OriginalURL, record.UserUUID)
	}

	query := fmt.Sprintf(
		"INSERT INTO urls (short_url, original_url, user_id) VALUES %s",
		strings.Join(values, ","),
	)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("database batch: %w", err)
	}

	return tx.Commit()
}

func (r *SQLRepository) GetURLsByUserID(ctx context.Context, userID string) ([]model.ShortenURLRecord, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT short_url, original_url, user_id FROM urls WHERE user_id = $1",
		userID,
	)

	if err != nil {
		return nil, fmt.Errorf("getURLsByUserID: %w", err)
	}
	defer rows.Close()

	var records = make([]model.ShortenURLRecord, 0)

	for rows.Next() {
		var record model.ShortenURLRecord

		if err := rows.Scan(
			&record.ShortURL,
			&record.OriginalURL,
			&record.UserUUID,
		); err != nil {
			return nil, fmt.Errorf("getURLsByUserID: scanning: %w", err)
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("getURLsByUserID:rows.Err: %w", err)
	}

	return records, nil
}

func (r *SQLRepository) DeleteBatch(ctx context.Context, urls []model.ShortenUrlDeleteRecord) error {
	if len(urls) == 0 {
		return nil
	}

	values := make([]string, 0, len(urls))
	args := make([]any, 0, len(urls)*2)

	for i, v := range urls {
		// Формируем ($1, $2), ($3, $4)...
		values = append(values, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		args = append(args, v.ShortURL, v.UserUUID)
	}

	query := fmt.Sprintf(`
		UPDATE urls AS u
		SET is_deleted = true
		FROM (VALUES %s) AS d(short_url, user_id)
		WHERE u.short_url = d.short_url AND u.user_id = d.user_id`,
		strings.Join(values, ","),
	)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("database delete batch: %w", err)
	}

	return nil
}

func (r *SQLRepository) GetStats(ctx context.Context) (model.URLStats, error) {
	var stats = model.URLStats{}

	err := r.db.QueryRowContext(ctx, "SELECT COUNT(id) as urls_count,  COUNT(DISTINCT user_id) as users_count from urls").Scan(&stats.Urls, &stats.Users)
	if err != nil {
		return stats, fmt.Errorf("database getstats:%w", err)
	}

	return stats, nil
}
