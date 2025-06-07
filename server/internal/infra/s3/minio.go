package s3

import (
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
)

// NewMinioClient initializes a new MinIO S3 client using the provided configuration
// and a custom HTTP transport. Returns a configured *minio.Client or an error.
func NewMinioClient(cfg *config.Config, transport *http.Transport) (*minio.Client, error) {
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Token),
		Secure:    true,
		Transport: transport,
	})
	if err != nil {
		return nil, e.InternalErr(err)
	}

	return client, nil
}
