package repostory_test

import (
	"context"
	"testing"

	"github.com/patraden/ya-practicum-gophkeeper/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/internal/logger"
	repo "github.com/patraden/ya-practicum-gophkeeper/internal/repository"
	"github.com/patraden/ya-practicum-gophkeeper/internal/storage/s3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validCert = `-----BEGIN CERTIFICATE-----
MIIB5jCCAYygAwIBAgIRAPIBNlr5AgVg7LmwwHZ5AZowCgYIKoZIzj0EAwIwOjEc
MBoGA1UEChMTQ2VydGdlbiBEZXZlbG9wbWVudDEaMBgGA1UECwwRcm9vdEAxNjVl
OGYxOTM5OGEwHhcNMjUwNDEzMjAwMjExWhcNMjYwNDEzMjAwMjExWjA6MRwwGgYD
VQQKExNDZXJ0Z2VuIERldmVsb3BtZW50MRowGAYDVQQLDBFyb290QDE2NWU4ZjE5
Mzk4YTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABCLCGjyg3/DUExQ7byXOdF1W
vPyjzXlfAV5OwEJdbv8MZTtTz4rsJEZeUU/HSV2OpLYfs/R8j6pKYD3zcp5djiWj
czBxMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggrBgEFBQcDATAPBgNVHRMB
Af8EBTADAQH/MB0GA1UdDgQWBBSIRoY2U0W5CZQxMWBnYYk32+rGYzAaBgNVHREE
EzARgglsb2NhbGhvc3SHBH8AAAEwCgYIKoZIzj0EAwIDSAAwRQIgdMipjU+5qR2i
EmgAFSzd0YmFKzIA3VvMKcsLWTPmQ3sCIQCieiGaNkihZq9PFotCp6lVCC+F/Dfs
0bVOhAWuQTRKog==
-----END CERTIFICATE-----`

func TestMinioDevIntegration(t *testing.T) {
	t.Skip("Disabled during development; enable when MinIO is available")
	t.Parallel()

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()
	ctx := context.Background()
	cfg := config.DefaultConfig()
	transportBuilder := s3.NewHTTPTransportBuilder("", []byte(validCert), log)

	transport, err := transportBuilder.Build()
	require.NoError(t, err)

	client, err := s3.NewMinioClient(cfg, transport)
	require.NoError(t, err)

	assert.True(t, client.IsOnline())

	storage := repo.NewS3Storage(client, log)

	err = storage.MakeBucket(ctx, "myfirstbucket")
	require.NoError(t, err)
}

// func TestMinioClient(t *testing.T) {

// 	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()
// 	ctx := context.Background()
// 	cfg := config.DefaultObjectStore()
// 	transportBuilder := s3.NewHTTPTransportBuilder("", []byte(validCert), log)

// 	transport, err := transportBuilder.Build()
// 	require.NoError(t, err)

// 	client, err := s3.NewMinioClient(cfg, transport)
// 	require.NoError(t, err)

// 	client.PutObject(minio.PutObjectOptions)

// }
