package minio

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/net/transport"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/s3"
	"github.com/rs/zerolog"
)

// MinIOClient wraps the minio.Client and provides additional configuration
// and logging capabilities for interacting with an S3-compatible MinIO backend.
type Client struct {
	s3.ClientOperator
	minio *minio.Client
	cfg   *s3.ClientConfig
	log   *zerolog.Logger
}

// NewClient initializes a new MinIO S3 client using the provided configuration.
func NewClient(cfg *s3.ClientConfig, log *zerolog.Logger) (*Client, error) {
	builder := transport.NewHTTPTransportBuilder(cfg.S3TLSCertPath, nil, log)

	httptrprt, err := builder.Build()
	if err != nil {
		log.Error().Err(err).
			Str("tls_cert_path", cfg.S3TLSCertPath).
			Msg("failed to build http transport")

		return nil, fmt.Errorf("%w: MinIO http transport: %v", e.ErrInvalidInput, err.Error())
	}

	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Token),
		Secure:    true,
		Transport: httptrprt,
	})
	if err != nil {
		log.Error().Err(err).
			Str("endpoint", cfg.S3Endpoint).
			Msg("failed to initialize MinIO client")

		return nil, fmt.Errorf("%w: MinIO client: %v", e.ErrInit, err.Error())
	}

	return &Client{
		minio: client,
		cfg:   cfg,
		log:   log,
	}, nil
}

func (c *Client) PutObject(
	ctx context.Context,
	bucketName, objectName, filePath string,
	opts s3.PutObjectOptions,
) (s3.UploadInfo, error) {
	ctxLog := c.log.With().
		Str("bucket", bucketName).
		Str("object", objectName).
		Str("file", filePath).Logger()

	ctxLog.Info().Msg("uploading object to storage")

	start := time.Now()
	info, err := c.minio.FPutObject(ctx, bucketName, objectName, filePath, opts)
	duration := time.Since(start)

	if err != nil {
		ctxLog.Error().Err(err).
			Dur("duration", duration).
			Msg("failed to upload object to storage")

		return s3.UploadInfo{}, e.ErrInternal
	}

	ctxLog.Info().
		Dur("duration", duration).
		Msg("uploaded object to storage")

	return info, nil
}

// GetObject downloads an object from the specified bucket and writes it to filePath.
func (c *Client) GetObject(
	ctx context.Context,
	bucketName, objectName, filePath string,
	opts s3.GetObjectOptions,
) error {
	ctxLog := c.log.With().
		Str("bucket", bucketName).
		Str("object", objectName).
		Str("file", filePath).Logger()

	ctxLog.Info().Msg("downloading object from storage")

	start := time.Now()
	err := c.minio.FGetObject(ctx, bucketName, objectName, filePath, opts)
	duration := time.Since(start)

	if err != nil {
		ctxLog.Error().Err(err).
			Dur("duration", duration).
			Msg("failed to download object from storage")

		return e.ErrInternal
	}

	ctxLog.Info().
		Dur("duration", duration).
		Msg("downloaded object from storage")

	return nil
}
