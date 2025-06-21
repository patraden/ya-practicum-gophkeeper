package s3

import (
	"context"
	"net/url"

	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/rs/zerolog"
)

// DummyS3 is a fake S3 client used for integration tests.
// It logs all operations but performs no real S3 interaction.
type DummyServerOperator struct {
	log *zerolog.Logger
}

// NewDummyServerOperator returns a new DummyServerOperator that logs simulated operations.
func NewDummyServerOperator(log *zerolog.Logger) *DummyServerOperator {
	return &DummyServerOperator{
		log: log,
	}
}

// MakeBucket logs a simulated bucket creation.
func (d *DummyServerOperator) MakeBucket(_ context.Context, bucketName string, tags map[string]string) error {
	d.log.Info().
		Str("bucket", bucketName).
		Interface("tags", tags).
		Msg("DummyServerOperator: simulated MakeBucket")

	return nil
}

// RemoveBucket logs a simulated bucket removal.
func (d *DummyServerOperator) RemoveBucket(_ context.Context, bucketName string) error {
	d.log.Info().
		Str("bucket", bucketName).
		Msg("DummyServerOperator: simulated RemoveBucket")

	return nil
}

// BucketExists logs the check and returns true (simulate that bucket always exists).
func (d *DummyServerOperator) BucketExists(_ context.Context, bucketName string) (bool, error) {
	d.log.Info().
		Str("bucket", bucketName).
		Msg("DummyServerOperator: simulated BucketExists (always true)")

	return true, nil
}

// GetPresignedURL logs the request and returns a dummy URL.
func (d *DummyServerOperator) GetPresignedURL(_ context.Context) (*url.URL, error) {
	d.log.Info().
		Msg("DummyServerOperator: simulated GetPresignedURL")

	url, err := url.Parse("https://dummy-s3.local/presigned-url")
	if err != nil {
		return nil, e.InternalErr(err)
	}

	// Return a dummy URL for testing purposes.
	return url, nil
}
