package s3_test

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/s3"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/testutil/certtest"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestNewMinIOClientSuccess(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()
	caCertPath, serverCertPath, serverKeyPath := certtest.GenerateTestCertificates(t, tmpDir, log)
	cert, err := tls.LoadX509KeyPair(serverCertPath, serverKeyPath)

	require.NoError(t, err)

	tserver := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tserver.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	tserver.StartTLS()
	defer tserver.Close()

	endpoint := tserver.Listener.Addr().String()

	cfg := &config.Config{
		S3TLSCertPath: caCertPath,
		S3Endpoint:    endpoint,
		S3AccessKey:   "test-access",
		S3SecretKey:   "test-secret",
	}

	client, err := s3.NewMinIOClient(cfg, log)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestNewMinIOClientTransportError(t *testing.T) {
	t.Parallel()

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()

	cfg := &config.Config{
		S3TLSCertPath: "/non/existent/path.crt", // invalid
		S3Endpoint:    "localhost:9000",
		S3AccessKey:   "access",
		S3SecretKey:   "secret",
	}

	client, err := s3.NewMinIOClient(cfg, log)
	require.ErrorIs(t, err, e.ErrInvalidInput)
	require.Nil(t, client)
	require.Contains(t, err.Error(), "MinIO http transport")
}

func TestNewMinIOClientInitError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()

	certPath, _, _ := certtest.GenerateTestCertificates(t, tmpDir, log)

	cfg := &config.Config{
		S3TLSCertPath: certPath,
		S3Endpoint:    "://invalid-url",
		S3AccessKey:   "access",
		S3SecretKey:   "secret",
	}

	client, err := s3.NewMinIOClient(cfg, log)
	require.ErrorIs(t, err, e.ErrInit)
	require.Nil(t, client)
	require.Contains(t, err.Error(), "MinIO client")
}

func TestNewMinIOClientWithDocker(t *testing.T) {
	t.Skip("Disabled during development; enable when MinIO docker is running")
	t.Parallel()

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()
	cfg := &config.Config{
		S3Endpoint:    "localhost:9000",
		S3AccessKey:   "gophkeeper",
		S3SecretKey:   "gophkeeper",
		S3AccountID:   "gophkeeper",
		S3TLSCertPath: "../../../../deployments/.certs/ca.cert",
		S3Region:      "eu-central-1",
		S3RedisRegion: "eu-central-1",
	}

	client, err := s3.NewMinIOClient(cfg, log)
	require.NoError(t, err)

	require.True(t, client.IsOnline())

	ctx := context.Background()

	err = client.MakeBucket(ctx, "myfirstdbucket", map[string]string{"mytag": "success"})
	require.NoError(t, err)
}
