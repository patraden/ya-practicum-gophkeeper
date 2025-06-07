package s3

import (
	"context"
	"net/url"

	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/rs/zerolog"
)

// DummyS3 is a fake S3 client used for integration tests.
// It logs all operations but performs no real S3 interaction.
type DummyS3 struct {
	log *zerolog.Logger
}

// NewDummyS3 returns a new DummyS3 that logs simulated operations.
func NewDummyS3(log *zerolog.Logger) *DummyS3 {
	return &DummyS3{
		log: log,
	}
}

// MakeBucket logs a simulated bucket creation.
func (d *DummyS3) MakeBucket(_ context.Context, bucketName string, tags map[string]string) error {
	d.log.Info().
		Str("bucket", bucketName).
		Interface("tags", tags).
		Msg("DummyS3: simulated MakeBucket")

	return nil
}

// RemoveBucket logs a simulated bucket removal.
func (d *DummyS3) RemoveBucket(_ context.Context, bucketName string) error {
	d.log.Info().
		Str("bucket", bucketName).
		Msg("DummyS3: simulated RemoveBucket")

	return nil
}

// BucketExists logs the check and returns true (simulate that bucket always exists).
func (d *DummyS3) BucketExists(_ context.Context, bucketName string) (bool, error) {
	d.log.Info().
		Str("bucket", bucketName).
		Msg("DummyS3: simulated BucketExists (always true)")

	return true, nil
}

// GetPresignedURL logs the request and returns a dummy URL.
func (d *DummyS3) GetPresignedURL(_ context.Context) (*url.URL, error) {
	d.log.Info().
		Msg("DummyS3: simulated GetPresignedURL")

	url, err := url.Parse("https://dummy-s3.local/presigned-url")
	if err != nil {
		return nil, e.InternalErr(err)
	}

	// Return a dummy URL for testing purposes.
	return url, nil
}
