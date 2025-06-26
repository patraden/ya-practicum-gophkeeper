package secret_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/secret"
	"github.com/stretchr/testify/require"
)

func TestNewMeta(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	secretID := uuid.New()
	data := secret.MetaData{
		"env":   "prod",
		"owner": "alice",
	}

	meta := secret.NewMeta(userID, secretID, data)

	require.Equal(t, userID, meta.UserID)
	require.Equal(t, secretID, meta.SecretID)
	require.Equal(t, data, meta.Data)
	require.WithinDuration(t, time.Now().UTC(), meta.CreatedAt, time.Second)
	require.WithinDuration(t, meta.CreatedAt, meta.UpdatedAt, time.Millisecond)
}

func TestMetaDataJSON(t *testing.T) {
	t.Parallel()

	input := secret.MetaData{
		"region": "eu",
		"role":   "storage",
	}

	data, err := input.MarshalJSON()
	require.NoError(t, err)

	var output secret.MetaData
	err = output.UnmarshalJSON(data)
	require.NoError(t, err)

	require.Equal(t, input, output)
}

func TestMetaJSONRoundTrip(t *testing.T) {
	t.Parallel()

	meta := secret.NewMeta(uuid.New(), uuid.New(), secret.MetaData{
		"type":    "binary",
		"expires": "2025-12-31",
	})

	data, err := meta.MarshalJSON()
	require.NoError(t, err)

	var restored secret.Meta
	err = restored.UnmarshalJSON(data)
	require.NoError(t, err)

	require.Equal(t, meta.UserID, restored.UserID)
	require.Equal(t, meta.SecretID, restored.SecretID)
	require.Equal(t, meta.Data, restored.Data)
	require.WithinDuration(t, meta.CreatedAt, restored.CreatedAt, time.Second)
	require.WithinDuration(t, meta.UpdatedAt, restored.UpdatedAt, time.Second)
}
