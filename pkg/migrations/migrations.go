package migrations

import (
	"database/sql"
	"embed"

	// Import pgx driver for SQL compatibility.
	_ "github.com/jackc/pgx/v5/stdlib"
	// Import SQLite driver anonymously to register with database/sql.
	_ "github.com/mattn/go-sqlite3"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/pressly/goose/v3"
)

const (
	DriverSQLite   = "sqlite3"
	DriverPostgres = "pgx"
)

// RunSQLite connects to a SQLite database using the provided DSN and applies
// embedded migration scripts from the given directory using the Goose library.
// Returns a standardized error if migration fails at any stage.
func RunSQLite(dsn string, embedFS embed.FS, dir string, logger *logger.Logger) error {
	db, err := sql.Open(DriverSQLite, dsn)
	if err != nil {
		logger.GetZeroLog().Error().Err(err).
			Str("driver_name", DriverSQLite).
			Msg("Failed to open db connection")

		return e.ErrOpen
	}

	return Run(db, embedFS, string(goose.DialectSQLite3), dir, logger)
}

// RunPG connects to a PostgreSQL database using the provided DSN and applies
// embedded migration scripts from the given directory using the Goose library.
// Returns a standardized error if migration fails at any stage.
func RunPG(dsn string, embedFS embed.FS, dir string, logger *logger.Logger) error {
	db, err := sql.Open(DriverPostgres, dsn)
	if err != nil {
		logger.GetZeroLog().Error().Err(err).
			Str("driver_name", DriverPostgres).
			Msg("Failed to open db connection")

		return e.ErrOpen
	}

	return Run(db, embedFS, string(goose.DialectPostgres), dir, logger)
}

// Run executes database migrations using the provided database connection,
// embedded filesystem, migration dialect, and migration directory.
// It uses Goose to apply all available migrations and logs progress and errors.
// Returns a standardized migration error if any operation fails.
func Run(db *sql.DB, embedFS embed.FS, dialect, dir string, logger *logger.Logger) error {
	defer db.Close()

	log := logger.GetZeroLog()

	if err := db.Ping(); err != nil {
		logger.GetZeroLog().Error().Err(err).
			Msg("DB is unreachable")

		return e.ErrUnavailable
	}

	goose.SetBaseFS(embedFS)
	goose.SetLogger(logger)

	if err := goose.SetDialect(dialect); err != nil {
		log.Error().Err(err).
			Str("dialect", dialect).
			Msg("Failed to set db dialect")

		return e.ErrInvalidInput
	}

	if err := goose.Up(db, dir); err != nil {
		log.Error().Err(err).
			Str("dialect", dialect).
			Msg("Failed to apply migrations")

		return e.InternalErr(err)
	}

	log.Info().
		Str("dialect", dialect).
		Str("dir", dir).
		Msg("Migrations applied successfully")

	return nil
}
