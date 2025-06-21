package minio_test

import (
	"context"
	"testing"
	"time"

	"github.com/patraden/ya-practicum-gophkeeper/client/internal/infra/minio"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/s3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestClientIntergationPutObject(t *testing.T) {
	t.Parallel()
	t.Skip("Enable when dev environment (pod and containers) are up and running")

	// create test file, macOS:
	// mkfile 5g /Users/d.patrakhin/Downloads/testfile_5gb.bin
	// mkfile 100m /Users/d.patrakhin/Downloads/testfile_100m.bin
	// get s3 temp creds:
	// go test -v ./server/internal/infra/minio/... -run ^TestIntegrationClientTokenProvisioning

	bucketName := "testbucket"
	log := logger.StdoutConsole(zerolog.DebugLevel).GetZeroLog()
	cfg := &s3.ClientConfig{
		S3Endpoint:    "localhost:9000",
		S3TLSCertPath: "../../../../deployments/.certs/ca.cert",
		S3AccessKey:   "",
		S3SecretKey:   "",
		S3Token:       "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	uploadOps := s3.PutObjectOptions{
		SendContentMd5: true,
	}
	filePath := "/Users/d.patrakhin/Downloads/testfile_100m.bin"
	objectName := "testfile_100m.bin"

	client, err := minio.NewClient(cfg, log)
	require.NoError(t, err)

	info, err := client.PutObject(ctx, bucketName, objectName, filePath, uploadOps)
	require.NoError(t, err)
	require.NotNil(t, info)
}

func TestClientIntergationGetObject(t *testing.T) {
	t.Parallel()
	t.Skip("Enable when dev environment (pod and containers) are up and running")

	bucketName := "testbucket"
	log := logger.StdoutConsole(zerolog.DebugLevel).GetZeroLog()
	cfg := &s3.ClientConfig{
		S3Endpoint:    "localhost:9000",
		S3TLSCertPath: "../../../../deployments/.certs/ca.cert",
		S3AccessKey:   "",
		S3SecretKey:   "",
		S3Token:       "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	downloadOps := s3.GetObjectOptions{}
	objectName := "testfile_5gb.bin"
	filePath := "/Users/d.patrakhin/Downloads/testfile_5gb.bin.copy"

	client, err := minio.NewClient(cfg, log)
	require.NoError(t, err)

	err = client.GetObject(ctx, bucketName, objectName, filePath, downloadOps)
	require.NoError(t, err)
}
