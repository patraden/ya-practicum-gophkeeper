package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/secret"
)

// SecretRequestIssued represents a secret request that has been issued but not necessarily completed.
type SecretRequestIssued struct {
	ID            int64              `json:"id"`
	UserID        uuid.UUID          `json:"user_id"`
	SecretID      uuid.UUID          `json:"secret_id"`
	Version       uuid.UUID          `json:"version"`
	ParentVersion uuid.UUID          `json:"parent_version,omitempty"`
	RequestType   secret.RequestType `json:"request_type"`
	PresignedURL  string             `json:"presigned_url,omitempty"`
	Token         int64              `json:"token"`
	ClientInfo    string             `json:"client_info,omitempty"`
	SecretSize    int                `json:"secret_size"`
	SecretHash    []byte             `json:"secret_hash,omitempty"`
	SecretDEK     []byte             `json:"secret_dek,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	ExpiresAt     time.Time          `json:"expires_at"`
}

// SecretRequestCompleted represents a completed secret request.
type SecretRequestCompleted struct {
	ID            int64                   `json:"id"`
	UserID        uuid.UUID               `json:"user_id"`
	SecretID      uuid.UUID               `json:"secret_id"`
	Version       uuid.UUID               `json:"version"`
	ParentVersion uuid.UUID               `json:"parent_version,omitempty"`
	RequestType   secret.RequestType      `json:"request_type"`
	PresignedURL  string                  `json:"presigned_url,omitempty"`
	Token         int64                   `json:"token"`
	ClientInfo    string                  `json:"client_info,omitempty"`
	SecretSize    int                     `json:"secret_size"`
	SecretHash    []byte                  `json:"secret_hash,omitempty"`
	SecretDEK     []byte                  `json:"secret_dek,omitempty"`
	CreatedAt     time.Time               `json:"created_at"`
	ExpiresAt     time.Time               `json:"expires_at"`
	FinishedAt    *time.Time              `json:"finished_at,omitempty"`
	Status        secret.RequestStatus    `json:"status"`
	CommittedBy   secret.RequestCommitter `json:"committed_by"`
}
