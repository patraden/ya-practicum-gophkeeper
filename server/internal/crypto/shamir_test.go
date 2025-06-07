//nolint:funlen // reason: long test functions are acceptable
package crypto_test

import (
	"testing"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/keys"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/crypto"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitterSplit(t *testing.T) {
	t.Parallel()

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()
	splitter := crypto.NewSplitter(log)

	secret := []byte("verysecretkeydata1234567890")
	secretHash := keys.HashREK(secret)
	total := 5
	threshold := 3

	t.Run("splits and returns correct number of shares", func(t *testing.T) {
		t.Parallel()

		shares, err := splitter.Split(secret, total, threshold)
		require.NoError(t, err)
		require.Len(t, shares, total)

		for _, share := range shares {
			require.NotEmpty(t, share)
		}
	})

	t.Run("fails on invalid params", func(t *testing.T) {
		t.Parallel()

		_, err := splitter.Split(secret, 2, 3) // total < threshold
		require.ErrorIs(t, err, e.ErrInvalidInput)
	})

	t.Run("combine with exactly threshold shares", func(t *testing.T) {
		t.Parallel()

		shares, err := splitter.Split(secret, total, threshold)
		require.NoError(t, err)

		partial := shares[:threshold]
		recovered, err := crypto.Combine(partial)
		require.NoError(t, err)
		require.Equal(t, secret, recovered, "secret is fully recovered")
		require.Equal(t, secretHash, keys.HashREK(recovered), "hash of the recovered secret is valid")
	})

	t.Run("combine with more than threshold shares", func(t *testing.T) {
		t.Parallel()

		shares, err := splitter.Split(secret, total, threshold)
		require.NoError(t, err)

		recovered, err := crypto.Combine(shares[:threshold+1])
		require.NoError(t, err)
		require.Equal(t, secret, recovered, "secret is fully recovered")
		require.Equal(t, secretHash, keys.HashREK(recovered), "hash of the recovered secret is valid")
	})

	t.Run("fails with too few shares", func(t *testing.T) {
		t.Parallel()

		shares, err := splitter.Split(secret, total, threshold)
		require.NoError(t, err)

		recovered, err := crypto.Combine(shares[:threshold-1])
		require.NoError(t, err)
		require.NotEqual(t, secret, recovered, "secret is not fully recovered")
		require.NotEqual(t, secretHash, keys.HashREK(recovered), "hash of the recovered secret is invalid")
	})
}

func TestCombine(t *testing.T) {
	t.Parallel()

	secret := []byte("verysecretkeydata1234567890")
	total := 5
	threshold := 3
	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()
	splitter := crypto.NewSplitter(log)

	shares, err := splitter.Split(secret, total, threshold)
	require.NoError(t, err)
	require.Len(t, shares, total)

	t.Run("success with valid shares", func(t *testing.T) {
		t.Parallel()

		recovered, err := crypto.Combine(shares[:threshold])
		require.NoError(t, err)
		require.Equal(t, secret, recovered)
	})

	t.Run("success with all shares", func(t *testing.T) {
		t.Parallel()

		recovered, err := crypto.Combine(shares)
		require.NoError(t, err)
		require.Equal(t, recovered, secret)
	})

	t.Run("error with empty shares slice", func(t *testing.T) {
		t.Parallel()

		_, err := crypto.Combine([][]byte{})
		require.ErrorIs(t, err, e.ErrInvalidInput)
	})

	t.Run("error with corrupted shares", func(t *testing.T) {
		t.Parallel()

		// deep copy.
		corruptedShares := make([][]byte, len(shares))
		for i := range shares {
			corruptedShares[i] = make([]byte, len(shares[i]))
			copy(corruptedShares[i], shares[i])
		}

		// Corrupt first share
		if len(corruptedShares[0]) > 0 {
			corruptedShares[0][0] ^= 0xFF
		}

		corruptedSecret, err := crypto.Combine(corruptedShares[:threshold])
		require.NoError(t, err)

		assert.NotEqual(t, secret, corruptedSecret)
	})
}
