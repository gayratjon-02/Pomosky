// Package postgres implements the repository ports on top of PostgreSQL.
//
// SQL lives in queries/ and sqlc turns it into the sqlcgen package (run
// `make sqlc`). The repositories here translate between sqlcgen rows and the
// models, and map database errors to the apperr sentinels.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
	"github.com/pressly/goose/v3"

	"github.com/gayratjon-02/Pomosky/internal/apperr"
	"github.com/gayratjon-02/Pomosky/internal/store/postgres/sqlcgen"
	"github.com/gayratjon-02/Pomosky/migrations"
)

// DB wraps the shared database handle used by all repositories.
type DB struct {
	sql *sql.DB
	q   *sqlcgen.Queries
}

// New wraps an open *sql.DB.
func New(db *sql.DB) *DB { return &DB{sql: db, q: sqlcgen.New(db)} }

// Open connects to PostgreSQL, checks the connection and sets pool limits.
func Open(ctx context.Context, url string) (*DB, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("postgres: open: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return New(db), nil
}

// Migrate applies every pending embedded migration.
func (d *DB) Migrate(ctx context.Context) error {
	goose.SetBaseFS(migrations.FS)
	defer goose.SetBaseFS(nil)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.UpContext(ctx, d.sql, "."); err != nil {
		return fmt.Errorf("postgres: migrate: %w", err)
	}
	return nil
}

// SQL exposes the underlying handle (health checks, tests).
func (d *DB) SQL() *sql.DB { return d.sql }

// Close closes the connection pool.
func (d *DB) Close() error { return d.sql.Close() }

// Ping checks that the database answers.
func (d *DB) Ping(ctx context.Context) error { return d.sql.PingContext(ctx) }

// mapErr translates database errors into the shared sentinels.
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return apperr.ErrNotFound
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" { // unique_violation
		return apperr.ErrConflict
	}
	return err
}

func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}
