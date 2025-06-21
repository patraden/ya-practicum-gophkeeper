//nolint:varnamelen // reason: function is reasonably short.
package repository

import (
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/pg"
)

// ToCreateUserParams maps a domain-level User to pg.CreateUserParams
// for use in SQL inserts via sqlc.
func ToCreateUserParams(u *user.User) pg.CreateUserParams {
	return pg.CreateUserParams{
		ID:         u.ID,
		Username:   u.Username,
		Role:       u.Role,
		Password:   u.Password,
		Salt:       u.Salt,
		Verifier:   u.Verifier,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
		BucketName: u.BucketName,
		IdentityID: u.IdentityID,
	}
}

// FromPGUser maps a pg.User (returned by sqlc) to a domain-level User model.
func FromPGUser(u pg.User) *user.User {
	return &user.User{
		ID:         u.ID,
		Username:   u.Username,
		Role:       u.Role,
		Password:   u.Password,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
		Salt:       u.Salt,
		Verifier:   u.Verifier,
		BucketName: u.BucketName,
		IdentityID: u.IdentityID,
	}
}

// ToCreateUserKeyParams maps a domain-level Key to pg.CreateUserKeyParams
// for SQL insert using sqlc.
func ToCreateUserKeyParams(k *user.Key) pg.CreateUserKeyParams {
	return pg.CreateUserKeyParams{
		UserID:    k.UserID,
		Kek:       k.Kek,
		Algorithm: k.Algorithm,
		CreatedAt: k.CreatedAt,
		UpdatedAt: k.UpdatedAt,
	}
}

// FromPGKey maps a pg.Key (returned by sqlc) to a domain-level Key model.
func FromPGKey(k pg.UserCryptoKey) *user.Key {
	return &user.Key{
		UserID:    k.UserID,
		Kek:       k.Kek,
		Algorithm: k.Algorithm,
		CreatedAt: k.CreatedAt,
		UpdatedAt: k.UpdatedAt,
	}
}

func ToCreateIdentityTokenParams(t *user.IdentityToken) pg.CreateIdentityTokenParams {
	return pg.CreateIdentityTokenParams{
		UserID:           t.UserID,
		AccessToken:      t.AccessToken,
		RefreshToken:     t.RefreshToken,
		ExpiresAt:        t.ExpiresAt,
		RefreshExpiresAt: t.RefreshExpiresAt,
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
	}
}
