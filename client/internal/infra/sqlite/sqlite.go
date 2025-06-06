package sqlite

import (
	"context"
	"database/sql"

	// Import SQLite driver anonymously to register with database/sql.
	_ "github.com/mattn/go-sqlite3"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
)

type DB struct {
	conn    *sql.DB
	Queries *Queries
}

func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, e.ErrOpen
	}

	return &DB{
		conn:    conn,
		Queries: New(conn),
	}, nil
}

func (db *DB) Ping(ctx context.Context) error {
	if err := db.conn.PingContext(ctx); err != nil {
		return e.ErrUnavailable
	}

	return nil
}

func (db *DB) Close() error {
	if err := db.conn.Close(); err != nil {
		return e.ErrClose
	}

	return nil
}
