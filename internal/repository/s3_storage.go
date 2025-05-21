package repostory

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/patraden/ya-practicum-gophkeeper/internal/domain/errors"
	"github.com/patraden/ya-practicum-gophkeeper/internal/storage/s3"
	"github.com/rs/zerolog"
)

type S3Storage struct {
	s3.BucketManager
	client s3.MinioClient
	log    *zerolog.Logger
}

func NewS3Storage(client *minio.Client, log *zerolog.Logger) *S3Storage {
	return &S3Storage{
		client: client,
		log:    log,
	}
}

func (s *S3Storage) MakeBucket(ctx context.Context, bucketName string) error {
	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		s.log.Error().Err(err).
			Msg("failed to check if bucket exists")

		return errors.ErrObjectBucketCreate
	}

	if exists {
		return errors.ErrObjectBucketAlreadyExists
	}

	err = s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
		Region:        "us-east-1",
		ObjectLocking: true,
	})
	if err != nil {
		s.log.Error().Err(err).
			Str("bucket_name", bucketName).
			Msg("failed to create new bucket")

		return errors.ErrObjectBucketCreate
	}

	return nil
}
