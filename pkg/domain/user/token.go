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
	AccessToken      string    `json:"access_token"`       // Access token from the identity provider
	ExpiresAt        time.Time `json:"expires_at"`         // UTC time when access token expires
	RefreshExpiresAt time.Time `json:"refresh_expires_at"` // UTC time when refresh token expires
	RefreshToken     string    `json:"refresh_token"`      // Refresh token for obtaining a new access token
	Scope            string    `json:"scope"`              // Granted scope from identity provider
	CreatedAt        time.Time `json:"created_at"`         // Time of token creation
	UpdatedAt        time.Time `json:"updated_at"`         // Time of last token update
}

// TokenFromJWT constructs an IdentityToken from a gocloak.JWT and user ID.
func TokenFromJWT(uid uuid.UUID, jwt *JWT) *IdentityToken {
	if jwt == nil {
		return nil
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Second * time.Duration(jwt.ExpiresIn))
	refreshExpiresAt := now.Add(time.Second * time.Duration(jwt.RefreshExpiresIn))

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

// IsValid returns true if the access token is still valid.
func (t *IdentityToken) IsValid() bool {
	return t.AccessToken != "" && t.ExpiresIn() > 0
}

// IsRefreshable returns true if the refresh token is still usable.
func (t *IdentityToken) IsRefreshable() bool {
	return t.RefreshToken != "" && t.RefreshExpiresIn() > 0
}

// ExpiresIn returns seconds until the access token expires, minus buffer.
// Returns 0 if already expired or within buffer range.
func (t *IdentityToken) ExpiresIn() int {
	remaining := int(t.ExpiresAt.Sub(time.Now().UTC()).Seconds()) - tokenExpirationBuffer
	if remaining < 0 {
		return 0
	}

	return remaining
}

// RefreshExpiresIn returns seconds until the refresh token expires, minus buffer.
// Returns 0 if already expired or within buffer range.
func (t *IdentityToken) RefreshExpiresIn() int {
	remaining := int(t.RefreshExpiresAt.Sub(time.Now().UTC()).Seconds()) - tokenExpirationBuffer
	if remaining < 0 {
		return 0
	}

	return remaining
}
