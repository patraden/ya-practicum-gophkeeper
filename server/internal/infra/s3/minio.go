package s3

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/notification"
	"github.com/minio/minio-go/v7/pkg/tags"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/net/transport"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/rs/zerolog"
)

// MinIOClient wraps the minio.Client and provides additional configuration
// and logging capabilities for interacting with an S3-compatible MinIO backend.
type MinIOClient struct {
	Client
	minio *minio.Client
	cfg   *config.Config
	log   *zerolog.Logger
}

// NewMinIOClient initializes a new MinIO S3 client using the provided configuration.
func NewMinIOClient(cfg *config.Config, log *zerolog.Logger) (*MinIOClient, error) {
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

	return &MinIOClient{
		minio: client,
		cfg:   cfg,
		log:   log,
	}, nil
}

func (c *MinIOClient) IsOnline() bool {
	return c.minio.IsOnline()
}

// logCtx constructs a structured logger pre-filled with bucket context.
func (c *MinIOClient) logCtx(bucketName string) zerolog.Logger {
	return c.log.With().
		Str("bucket_name", bucketName).
		Str("account_id", c.cfg.S3AccountID).
		Str("region", c.cfg.S3RedisRegion).Logger()
}

// BucketExists checks whether a bucket with the given name exists.
func (c *MinIOClient) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	logCtx := c.logCtx(bucketName)

	exists, err := c.minio.BucketExists(ctx, bucketName)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to check bucket existence")
		return false, e.InternalErr(err)
	}

	if exists {
		return true, nil
	}

	return false, nil
}

// MakeBucket creates a new bucket with the given name, tags, and sets up Redis notification events.
func (c *MinIOClient) MakeBucket(
	ctx context.Context,
	bucketName string,
	bucketTags map[string]string,
) error {
	logCtx := c.logCtx(bucketName)

	exists, err := c.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if exists {
		return e.ErrExists
	}

	err = c.minio.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: c.cfg.S3Region})
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to create new bucket")
		return e.InternalErr(err)
	}

	if len(bucketTags) > 0 {
		if err := c.setBucketTags(ctx, bucketName, bucketTags); err != nil {
			return err
		}
	}

	redisARN := notification.NewArn("minio", "sqs", c.cfg.S3RedisRegion, c.cfg.S3AccountID, "redis")
	queueConfig := notification.NewConfig(redisARN)

	queueConfig.AddEvents(
		notification.ObjectCreatedAll,
		notification.ObjectRemovedAll,
		notification.ObjectTransitionAll,
		notification.ObjectAccessedAll,
	)

	notifyCfg := notification.Configuration{}
	notifyCfg.AddQueue(queueConfig)

	err = c.minio.SetBucketNotification(ctx, bucketName, notifyCfg)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to set bucket notifications")
		return e.InternalErr(err)
	}

	return nil
}

// setBucketTags applies the provided tags to the specified bucket.
func (c *MinIOClient) setBucketTags(
	ctx context.Context,
	bucketName string,
	bucketTags map[string]string,
) error {
	logCtx := c.logCtx(bucketName)

	tgs, err := tags.NewTags(bucketTags, false)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to create tags")
		return e.InternalErr(err)
	}

	err = c.minio.SetBucketTagging(ctx, bucketName, tgs)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to set bucket tags")
		return e.InternalErr(err)
	}

	return nil
}

// GeneratePresignedPutURL generates a presigned PUT URL for uploading an object to a bucket.
func (c *MinIOClient) GeneratePresignedPutURL(
	ctx context.Context,
	bucketName, objectKey string,
	expiry time.Duration,
) (*url.URL, error) {
	logCtx := c.logCtx(bucketName)

	presignedURL, err := c.minio.PresignedPutObject(ctx, bucketName, objectKey, expiry)
	if err != nil {
		logCtx.Error().Err(err).
			Msg("failed to generate presigned PUT URL")

		return nil, e.InternalErr(err)
	}

	logCtx.Debug().
		Str("presigned_url", presignedURL.String()).
		Msg("generated presigned PUT URL")

	return presignedURL, nil
}

// RemoveBucket deletes the specified bucket if it exists and is empty.
func (c *MinIOClient) RemoveBucket(ctx context.Context, bucketName string) error {
	logCtx := c.logCtx(bucketName)

	exists, err := c.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if !exists {
		logCtx.Warn().Msg("bucket does not exist")
		return e.ErrNotFound
	}

	if err := c.minio.RemoveBucket(ctx, bucketName); err != nil {
		logCtx.Error().Err(err).Msg("failed to remove bucket")
		return e.InternalErr(err)
	}

	logCtx.Info().Msg("bucket removed successfully")

	return nil
}
