package user

import "time"

type Key struct {
	UserID    ID        `db:"user_id"`
	Kek       []byte    `db:"kek"`
	Algorithm string    `db:"algorithm"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewKey(id ID, kek []byte, algo string) *Key {
	now := time.Now().UTC()

	return &Key{
		UserID:    id,
		Kek:       kek,
		Algorithm: algo,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
