package auth

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
)

// Claims represents the JWT claims for a user.
// Includes the user ID and role along with standard JWT claims.
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Validate implements additional validations for claims.
func (c Claims) Validate() error {
	if userID, err := uuid.Parse(c.UserID); err != nil || userID == uuid.Nil {
		return errors.ErrInvalidInput
	}

	return nil
}
