package user

import (
	"crypto/rand"
	"time"

	"github.com/patraden/ya-practicum-gophkeeper/internal/domain/errors"
	"golang.org/x/crypto/bcrypt"
)

const passHashLenth = 16

type User struct {
	ID        ID        `json:"id"`
	Username  string    `json:"username"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	password  []byte    `json:"-"`
	Salt      []byte    `json:"-"`
}

func New(username string, role Role) *User {
	now := time.Now()

	return &User{
		ID:        NewUserID(),
		Username:  username,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
		password:  []byte{},
		Salt:      []byte{},
	}
}

func NewWithID(id string, username string, role Role) (*User, error) {
	uid, err := ParseID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return &User{
		ID:        uid,
		Username:  username,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
		password:  []byte{},
		Salt:      []byte{},
	}, nil
}

func (u *User) SetPassword(password string) error {
	salt := make([]byte, passHashLenth)
	if _, err := rand.Read(salt); err != nil {
		return errors.ErrUserPassHashing
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.ErrUserPassHashing
	}

	u.password = hashedPass
	u.Salt = salt

	return nil
}

func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword(u.password, []byte(password)) == nil && len(u.Salt) > 0
}
