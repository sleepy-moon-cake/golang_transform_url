package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var MigrationsFS embed.FS

func RunMigrations(db *sql.DB) error {
	sourceDriver, err := iofs.New(MigrationsFS, ".")
	if err != nil {
		slog.Error("SourceDriver error")
		return err
	}

	dbDriver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		slog.Error("dbDriver error")
		return err
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		sourceDriver,
		"pgx",
		dbDriver,
	)
	if err != nil {
		slog.Error("Migration error")
		return err
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
