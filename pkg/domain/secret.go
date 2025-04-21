package domain

import (
	"time"

	"github.com/google/uuid"
)

type Secret struct {
	ID        uuid.UUID         `json:"id"`
	MetaData  map[string]string `json:"meta_data"`
	Version   int               `json:"version"`
	Content   []byte            `json:"-"`
	URL       *string           `json:"content_url"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}
