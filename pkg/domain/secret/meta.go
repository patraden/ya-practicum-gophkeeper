package secret

import (
	"time"

	"github.com/google/uuid"
)

//easyjson:json
type MetaData map[string]string

//easyjson:json
type Meta struct {
	UserID    uuid.UUID `json:"user_id"`
	SecretID  uuid.UUID `json:"secret_id"`
	Data      MetaData  `json:"meta"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewMeta(
	userID uuid.UUID,
	secretID uuid.UUID,
	data MetaData,
) *Meta {
	now := time.Now().UTC()

	return &Meta{
		UserID:    userID,
		SecretID:  secretID,
		Data:      data,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
