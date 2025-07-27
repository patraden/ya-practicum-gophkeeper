//nolint:funlen // reason: long test functions are acceptable
package keys_test

import (
	"testing"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/keys"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestGenerateKeys(t *testing.T) {
	t.Parallel()

	t.Run("generated keys have correct length", func(t *testing.T) {
		t.Parallel()

		key, err := keys.REK()
		require.NoError(t, err)
		require.Len(t, key, keys.REKLength)

		usr := user.New("test_user", user.RoleUser)
		err = usr.SetPassword("user_password")
		require.NoError(t, err)

		key, err = keys.KEK(usr, "user_password")
		require.NoError(t, err)
		require.Len(t, key, keys.REKLength)

		key, err = keys.DEK()
		require.NoError(t, err)
		require.Len(t, key, keys.REKLength)
	})
}

func TestWrapUnwrap(t *testing.T) {
	t.Parallel()
	t.Run("wrap and unwrap DEK with KEK", func(t *testing.T) {
		t.Parallel()

		dek, err := keys.DEK()
		require.NoError(t, err)

		usr := user.New("test_user", user.RoleUser)
		err = usr.SetPassword("user_password")
		require.NoError(t, err)

		kek, err := keys.KEK(usr, "user_password")
		require.NoError(t, err)

		wrapped, err := keys.WrapDEK(kek, dek)
		require.NoError(t, err)
		require.NotNil(t, wrapped)
		require.Greater(t, len(wrapped), keys.DEKLength)

		unwrapped, err := keys.UnwrapDEK(kek, wrapped)
		require.NoError(t, err)
		require.Equal(t, dek, unwrapped)
	})

	t.Run("fail to wrap with invalid key length", func(t *testing.T) {
		t.Parallel()

		shortKey := make([]byte, 10)
		dek, err := keys.DEK()
		require.NoError(t, err)

		_, err = keys.WrapDEK(shortKey, dek)
		require.ErrorIs(t, e.ErrInvalidInput, err)
	})

	t.Run("fail to unwrap with invalid key length", func(t *testing.T) {
		t.Parallel()

		shortKey := make([]byte, 10)
		kek, err := keys.DEK() // just reuse a valid-length key for ciphertext
		require.NoError(t, err)

		wrapped, err := keys.WrapDEK(kek, kek)
		require.NoError(t, err)

		_, err = keys.UnwrapDEK(shortKey, wrapped)
		require.ErrorIs(t, e.ErrInvalidInput, err)
	})

	t.Run("fail to unwrap with tampered ciphertext", func(t *testing.T) {
		t.Parallel()

		dek, err := keys.DEK()
		require.NoError(t, err)
		kek, err := keys.DEK()
		require.NoError(t, err)
		wrapped, err := keys.WrapDEK(kek, dek)
		require.NoError(t, err)

		// Flip a bit in the ciphertext (after nonce)
		wrapped[len(wrapped)-1] ^= 0xFF

		_, err = keys.UnwrapDEK(kek, wrapped)
		require.Error(t, err)
	})

	t.Run("fail to unwrap with incorrect KEK", func(t *testing.T) {
		t.Parallel()

		dek, err := keys.DEK()
		require.NoError(t, err)
		kek, err := keys.DEK()
		require.NoError(t, err)
		otherKEK, err := keys.DEK()
		require.NoError(t, err)

		wrapped, err := keys.WrapDEK(kek, dek)
		require.NoError(t, err)

		_, err = keys.UnwrapDEK(otherKEK, wrapped)
		require.ErrorIs(t, err, e.ErrDecrypt)
	})

	t.Run("fail to unwrap with too short wrapped data", func(t *testing.T) {
		t.Parallel()

		kek, err := keys.DEK()
		require.NoError(t, err)

		wrapped := make([]byte, keys.DEKLength) // less than nonceSize

		_, err = keys.UnwrapDEK(kek, wrapped[:5]) // shorter than nonce
		require.ErrorIs(t, e.ErrInvalidInput, err)
	})
}
