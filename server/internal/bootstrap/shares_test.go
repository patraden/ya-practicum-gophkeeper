package bootstrap_test

import (
	json "encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/dto"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/bootstrap"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestWriteSharesFile(t *testing.T) {
	t.Parallel()

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()

	t.Run("successfully writes shares to file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "shares.json")

		shares := [][]byte{
			[]byte("abc"),
			[]byte("bcd"),
			[]byte("cde"),
		}

		err := bootstrap.WriteSharesFile(shares, filePath, log)
		require.NoError(t, err)

		// Read and decode
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)

		var result dto.ShamirShares

		require.NoError(t, json.Unmarshal(data, &result))
		require.Equal(t, len(shares), len(result.Shares))

		for i, share := range shares {
			require.Equal(t, share, result.Shares[i])
		}
	})

	t.Run("fails to open file", func(t *testing.T) {
		t.Parallel()

		// Invalid path
		err := bootstrap.WriteSharesFile(nil, "/invalid/path/shares.json", log)
		require.ErrorIs(t, err, e.ErrOpen)
	})
}
