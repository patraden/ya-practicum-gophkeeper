package s3

import (
	"context"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
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
	DoSomething(ctx context.Context)
}

// NewMinioClient initializes a new MinIO S3 client using the provided configuration
// and a custom HTTP transport. Returns a configured *minio.Client or an error.
func NewMinioClient(cfg *config.Config, transport *http.Transport) (*minio.Client, error) {
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Token),
		Secure:    true,
		Transport: transport,
	})
	if err != nil {
		return nil, e.ErrMinioClientCreate
	}

	return client, nil
}
