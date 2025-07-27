package identity_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/identity"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/pg"
	"github.com/stretchr/testify/require"
)

func setupTokenRepo(t *testing.T) (pgxmock.PgxPoolIface, identity.TokenCache, *user.IdentityToken) {
	t.Helper()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	db := &pg.DB{ConnPool: mockPool}

	repo := identity.NewPGIdentityTokenCache(db)

	token := &user.IdentityToken{
		UserID:           uuid.New(),
		AccessToken:      "access",
		RefreshToken:     "refresh",
		ExpiresAt:        time.Now().Add(time.Minute),
		RefreshExpiresAt: time.Now().Add(time.Hour),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return mockPool, repo, token
}

func TestIdentityTokenRepoUpsert(t *testing.T) {
	t.Parallel()

	mockPool, repo, token := setupTokenRepo(t)

	mockPool.ExpectExec(`INSERT INTO user_identity_tokens`).
		WithArgs(
			token.UserID, token.AccessToken, token.RefreshToken,
			token.ExpiresAt, token.RefreshExpiresAt, token.CreatedAt, token.UpdatedAt,
		).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err := repo.Upsert(context.Background(), token)
	require.NoError(t, err)
	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestIdentityTokenRepoGet(t *testing.T) {
	t.Parallel()
	mockPool, repo, token := setupTokenRepo(t)

	mockPool.ExpectQuery(`SELECT user_id, access_token, refresh_token`).
		WithArgs(token.UserID).
		WillReturnRows(pgxmock.NewRows([]string{
			"user_id", "access_token", "refresh_token",
			"expires_at", "refresh_expires_at", "created_at", "updated_at",
		}).AddRow(
			token.UserID, token.AccessToken, token.RefreshToken,
			token.ExpiresAt, token.RefreshExpiresAt, token.CreatedAt, token.UpdatedAt,
		))

	tok, err := repo.Get(context.Background(), token.UserID)
	require.NoError(t, err)
	require.Equal(t, token.UserID, tok.UserID)
	require.NoError(t, mockPool.ExpectationsWereMet())
}

func TestIdentityTokenRepoDelete(t *testing.T) {
	t.Parallel()

	mockPool, repo, token := setupTokenRepo(t)

	mockPool.ExpectExec(`DELETE FROM user_identity_tokens`).
		WithArgs(token.UserID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := repo.Delete(context.Background(), token.UserID)
	require.NoError(t, err)
	require.NoError(t, mockPool.ExpectationsWereMet())
}
