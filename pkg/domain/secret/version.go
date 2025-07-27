package secret

import (
	"time"

	"github.com/google/uuid"
)

// Version represents a version of a secret.
type Version struct {
	ID         uuid.UUID `json:"version"`
	UserID     uuid.UUID `json:"user_id"`
	SecretID   uuid.UUID `json:"secret_id"`
	ParentID   uuid.UUID `json:"parent_version,omitempty"`
	S3URL      string    `json:"s3_url"`
	SecretSize int64     `json:"secret_size"`
	SecretHash []byte    `json:"secret_hash"`
	SecretDEK  []byte    `json:"secret_dek"`
	CreatedAt  time.Time `json:"created_at"`
}

func NewVersion(
	userID uuid.UUID,
	secretID uuid.UUID,
	parentID uuid.UUID,
	s3URL string,
	secretSize int64,
	secretHash []byte,
	secretDEK []byte,
) *Version {
	now := time.Now().UTC()

	return &Version{
		ID:         uuid.New(),
		UserID:     userID,
		SecretID:   secretID,
		ParentID:   parentID,
		S3URL:      s3URL,
		SecretSize: secretSize,
		SecretHash: secretHash,
		SecretDEK:  secretDEK,
		CreatedAt:  now,
	}
}
