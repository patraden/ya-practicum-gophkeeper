package user_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser(t *testing.T) {
	t.Parallel()

	t.Run("SetPassword and CheckPassword with correct and wrong input", func(t *testing.T) {
		t.Parallel()

		usr := user.New("denis", user.RoleUser)

		err := usr.SetPassword("strongpass123")
		require.NoError(t, err)

		assert.NotEmpty(t, usr.Password, "Password hash should be set")
		assert.NotEmpty(t, usr.Salt, "Salt should be set")
		assert.NotEmpty(t, usr.Verifier, "Verifier should be set")

		assert.True(t, usr.CheckPassword("strongpass123"), "Password should match")
		assert.False(t, usr.CheckPassword("wrongpassword"), "Password should not match")
	})

	t.Run("CheckVerifier with correct and incorrect verifier", func(t *testing.T) {
		t.Parallel()

		usr := user.New("verifier", user.RoleUser)
		err := usr.SetPassword("verifypass")
		require.NoError(t, err)

		assert.True(t, usr.CheckVerifier(usr.Verifier), "Correct verifier should match")
		assert.False(t, usr.CheckVerifier([]byte("invalid")), "Incorrect verifier should not match")
	})

	t.Run("NewWithID valid and invalid UUIDs", func(t *testing.T) {
		t.Parallel()

		validID := uuid.New().String()
		usr, err := user.NewWithID(validID, "denis", user.RoleUser)

		require.NoError(t, err)
		assert.Equal(t, validID, usr.ID.String(), "Should assign correct ID")

		_, err = user.NewWithID("invalid-uuid", "denis", user.RoleUser)
		require.Error(t, err, "Should return error on invalid UUID")
	})
}
