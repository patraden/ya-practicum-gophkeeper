package minio

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
	"github.com/patraden/ya-practicum-gophkeeper/pkg/s3"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/rs/zerolog"
)

// MinIOClient wraps the minio.Client and provides additional configuration
// and logging capabilities for interacting with an S3-compatible MinIO backend.
type Client struct {
	s3.ServerOperator
	minio       *minio.Client
	webIDClient *WebIdentityClient
	cfg         *s3.ClientConfig
	log         zerolog.Logger
}

// NewMinIOClient initializes a new MinIO S3 client using the provided configuration.
func NewClient(config *config.Config, log zerolog.Logger) (*Client, error) {
	cfg := &s3.ClientConfig{
		S3Endpoint:    config.S3Endpoint,
		S3TLSCertPath: config.S3TLSCertPath,
		S3AccessKey:   config.S3AccessKey,
		S3SecretKey:   config.S3SecretKey,
		S3AccountID:   config.S3AccountID,
		S3Region:      config.S3Region,
		S3Token:       config.S3Token,
	}

	builder := transport.NewHTTPTransportBuilder(cfg.S3TLSCertPath, nil, log)

	httptrprt, err := builder.Build()
	if err != nil {
		return nil, err
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

		return nil, fmt.Errorf("[%w] MinIO client", e.ErrInit)
	}

	secure := cfg.S3TLSCertPath != ""
	webURL := getWebURL(cfg.S3Endpoint, secure)
	webIDClient := NewMinioWebIdentityClient(webURL, nil, httptrprt, log)

	return &Client{
		minio:       client,
		webIDClient: webIDClient,
		cfg:         cfg,
		log:         log,
	}, nil
}

func (c *Client) IsOnline() bool {
	return c.minio.IsOnline()
}

// getWebURL constructs the full MinIO web URL from endpoint and TLS setting.
func getWebURL(endpoint string, secure bool) string {
	scheme := "http"
	if secure {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s", scheme, endpoint)
}

// logCtx constructs a structured logger pre-filled with bucket context.
func (c *Client) logCtx(bucketName string) zerolog.Logger {
	return c.log.With().
		Str("bucket_name", bucketName).
		Str("account_id", c.cfg.S3AccountID).
		Str("region", c.cfg.S3Region).Logger()
}

// BucketExists checks whether a bucket with the given name exists.
func (c *Client) BucketExists(ctx context.Context, bucketName string) (bool, error) {
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
func (c *Client) MakeBucket(
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
		return fmt.Errorf("[%w] MinIO bucket", e.ErrExists)
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

	return nil
}

func (c *Client) SetBucketNotification(ctx context.Context, bucketName string) error {
	logCtx := c.logCtx(bucketName)

	redisARN := notification.NewArn("minio", "sqs", c.cfg.S3Region, c.cfg.S3AccountID, "redis")
	queueConfig := notification.NewConfig(redisARN)

	queueConfig.AddEvents(
		notification.ObjectCreatedAll,
		notification.ObjectRemovedAll,
		notification.ObjectTransitionAll,
		notification.ObjectAccessedAll,
	)

	notifyCfg := notification.Configuration{}
	notifyCfg.AddQueue(queueConfig)

	err := c.minio.SetBucketNotification(ctx, bucketName, notifyCfg)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to set bucket notifications")
		return e.InternalErr(err)
	}

	return nil
}

// setBucketTags applies the provided tags to the specified bucket.
func (c *Client) setBucketTags(
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
func (c *Client) GeneratePresignedPutURL(
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
func (c *Client) RemoveBucket(ctx context.Context, bucketName string) error {
	logCtx := c.logCtx(bucketName)

	exists, err := c.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if !exists {
		logCtx.Info().Msg("bucket does not exist")
		return fmt.Errorf("[%w] MinIO bucket", e.ErrNotFound)
	}

	if err := c.minio.RemoveBucket(ctx, bucketName); err != nil {
		logCtx.Error().Err(err).Msg("failed to remove bucket")
		return e.InternalErr(err)
	}

	logCtx.Info().Msg("bucket removed successfully")

	return nil
}

func (c *Client) AssumeRole(
	ctx context.Context,
	identityToken string,
	durationSeconds int,
) (*s3.TemporaryCredentials, error) {
	return c.webIDClient.AssumeRole(ctx, identityToken, durationSeconds)
}

func (c *Client) AddCannedPolicy(_ context.Context, _ string, _ []byte) error {
	// In a future iteration, implement support for a custom MinIO policy that restricts
	// access to objects based on the user_id extracted from the identity JWT token.
	// This policy should be created during server installation (via the `-install` flag)
	// and replace the current value of the user's "policy" attribute in the identity provider.
	return e.ErrNotImplemented
}
