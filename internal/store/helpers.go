package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// execer is satisfied by both *sql.DB and *sql.Tx, letting a query helper
// run either standalone or inside a caller-managed transaction.
type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func mapScanErr(err error, msg string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return fmt.Errorf("%s: %w", msg, err)
}
