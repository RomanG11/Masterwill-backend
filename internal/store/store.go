// Package store is the persistence layer: thin, SQL-first repositories over
// the shared *sql.DB. Handlers depend on this package, never on database/sql
// directly, so the query shape for each domain lives in one place.
package store

import "database/sql"

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}
