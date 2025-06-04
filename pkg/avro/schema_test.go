package avro_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hamba/avro/v2"
	uavro "github.com/patraden/ya-practicum-gophkeeper/pkg/avro"
	"github.com/stretchr/testify/require"
)

func TestAvroSchemaFileValidSchema(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	schemaContent := `
	{
		"type": "record",
		"name": "TestSchema",
		"fields": [
			{ "name": "id", "type": "int" }
		]
	}`

	schemaPath := filepath.Join(tmpDir, "test_schema.avsc")
	err := os.WriteFile(schemaPath, []byte(schemaContent), 0o600)
	require.NoError(t, err)

	schemaFile := uavro.NewSchemaFile(schemaPath)
	schema, err := schemaFile.Read()
	require.NoError(t, err)
	require.NotNil(t, schema)
	require.Equal(t, avro.Record, schema.Type())
}

func TestAvroSchemaFileInvalidSchema(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	invalidSchema := `not a valid schema`

	schemaPath := filepath.Join(tmpDir, "bad_schema.avsc")
	err := os.WriteFile(schemaPath, []byte(invalidSchema), 0o600)
	require.NoError(t, err)

	schemaFile := uavro.NewSchemaFile(schemaPath)
	_, err = schemaFile.Read()
	require.Error(t, err)
}

func TestAvroSchemaFileNotFound(t *testing.T) {
	t.Parallel()

	schemaFile := uavro.NewSchemaFile("nonexistent.avsc")
	_, err := schemaFile.Read()
	require.Error(t, err)
	require.ErrorContains(t, err, "no such file or directory")
}
