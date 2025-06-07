package s3

import (
	"context"
	"net/url"
)

// BucketManager interface for S3 bucker manager.
type BucketManager interface {
	MakeBucket(ctx context.Context, bucketName string, tags map[string]string) error
	RemoveBucket(ctx context.Context, bucketName string) error
	BucketExists(ctx context.Context, bucketName string) (bool, error)
}

// Client abstracts S3 operations used by the application.
type Client interface {
	BucketManager
	GetPresignedURL(ctx context.Context) (*url.URL, error)
}
