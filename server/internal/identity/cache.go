package identity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/pg"
)

type TokenCache interface {
	Get(ctx context.Context, userID uuid.UUID) (*user.IdentityToken, error)
	Upsert(ctx context.Context, token *user.IdentityToken) error
	Delete(ctx context.Context, userID uuid.UUID) error
}

// IdentityTokenRepo implements IdentityTokenRepository using PostgreSQL.
type PGIdentityTokenCache struct {
	TokenCache
	connPool pg.ConnectionPool
	queries  *pg.Queries
}

func NewPGIdentityTokenCache(db *pg.DB) *PGIdentityTokenCache {
	return &PGIdentityTokenCache{
		connPool: db.ConnPool,
		queries:  pg.New(db.ConnPool),
	}
}

func (repo *PGIdentityTokenCache) Delete(ctx context.Context, userID uuid.UUID) error {
	if err := repo.queries.DeleteIdentityToken(ctx, userID); err != nil {
		return e.InternalErr(err)
	}

	return nil
}

func (repo *PGIdentityTokenCache) Upsert(ctx context.Context, token *user.IdentityToken) error {
	err := repo.queries.CreateIdentityToken(ctx, pg.CreateIdentityTokenParams{
		UserID:           token.UserID,
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		ExpiresAt:        token.ExpiresAt,
		RefreshExpiresAt: token.RefreshExpiresAt,
		CreatedAt:        token.CreatedAt,
		UpdatedAt:        token.UpdatedAt,
	})
	if err != nil {
		return e.InternalErr(err)
	}

	return nil
}

func (repo *PGIdentityTokenCache) Get(ctx context.Context, userID uuid.UUID) (*user.IdentityToken, error) {
	row, err := repo.queries.GetIdentityToken(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("[%w] identity token", e.ErrNotFound)
	}

	if err != nil {
		return nil, e.InternalErr(err)
	}

	return &user.IdentityToken{
		UserID:           row.UserID,
		AccessToken:      row.AccessToken,
		RefreshToken:     row.RefreshToken,
		ExpiresAt:        row.ExpiresAt,
		RefreshExpiresAt: row.RefreshExpiresAt,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}, nil
}

// InMemoryIdentityTokenCache provides a thread-safe in-memory token cache.
// Useful for development, testing, or fallback when PG is unavailable.
type InMemoryIdentityTokenCache struct {
	mu     sync.RWMutex
	tokens map[uuid.UUID]*user.IdentityToken
}

// NewInMemoryIdentityTokenCache initializes a new in-memory token cache.
func NewInMemoryIdentityTokenCache() *InMemoryIdentityTokenCache {
	return &InMemoryIdentityTokenCache{
		tokens: make(map[uuid.UUID]*user.IdentityToken),
	}
}

// Get returns the token for a given user, or an error if not found.
func (c *InMemoryIdentityTokenCache) Get(_ context.Context, userID uuid.UUID) (*user.IdentityToken, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	token, ok := c.tokens[userID]
	if !ok {
		return nil, fmt.Errorf("[%w] identity token", e.ErrNotFound)
	}

	return token, nil
}

// Upsert adds or updates the token for a given user.
func (c *InMemoryIdentityTokenCache) Upsert(_ context.Context, token *user.IdentityToken) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	token.UpdatedAt = time.Now().UTC()
	c.tokens[token.UserID] = token

	return nil
}

// Delete removes the token for a given user.
func (c *InMemoryIdentityTokenCache) Delete(_ context.Context, userID uuid.UUID) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.tokens, userID)

	return nil
}
