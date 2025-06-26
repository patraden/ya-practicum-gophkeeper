package secret

import (
	"time"

	"github.com/google/uuid"
)

type (
	RequestType      string
	RequestStatus    string
	RequestCommitter string
)

const (
	RequestTypePut RequestType = "put"
	RequestTypeGet RequestType = "get"

	RequestStatusCompleted RequestStatus = "completed"
	RequestStatusAborted   RequestStatus = "aborted"
	RequestStatusExpired   RequestStatus = "expired"
	RequestStatusCancelled RequestStatus = "cancelled"

	RequestCommitterUser   RequestCommitter = "user"
	RequestCommitterServer RequestCommitter = "server"
	RequestCommitterS3     RequestCommitter = "s3"
)

type Secret struct {
	ID               uuid.UUID `json:"secret_id"`
	UserID           uuid.UUID `json:"user_id"`
	Name             string    `json:"secret_name"`
	CurrentVersionID uuid.UUID `json:"current_version"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	CurrentVersion   *Version  `json:"-"`
	Meta             *Meta     `json:"-"`
}

func (s *Secret) SetCurrentVersion(
	parentVersion uuid.UUID,
	s3URL string,
	secretSize uint64,
	secretHash []byte,
	secretDEK []byte,
) *Secret {
	ver := NewVersion(s.UserID, s.ID, parentVersion, s3URL, secretSize, secretHash, secretDEK)

	return &Secret{
		ID:               s.ID,
		UserID:           s.UserID,
		Name:             s.Name,
		CurrentVersionID: ver.ID,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
		CurrentVersion:   ver,
		Meta:             s.Meta,
	}
}

func (s *Secret) SetMeta(data map[string]string) *Secret {
	meta := NewMeta(s.UserID, s.ID, data)

	return &Secret{
		ID:               s.ID,
		UserID:           s.UserID,
		Name:             s.Name,
		CurrentVersionID: s.CurrentVersionID,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
		CurrentVersion:   s.CurrentVersion,
		Meta:             meta,
	}
}

func NewSecret(userID uuid.UUID, secretName string) *Secret {
	now := time.Now().UTC()

	return &Secret{
		ID:               uuid.New(),
		UserID:           userID,
		Name:             secretName,
		CurrentVersionID: uuid.Nil,
		CreatedAt:        now,
		UpdatedAt:        now,
		CurrentVersion:   nil,
		Meta:             nil,
	}
}
