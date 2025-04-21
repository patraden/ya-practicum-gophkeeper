package model

import (
	"github.com/hamba/avro/v2"
	uavro "github.com/patraden/ya-practicum-gophkeeper/pkg/utils/avro"
)

// UnmarshalUserCreds deserializes UserCredentials from avro binary.
func UnmarshalUserCreds(schemaFile *uavro.SchemaFile, val []byte) (*UserCredentials, error) {
	creds := &UserCredentials{}

	schema, err := schemaFile.Read()
	if err != nil {
		return nil, err
	}

	if err := avro.Unmarshal(schema, val, creds); err != nil {
		return nil, err
	}

	return creds, nil
}

// Marshal serializes UserCredentials to avro binary.
func (c *UserCredentials) Marshal(schemaFile *uavro.SchemaFile) ([]byte, error) {
	schema, err := schemaFile.Read()
	if err != nil {
		return []byte{}, err
	}

	return avro.Marshal(schema, c)
}
