package secret

import (
	"time"

	"github.com/google/uuid"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/s3"
)

type InitRequest struct {
	UserID        uuid.UUID
	SecretID      uuid.UUID
	SecretName    string
	S3URL         string
	Version       uuid.UUID
	ParentVersion uuid.UUID
	RequestType   RequestType
	Token         int64
	ClientInfo    string
	SecretSize    int64
	SecretHash    []byte
	SecretDEK     []byte
	MetaData      MetaData
	CreatedAt     time.Time
	ExpiresAt     time.Time
	S3Creds       *s3.TemporaryCredentials
	User          *user.User
}

func (req *InitRequest) UploadDuration() int {
	const (
		baseBufferSeconds = 15 * 60          // Base minimum duration for small files
		uploadSpeedBps    = 1 * 1024 * 1024  // Assume 1 MB/s upload speed
		maxDuration       = 60 * 60 * 24 * 7 // Cap duration to 15 minutes (900 seconds)
		durationFactor    = 1.25             // Adds a 25% buffer
	)

	estimatedUploadTime := max(int(req.SecretSize)/uploadSpeedBps, baseBufferSeconds)
	totalDuration := int(float64(estimatedUploadTime) * durationFactor)

	if totalDuration > maxDuration {
		return maxDuration
	}

	return totalDuration
}

type CommitRequest struct {
	UserID        uuid.UUID
	SecretID      uuid.UUID
	S3URL         string
	Version       uuid.UUID
	ParentVersion uuid.UUID
	RequestType   RequestType
	Token         int64
	ClientInfo    string
	SecretSize    int64
	SecretHash    []byte
	SecretDEK     []byte
	CreatedAt     time.Time
	ExpiresAt     time.Time
	FinishedAt    time.Time
	Status        RequestStatus
	CommittedBy   RequestCommitter
}
