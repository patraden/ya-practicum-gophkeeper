package sqlite_test

import (
	"path/filepath"
	"testing"

	"github.com/patraden/ya-practicum-gophkeeper/client/internal/infra/sqlite"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/stretchr/testify/require"
)

func makeTestDB(t *testing.T) *sqlite.DB {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sqlite.NewDB(dbPath)
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
	require.ErrorIs(t, err, errors.ErrUnavailable)
}
