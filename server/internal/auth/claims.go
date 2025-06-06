package auth

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
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
	if userID, err := user.ParseID(c.UserID); err != nil || userID.IsNil() {
		return errors.ErrInvalidInput
	}

	return nil
}
