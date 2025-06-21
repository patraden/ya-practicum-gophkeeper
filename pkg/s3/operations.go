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
	SetBucketNotification(ctx context.Context, bucketName string) error
}

type URLManager interface {
	GeneratePresignedPutURL(
		ctx context.Context,
		bucketName,
		objectKey string,
		expiry time.Duration,
	) (*url.URL, error)
}

// SecurityManager defines methods for S3-related access control and identity federation.
type SecurityManager interface {
	// AssumeRole performs a STS AssumeRoleWithWebIdentity and returns temporary credentials.
	AssumeRole(ctx context.Context, identityToken string, durationSeconds int) (*TemporaryCredentials, error)
	// AddCannedPolicy attaches a pre-defined policy by name.
	AddCannedPolicy(ctx context.Context, name string, policyJSON []byte) error
}

// ServerOperator defines S3 operations required by the backend server.
// It includes bucket lifecycle management and presigned URL generation.
type ServerOperator interface {
	BucketManager
	URLManager
	SecurityManager
}

// ClientOperator defines S3 operations used by clients to upload and download objects.
type ClientOperator interface {
	PutObject(
		ctx context.Context,
		bucketName, objectName, filePath string,
		opts PutObjectOptions,
	) (UploadInfo, error)
	GetObject(
		ctx context.Context,
		bucketName, objectName, filePath string,
		opts GetObjectOptions,
	) error
}
