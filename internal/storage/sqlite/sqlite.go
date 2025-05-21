package sqlite

import (
	"context"
	"database/sql"
	"embed"

	// Import SQLite driver anonymously to register with database/sql.
	_ "github.com/mattn/go-sqlite3"
	"github.com/patraden/ya-practicum-gophkeeper/internal/domain/errors"
	"github.com/patraden/ya-practicum-gophkeeper/internal/logger"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

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

		return nil, errors.ErrSQLiteInit
	}

	return &DB{
		conn:    conn,
		Queries: New(conn),
		log:     log,
	}, nil
}

func (db *DB) Migrate() error {
	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(logger.Stdout(zerolog.DebugLevel))

	if err := goose.SetDialect("sqlite3"); err != nil {
		db.log.Error().Err(err).
			Msg("failed to set migrations dialect")

		return errors.ErrSQLiteInit
	}

	if err := goose.Up(db.conn, "migrations"); err != nil {
		db.log.Error().Err(err).
			Msg("failed to apply migrations")

		return errors.ErrSQLiteInit
	}

	return nil
}

func (db *DB) Ping(ctx context.Context) error {
	if err := db.conn.PingContext(ctx); err != nil {
		db.log.Error().Err(err).
			Msg("failed to ping sqlite db")

		return errors.ErrSQLiteConn
	}

	return nil
}

func (db *DB) Close() error {
	if err := db.conn.Close(); err != nil {
		db.log.Info().Msg("closed sqlite db")

		return errors.ErrSQLiteClose
	}

	return nil
}
