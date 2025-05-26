package s3_test

import (
	"net/http"
	"testing"

	"github.com/patraden/ya-practicum-gophkeeper/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/internal/domain/errors"
	"github.com/patraden/ya-practicum-gophkeeper/internal/storage/s3"
	"github.com/stretchr/testify/require"
)

func TestNewMinioClient_Invalid(t *testing.T) {
	t.Parallel()

	transport := &http.Transport{}
	cfg := &config.Config{
		S3Endpoint:  "http://invalid:1234",
		S3AccessKey: "key",
		S3SecretKey: "secret",
		S3Token:     "token",
	}

	client, err := s3.NewMinioClient(cfg, transport)
	require.ErrorIs(t, err, errors.ErrMinioClientCreate)
	require.Nil(t, client)
}
