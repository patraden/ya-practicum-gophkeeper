package migrations_test

import (
	"embed"
	"path/filepath"
	"testing"

	"github.com/patraden/ya-practicum-gophkeeper/internal/domain/errors"
	"github.com/patraden/ya-practicum-gophkeeper/internal/logger"
	"github.com/patraden/ya-practicum-gophkeeper/internal/storage/migrations"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/sqlite/*.sql
var testMigrations embed.FS

func makeTempDB(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	return dbPath
}

func TestRunSQLite(t *testing.T) {
	t.Parallel()

	log := logger.Stdout(zerolog.DebugLevel)

	validDSN := makeTempDB(t)

	tests := []struct {
		name    string
		dsn     string
		dir     string
		fs      embed.FS
		wantErr error
	}{
		{
			name:    "valid migration",
			dsn:     validDSN,
			dir:     "testdata/sqlite",
			fs:      testMigrations,
			wantErr: nil,
		},
		{
			name:    "invalid migration dir",
			dsn:     validDSN,
			dir:     "testdata/unknown",
			fs:      testMigrations,
			wantErr: errors.ErrDBMigration,
		},
		{
			name:    "invalid dsn path",
			dsn:     "/nonexistingpath/test.db",
			dir:     "testdata/sqlite",
			fs:      testMigrations,
			wantErr: errors.ErrDBMigration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := migrations.RunSQLite(tt.dsn, tt.fs, tt.dir, log)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
