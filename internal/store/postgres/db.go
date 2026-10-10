// Package postgres implements the repository ports on top of PostgreSQL.
//
// The package works with a database/sql handle that is injected from the
// composition root; the driver (e.g. pgx/stdlib) is imported in cmd/server
// when it is added. SQL lives in queries/ and is turned into Go by sqlc.
package postgres

import "database/sql"

// DB wraps the shared database handle used by all repositories.
type DB struct {
	sql *sql.DB
}

// New wraps an open *sql.DB.
func New(db *sql.DB) *DB { return &DB{sql: db} }

// SQL exposes the underlying handle (health checks, migrations).
func (d *DB) SQL() *sql.DB { return d.sql }
