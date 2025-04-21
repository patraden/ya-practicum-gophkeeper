package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Password  []byte    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func New(username string) *User {
	return &User{
		ID:        uuid.New(),
		Username:  username,
		Password:  []byte{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func NewWithID(id string, username string) (*User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        uid,
		Username:  username,
		Password:  []byte{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}
