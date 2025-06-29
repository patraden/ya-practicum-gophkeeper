package sqlite

import (
	"embed"
	"path"

	"github.com/patraden/ya-practicum-gophkeeper/client/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/migrations"
)

const migrationsDir = "migrations"

//go:embed migrations/*.sql
var embedMigrations embed.FS

// RunClientMigrations applies all SQLite migrations embedded in the binary.
func RunClientMigrations(config *config.Config, logger logger.Logger) error {
	dsn := path.Join(config.InstallDir, config.DatabaseFileName)
	return migrations.RunSQLite(dsn, embedMigrations, migrationsDir, logger)
}
