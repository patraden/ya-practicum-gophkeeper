package card_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/patraden/ya-practicum-gophkeeper/internal/domain/card"
	uavro "github.com/patraden/ya-practicum-gophkeeper/pkg/utils/avro"
	"github.com/stretchr/testify/require"
)

func TestBankCardSerializeDeserialize(t *testing.T) {
	t.Parallel()

	schemaFile := uavro.NewSchemaFile("../../../avro/card.avsc")

	original := &card.BankCard{
		UserID:         uuid.NewString(),
		CardholderName: "Denis Patrakhin",
		CardNumber:     "1234567812345678",
		ExpiryMonth:    12,
		ExpiryYear:     2029,
		Cvv:            123,
	}

	require.True(t, original.IsValid())

	// Marshal the BankCard
	data, err := original.Marshal(schemaFile)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Unmarshal it back
	card, err := card.UnmarshalBankCard(schemaFile, data)
	require.NoError(t, err)
	require.NotNil(t, card)

	require.NotSame(t, original, card, "deserialized card should be a different object in memory")

	require.Equal(t, original.UserID, card.UserID)
	require.Equal(t, original.CardholderName, card.CardholderName)
	require.Equal(t, original.CardNumber, card.CardNumber)
	require.Equal(t, original.ExpiryMonth, card.ExpiryMonth)
	require.Equal(t, original.ExpiryYear, card.ExpiryYear)
	require.Equal(t, original.Cvv, card.Cvv)
}
