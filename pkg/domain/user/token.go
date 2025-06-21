package user

import (
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/google/uuid"
)

// tokenExpirationBuffer is subtracted from expiration times to avoid edge cases
// caused by clock drift, slow networks, or slight delays in token usage.
const tokenExpirationBuffer = 60 // seconds

// JWT represents a JSON Web Token returned by the identity provider.
// This is an alias for gocloak.JWT to decouple usage from the external library.
type JWT = gocloak.JWT

// IdentityToken represents an identity provider token associated with a user.
type IdentityToken struct {
	UserID           uuid.UUID `json:"user_id"`            // Internal GophKeeper user ID
	AccessToken      string    `json:"access_token"`       // Access token provided by the identity provider
	ExpiresAt        time.Time `json:"expires_at"`         // UTC time when the access token expires
	RefreshExpiresAt time.Time `json:"refresh_expires_at"` // UTC time when the refresh token expires
	RefreshToken     string    `json:"refresh_token"`      // Refresh token used to obtain a new access token
	Scope            string    `json:"scope"`              // Scope granted by the identity provider
	CreatedAt        time.Time `json:"created_at"`         // Timestamp of token creation
	UpdatedAt        time.Time `json:"updated_at"`         // Timestamp of last token update
}

// TokenFromJWT constructs an IdentityToken from a gocloak.JWT and the associated user ID.
func TokenFromJWT(uid uuid.UUID, jwt *JWT) *IdentityToken {
	if jwt == nil {
		return nil
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Second * time.Duration(jwt.ExpiresIn-tokenExpirationBuffer))
	refreshExpiresAt := now.Add(time.Second * time.Duration(jwt.RefreshExpiresIn-tokenExpirationBuffer))

	return &IdentityToken{
		UserID:           uid,
		AccessToken:      jwt.AccessToken,
		ExpiresAt:        expiresAt,
		RefreshToken:     jwt.RefreshToken,
		RefreshExpiresAt: refreshExpiresAt,
		Scope:            jwt.Scope,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// IsValid checks whether the access token is still valid, based on current UTC time.
func (t *IdentityToken) IsValid() bool {
	return t.AccessToken != "" && time.Now().UTC().Before(t.ExpiresAt)
}

// IsRefreshable checks whether the refresh token can still be used to obtain a new access token.
func (t *IdentityToken) IsRefreshable() bool {
	return t.RefreshToken != "" && time.Now().UTC().Before(t.RefreshExpiresAt)
}
