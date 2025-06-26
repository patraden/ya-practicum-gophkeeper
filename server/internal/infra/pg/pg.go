package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
)

const (
	pgPoolMaxSize     = 30
	pgPoolMinSize     = 3
	pgMaxConnIdleTime = 30 * time.Second
	pgMaxConnLifeTime = 5 * time.Second
)

// QueryFunc defines a function that accepts a pg-generated Queries instance
// and performs a set of SQL operations.
type QueryFunc = func(*Queries) error

// DB wraps a PostgreSQL connection pool and provides methods for common
// database lifecycle operations.
type DB struct {
	ConnPool ConnectionPool
}

// DBWithPool returns a DB instance using the provided connection pool.
// Useful for testing or injecting a mock pool.
func DBWithPool(pool ConnectionPool) (*DB, error) {
	if pool == nil {
		return nil, e.ErrInvalidInput
	}

	return &DB{ConnPool: pool}, nil
}

// NewDB creates and initializes a new DB instance using the provided connection string.
// Returns a configured *DB or an error if the connection config is invalid or the pool fails to initialize.
func NewDB(ctx context.Context, connString string) (*DB, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("[%w] pg connection string", e.ErrParse)
	}

	// configuring pool
	config.MaxConns = pgPoolMaxSize
	config.MinConns = pgPoolMinSize
	config.MaxConnLifetime = pgMaxConnLifeTime
	config.MaxConnIdleTime = pgMaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("[%w] pg config", e.ErrInvalidInput)
	}

	return &DB{ConnPool: pool}, nil
}

// Ping checks whether the database connection pool is alive and ready to use.
// Returns ErrNotReady if the pool is nil, or ErrUnavailable if the ping fails.
func (db *DB) Ping(ctx context.Context) error {
	if db.ConnPool == nil {
		return fmt.Errorf("[%w] pg connection pool", e.ErrNotReady)
	}

	if err := db.ConnPool.Ping(ctx); err != nil {
		return fmt.Errorf("[%w] pg connection", e.ErrUnavailable)
	}

	return nil
}

// Close shuts down the underlying database connection pool if it exists.
func (db *DB) Close() {
	if db.ConnPool == nil {
		return
	}

	db.ConnPool.Close()
	db.ConnPool = nil
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation
}
