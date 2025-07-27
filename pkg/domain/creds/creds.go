package creds

import (
	"github.com/hamba/avro/v2"
	uavro "github.com/patraden/ya-practicum-gophkeeper/pkg/avro"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
)

// UnmarshalUserCreds deserializes UserCredentials from avro binary.
func UnmarshalUserCreds(schemaFile *uavro.SchemaFile, val []byte) (*UserCredentials, error) {
	creds := &UserCredentials{}

	schema, err := schemaFile.Read()
	if err != nil {
		return nil, err
	}

	if err := avro.Unmarshal(schema, val, creds); err != nil {
		return nil, e.ErrUnmarshal
	}

	return creds, nil
}

// Marshal serializes UserCredentials to avro binary.
func (c *UserCredentials) Marshal(schemaFile *uavro.SchemaFile) ([]byte, error) {
	schema, err := schemaFile.Read()
	if err != nil {
		return []byte{}, e.ErrRead
	}

	avro, err := avro.Marshal(schema, c)
	if err != nil {
		return []byte{}, e.ErrMarshal
	}

	return avro, nil
}
