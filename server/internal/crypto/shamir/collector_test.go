//nolint:funlen // reason: long test functions are acceptable
package shamir_test

import (
	"slices"
	"testing"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/keys"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/crypto/shamir"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func cloneShares(t *testing.T, shares [][]byte) [][]byte {
	t.Helper()

	cloned := make([][]byte, len(shares))
	for i, share := range shares {
		cloned[i] = slices.Clone(share)
	}

	return cloned
}

//nolint:varnamelen // reason: allow short var for collector.
func TestCollector(t *testing.T) {
	t.Parallel()

	secret := []byte("ultrasecret-shamir-data")
	secretHash := keys.HashREK(secret)

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()
	splitter := shamir.NewSplitter(log)
	originalShares, err := splitter.Split(secret)

	require.NoError(t, err)
	require.Len(t, originalShares, shamir.TotalShares)

	t.Run("collects shares up to threshold", func(t *testing.T) {
		t.Parallel()

		c := shamir.NewCollector(log)
		shares := cloneShares(t, originalShares)

		for i := range shamir.ThresholdShares {
			require.NoError(t, c.Collect(shares[i]))
		}

		require.True(t, c.IsThresholdMet())
		require.Equal(t, shamir.ThresholdShares, c.Size())

		log.Info().Str("message", c.StatusMessage()).Msg("collector status")
	})

	t.Run("duplicate shares are ignored", func(t *testing.T) {
		t.Parallel()

		c := shamir.NewCollector(log)
		shares := cloneShares(t, originalShares)

		share0 := shares[0]
		share1 := shares[1]
		share2 := shares[2]
		share3 := shares[3]
		share4 := shares[4]
		share1dup := slices.Clone(shares[1]) // fresh copy again

		require.NoError(t, c.Collect(share0))
		require.NoError(t, c.Collect(share1))
		require.NoError(t, c.Collect(share2))
		require.NoError(t, c.Collect(share3))
		require.NoError(t, c.Collect(share1dup)) // will be deduplicated correctly
		require.NoError(t, c.Collect(share4))

		require.True(t, c.IsThresholdMet())
		require.Equal(t, 5, c.Size())
	})

	t.Run("returns error if collecting beyond threshold", func(t *testing.T) {
		t.Parallel()

		c := shamir.NewCollector(log)
		shares := cloneShares(t, originalShares)

		for i := range shamir.ThresholdShares {
			require.NoError(t, c.Collect(shares[i]))
		}

		err := c.Collect(shares[shamir.ThresholdShares]) // one extra
		require.ErrorIs(t, err, e.ErrConflict)
	})

	t.Run("reconstructs secret correctly", func(t *testing.T) {
		t.Parallel()

		c := shamir.NewCollector(log)
		shares := cloneShares(t, originalShares)

		for i := range shamir.ThresholdShares {
			require.NoError(t, c.Collect(shares[i]))
		}

		rek, err := c.Reconstruct()
		require.NoError(t, err)

		require.Equal(t, secret, rek)
		require.Equal(t, secretHash, keys.HashREK(rek))
	})

	t.Run("reconstruct fails if threshold not met", func(t *testing.T) {
		t.Parallel()

		c := shamir.NewCollector(log)
		shares := cloneShares(t, originalShares)

		require.NoError(t, c.Collect(shares[0]))

		_, err := c.Reconstruct()
		require.ErrorIs(t, err, e.ErrNotReady)
	})

	t.Run("reset securely clears all collected shares", func(t *testing.T) {
		t.Parallel()

		c := shamir.NewCollector(log)
		shares := cloneShares(t, originalShares)

		for i := range shamir.ThresholdShares {
			require.NoError(t, c.Collect(shares[i]))
		}

		require.True(t, c.IsThresholdMet())
		c.Reset()

		require.False(t, c.IsThresholdMet())
		require.Equal(t, 0, c.Size())

		// Further reconstruct fails
		_, err := c.Reconstruct()
		require.ErrorIs(t, err, e.ErrNotReady)
	})
}
