package model_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/patraden/ya-practicum-gophkeeper/internal/client/model"
	uavro "github.com/patraden/ya-practicum-gophkeeper/pkg/utils/avro"
	"github.com/stretchr/testify/require"
)

func TestUserCredsSerializeDeserialize(t *testing.T) {
	t.Parallel()

	schemaFile := uavro.NewSchemaFile("../../../avro/creds.avsc")

	original := &model.UserCredentials{
		UserID:         uuid.NewString(),
		Username:       "patraden",
		HashedPassword: []byte("password"),
	}

	// Marshal the UserCredentials
	data, err := original.Marshal(schemaFile)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Unmarshal it back
	creds, err := model.UnmarshalUserCreds(schemaFile, data)
	require.NoError(t, err)
	require.NotNil(t, creds)

	require.NotSame(t, original, creds, "deserialized creds should be a different object in memory")

	require.Equal(t, original.UserID, creds.UserID)
	require.Equal(t, original.Username, creds.Username)
	require.Equal(t, original.HashedPassword, creds.HashedPassword)
}
