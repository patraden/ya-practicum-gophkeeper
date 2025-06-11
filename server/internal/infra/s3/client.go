package s3

import (
	"context"
	"net/url"
	"time"
)

// BucketManager interface for S3 bucker manager.
type BucketManager interface {
	MakeBucket(ctx context.Context, bucketName string, tags map[string]string) error
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	RemoveBucket(ctx context.Context, bucketName string) error
}

// Client abstracts S3 operations used by the application.
type Client interface {
	BucketManager
	GeneratePresignedPutURL(ctx context.Context, bucketName, objectKey string, expiry time.Duration) (*url.URL, error)
}
