package avro

import (
	"os"
	"sync"

	"github.com/hamba/avro/v2"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
)

type SchemaFile struct {
	schema   avro.Schema
	err      error
	filePath string
	once     sync.Once
}

func NewSchemaFile(filePath string) *SchemaFile {
	return &SchemaFile{
		schema:   nil,
		err:      nil,
		filePath: filePath,
		once:     sync.Once{},
	}
}

func (s *SchemaFile) Read() (avro.Schema, error) {
	s.once.Do(func() {
		s.schema, s.err = loadSchema(s.filePath)
	})

	return s.schema, s.err
}

func loadSchema(filePath string) (avro.Schema, error) {
	schemaData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, e.ErrRead
	}

	schema, err := avro.Parse(string(schemaData))
	if err != nil {
		return nil, e.ErrParse
	}

	return schema, nil
}
