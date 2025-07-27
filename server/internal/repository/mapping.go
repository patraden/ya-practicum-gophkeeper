//nolint:varnamelen // reason: function is reasonably short.
package repository

import (
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/secret"
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

func ToCreateSecretInitRequestParams(req *secret.InitRequest) pg.CreateSecretInitRequestParams {
	metaData, err := req.MetaData.MarshalJSON()
	if err != nil {
		metaData = []byte{}
	}

	return pg.CreateSecretInitRequestParams{
		UserID:           req.UserID,
		SecretID:         req.SecretID,
		SecretName:       req.SecretName,
		S3Url:            req.S3URL,
		VersionID:        req.VersionID,
		CurrentVersionID: req.ParentVersionID,
		RequestType:      pg.RequestType(req.RequestType),
		Token:            req.Token,
		ClientInfo:       req.ClientInfo,
		SecretSize:       req.SecretSize,
		SecretHash:       req.SecretHash,
		SecretDek:        req.SecretDEK,
		Meta:             metaData,
		CreatedAt:        req.CreatedAt,
		ExpiresAt:        req.ExpiresAt,
	}
}

func FromCreateSecretInitRequestParams(row pg.CreateSecretInitRequestRow) *secret.InitRequest {
	var metaData secret.MetaData
	if err := metaData.UnmarshalJSON(row.Meta); err != nil {
		metaData = secret.MetaData{}
	}

	return &secret.InitRequest{
		UserID:          row.UserID,
		SecretID:        row.SecretID,
		SecretName:      row.SecretName,
		S3URL:           row.S3Url,
		VersionID:       row.VersionID,
		ParentVersionID: row.ParentVersionID,
		RequestType:     secret.RequestType(row.RequestType),
		Token:           row.Token,
		ClientInfo:      row.ClientInfo,
		SecretSize:      row.SecretSize,
		SecretHash:      row.SecretHash,
		SecretDEK:       row.SecretDek,
		MetaData:        metaData,
		CreatedAt:       row.CreatedAt,
		ExpiresAt:       row.ExpiresAt,
	}
}

func FromPGIdentityToken(t pg.UserIdentityToken) *user.IdentityToken {
	return &user.IdentityToken{
		UserID:           t.UserID,
		AccessToken:      t.AccessToken,
		ExpiresAt:        t.ExpiresAt,
		RefreshExpiresAt: t.RefreshExpiresAt,
		RefreshToken:     t.RefreshToken,
		Scope:            "",
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
	}
}
