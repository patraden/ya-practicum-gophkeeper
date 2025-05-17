package s3

import (
	"context"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/patraden/ya-practicum-gophkeeper/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/internal/domain/errors"
)

// BucketManager interface for s3 bucker manager.
type BucketManager interface {
	MakeBucket(ctx context.Context, bucketName string) (err error)
}

// MinioClient abstracts MinIO operations used by the application.
// It is designed for easy mocking in tests and decoupling from the MinIO SDK.
type MinioClient interface {
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) (err error)
}

// NewMinioClient initializes a new MinIO client using the provided configuration
// and a custom HTTP transport. Returns a configured *minio.Client or an error.
func NewMinioClient(cfg *config.ObjectStorageConfig, transport *http.Transport) (*minio.Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, cfg.Token),
		Secure:    true,
		Transport: transport,
	})
	if err != nil {
		return nil, errors.ErrMinioClientCreate
	}

	return client, nil
}
