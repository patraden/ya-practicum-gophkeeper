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
	cfg := &config.S3Config{
		Endpoint:  "http://invalid:1234",
		AccessKey: "key",
		SecretKey: "secret",
		Token:     "token",
	}

	client, err := s3.NewMinioClient(cfg, transport)
	require.ErrorIs(t, err, errors.ErrMinioClientCreate)
	require.Nil(t, client)
}
