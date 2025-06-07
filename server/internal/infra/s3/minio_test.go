package s3_test

import (
	"net/http"
	"testing"

	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/s3"
	"github.com/stretchr/testify/require"
)

func TestNewMinioClientInvalid(t *testing.T) {
	t.Parallel()

	transport := &http.Transport{}
	cfg := &config.Config{
		S3Endpoint:  "http://invalid:1234",
		S3AccessKey: "key",
		S3SecretKey: "secret",
		S3Token:     "token",
	}

	client, err := s3.NewMinioClient(cfg, transport)
	require.ErrorIs(t, err, e.ErrInternal)
	require.Nil(t, client)
}
