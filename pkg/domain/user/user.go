package user

import (
	"bytes"
	"crypto/rand"
	"time"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/auth"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const passHashLength = 16

// User represents an application user with credentials and metadata.
type User struct {
	ID        ID        `json:"id"`         // Unique user identifier
	Username  string    `json:"username"`   // Username of the user
	Role      Role      `json:"role"`       // Role assigned to the user
	CreatedAt time.Time `json:"created_at"` // Timestamp of user creation
	UpdatedAt time.Time `json:"updated_at"` // Timestamp of last user update
	Password  []byte    `json:"-"`          // Bcrypt-hashed password (not exposed in JSON)
	Salt      []byte    `json:"salt"`       // Random salt used for verifier generation
	Verifier  []byte    `json:"verifier"`   // HMAC-based verifier derived from password and salt
}

// New creates a new user with a generated ID and current timestamps.
func New(username string, role Role) *User {
	now := time.Now().UTC()

	return &User{
		ID:        NewUserID(),
		Username:  username,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
		Password:  []byte{},
		Salt:      []byte{},
		Verifier:  []byte{},
	}
}

// NewWithID creates a new user with a provided string ID.
// Returns an error if the ID cannot be parsed.
func NewWithID(id string, username string, role Role) (*User, error) {
	uid, err := ParseID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	return &User{
		ID:        uid,
		Username:  username,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
		Password:  []byte{},
		Salt:      []byte{},
		Verifier:  []byte{},
	}, nil
}

// SetPassword hashes the given password, generates a salt and verifier, and updates the user.
func (u *User) SetPassword(password string) error {
	salt := make([]byte, passHashLength)

	if _, err := rand.Read(salt); err != nil {
		return e.ErrGenerate
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return e.ErrGenerate
	}

	verifier := auth.GenerateVerifier(password, salt)

	u.Password = hashedPass
	u.Salt = salt
	u.Verifier = verifier

	return nil
}

// CheckPassword verifies that the provided password matches the stored hash and verifier.
func (u *User) CheckPassword(password string) bool {
	if len(u.Salt) == 0 || len(u.Password) == 0 || len(u.Verifier) == 0 {
		return false
	}

	if err := bcrypt.CompareHashAndPassword(u.Password, []byte(password)); err != nil {
		return false
	}

	return auth.VerifyVerifier(password, u.Salt, u.Verifier)
}

// CheckVerifier returns true if the given verifier matches the stored verifier.
func (u *User) CheckVerifier(verifier []byte) bool {
	return bytes.Equal(u.Verifier, verifier)
}
