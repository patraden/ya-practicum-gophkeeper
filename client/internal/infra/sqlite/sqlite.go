package sqlite

import (
	"context"
	"database/sql"

	// Import SQLite driver anonymously to register with database/sql.
	_ "github.com/mattn/go-sqlite3"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/rs/zerolog"
)

type DB struct {
	conn    *sql.DB
	Queries *Queries
	log     *zerolog.Logger
}

func NewDB(dbPath string, log *zerolog.Logger) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Error().Err(err).
			Str("db_path", dbPath).
			Msg("failed to open db")

		return nil, errors.ErrDBInit
	}

	return &DB{
		conn:    conn,
		Queries: New(conn),
		log:     log,
	}, nil
}

func (db *DB) Ping(ctx context.Context) error {
	if err := db.conn.PingContext(ctx); err != nil {
		db.log.Error().Err(err).
			Msg("failed to ping sqlite db")

		return errors.ErrDBConn
	}

	return nil
}

func (db *DB) Close() error {
	if err := db.conn.Close(); err != nil {
		db.log.Info().Msg("closed sqlite db")

		return errors.ErrDBClose
	}

	return nil
}
