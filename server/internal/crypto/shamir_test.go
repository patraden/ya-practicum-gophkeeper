//nolint:funlen // reason: long test functions are acceptable
package crypto_test

import (
	"testing"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/keys"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/crypto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestStdoutSplitterSplitAndDistribute(t *testing.T) {
	t.Parallel()

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()
	splitter := crypto.NewStdoutSplitter(log)

	secret := []byte("verysecretkeydata1234567890")
	secretHash := keys.HashREK(secret)
	total := 5
	threshold := 3

	t.Run("splits and returns correct number of shares", func(t *testing.T) {
		t.Parallel()

		shares, err := splitter.SplitAndDistribute(secret, total, threshold)
		require.NoError(t, err)
		require.Len(t, shares, total)

		for _, share := range shares {
			require.NotEmpty(t, share)
		}
	})

	t.Run("fails on invalid params", func(t *testing.T) {
		t.Parallel()

		_, err := splitter.SplitAndDistribute(secret, 2, 3) // total < threshold
		require.ErrorIs(t, err, e.ErrInternal)
	})

	t.Run("combine with exactly threshold shares", func(t *testing.T) {
		t.Parallel()

		shares, err := splitter.SplitAndDistribute(secret, total, threshold)
		require.NoError(t, err)

		partial := shares[:threshold]
		recovered, err := crypto.Combine(partial)
		require.NoError(t, err)
		require.Equal(t, secret, recovered, "secret is fully recovered")
		require.Equal(t, secretHash, keys.HashREK(recovered), "hash of the recovered secret is valid")
	})

	t.Run("combine with more than threshold shares", func(t *testing.T) {
		t.Parallel()

		shares, err := splitter.SplitAndDistribute(secret, total, threshold)
		require.NoError(t, err)

		recovered, err := crypto.Combine(shares[:threshold+1])
		require.NoError(t, err)
		require.Equal(t, secret, recovered, "secret is fully recovered")
		require.Equal(t, secretHash, keys.HashREK(recovered), "hash of the recovered secret is valid")
	})

	t.Run("fails with too few shares", func(t *testing.T) {
		t.Parallel()

		shares, err := splitter.SplitAndDistribute(secret, total, threshold)
		require.NoError(t, err)

		recovered, err := crypto.Combine(shares[:threshold-1])
		require.NoError(t, err)
		require.NotEqual(t, secret, recovered, "secret is not fully recovered")
		require.NotEqual(t, secretHash, keys.HashREK(recovered), "hash of the recovered secret is invalid")
	})
}
