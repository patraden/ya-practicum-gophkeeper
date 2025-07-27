package pg

import (
	"embed"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/migrations"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
)

const migrationsDir = "migrations"

//go:embed migrations/*.sql
var embedMigrations embed.FS

// RunServerMigrations applies all PostgreSQL migrations embedded in the binary.
// This is intended for server-side migrations.
func RunServerMigrations(config *config.Config, logger logger.Logger) error {
	return migrations.RunPG(config.DatabaseDSN, embedMigrations, migrationsDir, logger)
}
