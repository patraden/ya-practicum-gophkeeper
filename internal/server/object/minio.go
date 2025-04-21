package object

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/patraden/ya-practicum-gophkeeper/internal/server/model"
)

type MinioStorage struct {
	BucketManager
	client *minio.Client
}

func (s *MinioStorage) MakeBucket(ctx context.Context, bucketName string) error {
	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return model.ErrObjectBucketCreate
	}

	if exists {
		return model.ErrObjectBucketAlreadyExists
	}

	err = s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
		Region:        "us-east-1",
		ObjectLocking: true,
	})
	if err != nil {
		return model.ErrObjectBucketCreate
	}

	return nil
}
