package sqlite_test

import (
	"path/filepath"
	"testing"

	"github.com/patraden/ya-practicum-gophkeeper/internal/domain/errors"
	"github.com/patraden/ya-practicum-gophkeeper/internal/logger"
	"github.com/patraden/ya-practicum-gophkeeper/internal/storage/sqlite"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func makeTestDB(t *testing.T) *sqlite.DB {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()

	db, err := sqlite.NewDB(dbPath, log)
	require.NoError(t, err)

	return db
}

func TestDBInitAndPing(t *testing.T) {
	t.Parallel()

	db := makeTestDB(t)

	err := db.Ping(t.Context())
	require.NoError(t, err)

	db.Close()

	err = db.Ping(t.Context())
	require.ErrorIs(t, err, errors.ErrDBConn)
}
