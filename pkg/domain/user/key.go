package user

import (
	"time"

	"github.com/google/uuid"
)

type Key struct {
	UserID    uuid.UUID `db:"user_id"`
	Kek       []byte    `db:"kek"`
	Algorithm string    `db:"algorithm"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewKey(id uuid.UUID, kek []byte, algo string) *Key {
	now := time.Now().UTC()

	return &Key{
		UserID:    id,
		Kek:       kek,
		Algorithm: algo,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
