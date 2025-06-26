package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/keys"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/secret"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/rs/zerolog"
)

// SecretUploadInitRequest represents an upload request that is in progress.
type SecretUploadInitRequest struct {
	UserID        string             `json:"user_id"`
	SecretID      string             `json:"secret_id"`
	SecretName    string             `json:"secret_name"`
	Version       string             `json:"version"`
	ParentVersion string             `json:"parent_version,omitempty"`
	RequestType   secret.RequestType `json:"request_type"`
	S3URL         string             `json:"url,omitempty"`
	Token         int64              `json:"token"`
	ClientInfo    string             `json:"client_info"`
	SecretSize    int64              `json:"secret_size"`
	SecretHash    []byte             `json:"secret_hash,omitempty"`
	SecretDEK     []byte             `json:"secret_dek,omitempty"`
	MetaData      secret.MetaData    `json:"meta,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	ExpiresAt     time.Time          `json:"expires_at"`
}

func (r *SecretUploadInitRequest) ToDomain(log zerolog.Logger) (*secret.InitRequest, error) {
	userID, err := uuid.Parse(r.UserID)
	if err != nil {
		log.Error().Err(err).
			Msg("validation failed: invalid user_id UUID")

		return nil, e.ErrValidation
	}

	secretID, err := uuid.Parse(r.SecretID)
	if err != nil {
		log.Error().Err(err).
			Msg("validation failed: invalid secret_id UUID")

		return nil, e.ErrValidation
	}

	version, err := uuid.Parse(r.Version)
	if err != nil {
		log.Error().Err(err).
			Msg("validation failed: invalid version UUID")

		return nil, e.ErrValidation
	}

	var parent uuid.UUID

	if r.ParentVersion != "" {
		pparent, err := uuid.Parse(r.ParentVersion)
		if err != nil {
			log.Error().Err(err).
				Msg("validation failed: invalid parent_version UUID")

			return nil, e.ErrValidation
		}

		parent = pparent
	}

	return &secret.InitRequest{
		UserID:        userID,
		SecretID:      secretID,
		SecretName:    r.SecretName,
		Version:       version,
		ParentVersion: parent,
		RequestType:   r.RequestType,
		S3URL:         r.S3URL,
		Token:         r.Token,
		ClientInfo:    r.ClientInfo,
		SecretSize:    r.SecretSize,
		SecretHash:    r.SecretHash,
		SecretDEK:     r.SecretDEK,
		MetaData:      r.MetaData,
		CreatedAt:     r.CreatedAt,
		ExpiresAt:     r.ExpiresAt,
	}, nil
}

func (r *SecretUploadInitRequest) Validate(kek []byte, log zerolog.Logger) error {
	switch {
	case r.SecretName == "":
		log.Error().Msg("validation failed: secret_name is empty")
	case r.RequestType != secret.RequestTypePut:
		log.Error().
			Str("type", string(r.RequestType)).
			Msg("validation failed: request_type must be 'put'")
	case r.ClientInfo == "":
		log.Error().Msg("validation failed: client_info is empty")
	case r.SecretSize == 0:
		log.Error().
			Int64("size", r.SecretSize).
			Msg("validation failed: secret_size must be positive")
	case len(r.SecretHash) == 0:
		log.Error().
			Int("hash_len", len(r.SecretHash)).
			Msg("validation failed: secret_hash must be positive")
	case len(r.SecretDEK) == 0:
		log.Error().
			Int("dek_len", len(r.SecretDEK)).
			Msg("validation failed: secret_dek must be positive")
	case r.ExpiresAt.Before(time.Now().UTC()):
		log.Error().
			Time("expires_at", r.ExpiresAt).
			Msg("validation failed: expires_at in the past")
	default:
		if _, err := keys.UnwrapDEK(kek, r.SecretDEK); err != nil {
			log.Error().Err(err).
				Msg("validation failed: failed to unwrap secret_dek")
		} else {
			return nil
		}
	}

	return e.ErrValidation
}

// SecretUploadCommitRequest represents a finalized and committed upload.
type SecretUploadCommitRequest struct {
	UserID        string                  `json:"user_id"`
	SecretID      string                  `json:"secret_id"`
	Version       string                  `json:"version"`
	ParentVersion *string                 `json:"parent_version,omitempty"`
	RequestType   secret.RequestType      `json:"request_type"`
	S3URL         string                  `json:"url,omitempty"`
	Token         int64                   `json:"token"`
	ClientInfo    string                  `json:"client_info"`
	SecretSize    int64                   `json:"secret_size"`
	SecretHash    []byte                  `json:"secret_hash,omitempty"`
	SecretDEK     []byte                  `json:"secret_dek,omitempty"`
	CreatedAt     time.Time               `json:"created_at"`
	ExpiresAt     time.Time               `json:"expires_at"`
	FinishedAt    time.Time               `json:"finished_at"`
	Status        secret.RequestStatus    `json:"status"`
	CommittedBy   secret.RequestCommitter `json:"committed_by"`
}

func (r *SecretUploadCommitRequest) ToDomain(log zerolog.Logger) (*secret.CommitRequest, error) {
	userID, err := uuid.Parse(r.UserID)
	if err != nil {
		log.Error().Err(err).
			Msg("validation failed: invalid user_id UUID")
		return nil, e.ErrValidation
	}

	secretID, err := uuid.Parse(r.SecretID)
	if err != nil {
		log.Error().Err(err).
			Msg("validation failed: invalid secret_id UUID")
		return nil, e.ErrValidation
	}

	version, err := uuid.Parse(r.Version)
	if err != nil {
		log.Error().Err(err).
			Msg("validation failed: invalid version UUID")
		return nil, e.ErrValidation
	}

	var parent uuid.UUID

	if r.ParentVersion != nil {
		pparent, err := uuid.Parse(*r.ParentVersion)
		if err != nil {
			log.Error().Err(err).
				Msg("validation failed: invalid parent_version UUID")
			return nil, e.ErrValidation
		}

		parent = pparent
	}

	return &secret.CommitRequest{
		UserID:        userID,
		SecretID:      secretID,
		S3URL:         r.S3URL,
		Version:       version,
		ParentVersion: parent,
		RequestType:   r.RequestType,
		Token:         r.Token,
		ClientInfo:    r.ClientInfo,
		SecretSize:    r.SecretSize,
		SecretHash:    r.SecretHash,
		SecretDEK:     r.SecretDEK,
		CreatedAt:     r.CreatedAt,
		ExpiresAt:     r.ExpiresAt,
		FinishedAt:    r.FinishedAt,
		Status:        r.Status,
		CommittedBy:   r.CommittedBy,
	}, nil
}

// Validate validates UUID fields and basic constraints.
func (r *SecretUploadCommitRequest) Validate(log zerolog.Logger) error {
	switch {
	case r.RequestType != secret.RequestTypePut:
		log.Error().
			Str("type", string(r.RequestType)).
			Msg("validation failed: request_type must be 'put'")
	case r.ClientInfo == "":
		log.Error().Msg("validation failed: client_info is empty")
	case r.SecretSize == 0:
		log.Error().
			Int64("size", r.SecretSize).
			Msg("validation failed: secret_size must be positive")
	case len(r.SecretHash) == 0:
		log.Error().
			Int("hash_len", len(r.SecretHash)).
			Msg("validation failed: secret_hash must be positive")
	case len(r.SecretDEK) == 0:
		log.Error().
			Int("dek_len", len(r.SecretDEK)).
			Msg("validation failed: secret_dek must be positive")
	case r.ExpiresAt.Before(time.Now().UTC()):
		log.Error().
			Time("expires_at", r.ExpiresAt).
			Msg("validation failed: expires_at in the past")
	case r.FinishedAt.Before(r.CreatedAt):
		log.Error().
			Time("finished_at", r.FinishedAt).
			Time("created_at", r.CreatedAt).
			Msg("validation failed: finished_at is before created_at")
	default:
		return nil
	}

	return e.ErrValidation
}
