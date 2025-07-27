package shamir_test

import (
	"testing"

	sham "github.com/hashicorp/vault/shamir"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/keys"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/crypto/shamir"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestSplitter(t *testing.T) {
	t.Parallel()

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()
	splitter := shamir.NewSplitter(log)

	secret := []byte("verysecretkeydata1234567890")
	secretHash := keys.HashREK(secret)

	t.Run("splits and returns correct number of shares", func(t *testing.T) {
		t.Parallel()

		shares, err := splitter.Split(secret)
		require.NoError(t, err)
		require.Len(t, shares, shamir.TotalShares)

		for _, share := range shares {
			require.NotEmpty(t, share)
		}
	})

	t.Run("combine with exactly threshold shares", func(t *testing.T) {
		t.Parallel()

		shares, err := splitter.Split(secret)
		require.NoError(t, err)

		partial := shares[:shamir.ThresholdShares]
		recovered, err := sham.Combine(partial)
		require.NoError(t, err)
		require.Equal(t, secret, recovered, "secret is fully recovered")
		require.Equal(t, secretHash, keys.HashREK(recovered), "hash of the recovered secret is valid")
	})

	t.Run("combine with more than threshold shares", func(t *testing.T) {
		t.Parallel()

		shares, err := splitter.Split(secret)
		require.NoError(t, err)

		recovered, err := sham.Combine(shares[:shamir.ThresholdShares+1])
		require.NoError(t, err)
		require.Equal(t, secret, recovered, "secret is fully recovered")
		require.Equal(t, secretHash, keys.HashREK(recovered), "hash of the recovered secret is valid")
	})

	t.Run("fails with too few shares", func(t *testing.T) {
		t.Parallel()

		shares, err := splitter.Split(secret)
		require.NoError(t, err)

		recovered, err := sham.Combine(shares[:shamir.ThresholdShares-1])
		require.NoError(t, err)
		require.NotEqual(t, secret, recovered, "secret is not fully recovered")
		require.NotEqual(t, secretHash, keys.HashREK(recovered), "hash of the recovered secret is invalid")
	})
}
