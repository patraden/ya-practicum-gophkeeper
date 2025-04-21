package model

import (
	"github.com/hamba/avro/v2"
	uavro "github.com/patraden/ya-practicum-gophkeeper/pkg/utils/avro"
)

const cardNumberLength = 16

// UnmarshalBankCard deserializes BankCard from avro binary.
func UnmarshalBankCard(schemaFile *uavro.SchemaFile, val []byte) (*BankCard, error) {
	card := &BankCard{}

	schema, err := schemaFile.Read()
	if err != nil {
		return nil, err
	}

	if err := avro.Unmarshal(schema, val, card); err != nil {
		return nil, err
	}

	return card, nil
}

// IsValid validates BankCard data consistency.
func (b *BankCard) IsValid() bool {
	// optionally can add here luhn algo
	// https://github.com/ShiraazMoollatjie/goluhn
	return len(b.CardNumber) == cardNumberLength
}

// Marshal serializes BankCard to avro binary.
func (b *BankCard) Marshal(schemaFile *uavro.SchemaFile) ([]byte, error) {
	schema, err := schemaFile.Read()
	if err != nil {
		return []byte{}, err
	}

	return avro.Marshal(schema, b)
}
