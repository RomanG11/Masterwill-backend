package db

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	pgx5 "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations applies every pending migration under migrations/ against
// conn, in order (files are named NNNN_description.up/down.sql — the
// numeric prefix must stay zero-padded and monotonically increasing, since
// that's the only thing the source driver sorts on). Progress is tracked in
// a schema_migrations table Postgres-side, so this is safe and cheap to
// call on every startup: a fresh database gets built up from scratch, an
// existing one only runs what it's missing.
//
// To add a migration: drop a new NNNN_description.up.sql (and matching
// .down.sql) next to 0001_init_schema.up.sql, then restart the API.
func RunMigrations(conn *sql.DB) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("load embedded migrations: %w", err)
	}

	driver, err := pgx5.WithInstance(conn, &pgx5.Config{})
	if err != nil {
		return fmt.Errorf("init migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "pgx5", driver)
	if err != nil {
		return fmt.Errorf("init migrator: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
