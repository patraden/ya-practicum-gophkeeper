package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	// Import SQLite driver anonymously to register with database/sql.
	"github.com/mattn/go-sqlite3"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
)

// QueryFunc defines a function that accepts a pg-generated Queries instance
// and performs a set of SQL operations.
type QueryFunc = func(*Queries) error

type DB struct {
	Conn    *sql.DB
	Queries *Queries
}

func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("[%w] sqlite3 db", e.ErrOpen)
	}

	return &DB{
		Conn:    conn,
		Queries: New(conn),
	}, nil
}

func (db *DB) Ping(ctx context.Context) error {
	if err := db.Conn.PingContext(ctx); err != nil {
		return fmt.Errorf("[%w] sqlite3 db", e.ErrUnavailable)
	}

	return nil
}

func (db *DB) Close() error {
	if err := db.Conn.Close(); err != nil {
		return fmt.Errorf("[%w] sqlite3 db", e.ErrClose)
	}

	return nil
}

func IsUniqueViolation(err error) bool {
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		// Code 2067 = constraint failed; unique index/constraint
		return sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique
	}
	// Fallback for wrapped string error
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
