package user

import (
	"time"
)

type User struct {
	ID        ID        `json:"id"`
	Username  string    `json:"username"`
	Password  []byte    `json:"-"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func New(username string, role Role) *User {
	return &User{
		ID:        NewUserID(),
		Username:  username,
		Password:  []byte{},
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func NewWithID(id string, username string, role Role) (*User, error) {
	uid, err := ParseUserID(id)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        uid,
		Username:  username,
		Password:  []byte{},
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}
