package keystore_test

import (
	"testing"

	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/crypto/keystore"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest // reason: key store test must be run consequently
func TestKeyStore(t *testing.T) {
	t.Parallel()

	kstore := keystore.NewInMemoryKeystore()

	t.Run("Uninitialized returns error", func(t *testing.T) {
		kstore.Wipe()

		_, err := kstore.Get()
		require.ErrorIs(t, err, e.ErrNotReady)
	})

	t.Run("Initialize stores REK securely", func(t *testing.T) {
		kstore.Wipe()

		original := []byte("super-secret-key")
		key := make([]byte, len(original))
		copy(key, original)

		err := kstore.Load(key)
		require.NoError(t, err)

		out, err := kstore.Get()
		require.NoError(t, err)
		require.True(t, kstore.IsLoaded())
		require.Equal(t, original, out)
	})

	t.Run("Reinitialization has no effect", func(t *testing.T) {
		kstore.Wipe()

		key1 := []byte("first-key")
		key2 := []byte("second-key")
		key1Copy := make([]byte, len(key1))
		copy(key1Copy, key1)

		err := kstore.Load(key1)
		require.NoError(t, err)

		err = kstore.Load(key2)
		require.ErrorIs(t, err, e.ErrConflict)

		out, err := kstore.Get()
		require.NoError(t, err)
		require.True(t, kstore.IsLoaded())
		require.Equal(t, key1Copy, out)
	})

	t.Run("Wipe clears the key", func(t *testing.T) {
		kstore.Wipe()

		key := []byte("to-be-wiped")
		err := kstore.Load(key)
		require.NoError(t, err)
		kstore.Wipe()

		_, err = kstore.Get()
		require.ErrorIs(t, err, e.ErrNotReady)
	})
}
