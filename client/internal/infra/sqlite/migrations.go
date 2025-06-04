package sqlite

import (
	"embed"

	"github.com/patraden/ya-practicum-gophkeeper/client/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/migrations"
)

const migrationsDir = "migrations"

//go:embed migrations/*.sql
var embedMigrations embed.FS

// RunClientMigrations applies all SQLite migrations embedded in the binary.
func RunClientMigrations(config *config.Config, logger *logger.Logger) error {
	return migrations.RunSQLite(config.DatabaseDSN, embedMigrations, migrationsDir, logger)
}
