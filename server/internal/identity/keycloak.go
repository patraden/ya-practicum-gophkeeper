package identity

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/keycloak"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/pg"
	"github.com/rs/zerolog"
)

// KeycloakCachedManager is a concrete implementation of IdentityManager.
// It manages user accounts and token lifecycles using a Keycloak-compatible identity provider.
type KeycloakCachedManager struct {
	Manager
	client   Client
	cache    TokenCache
	admToken *user.IdentityToken
	log      zerolog.Logger
	mu       sync.Mutex
}

func NewCachedManager(client Client, cache TokenCache, log zerolog.Logger) *KeycloakCachedManager {
	return &KeycloakCachedManager{
		client: client,
		cache:  cache,
		log:    log,
	}
}

func KeycloakPGManager(cfg *config.Config, db *pg.DB, log zerolog.Logger) (*KeycloakCachedManager, error) {
	client, err := keycloak.NewClient(cfg, log)
	if err != nil {
		return nil, err
	}

	cache := NewPGIdentityTokenCache(db)

	return NewCachedManager(client, cache, log), nil
}

// adminToken ensures a valid cached admin token is available.
// It refreshes the token if it's close to expiring or absent.
func (im *KeycloakCachedManager) adminToken(ctx context.Context) (string, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	if im.admToken != nil && im.admToken.IsValid() {
		return im.admToken.AccessToken, nil
	}

	jwt, err := im.client.LoginClient(ctx)
	if err != nil {
		return "", err
	}

	im.admToken = user.TokenFromJWT(uuid.Nil, jwt)

	return im.admToken.AccessToken, nil
}

// RefreshToken exchanges a valid refresh token for a new JWT token.
// It returns a refreshed access and refresh token pair, or an error on failure.
func (im *KeycloakCachedManager) RefreshToken(
	ctx context.Context,
	token *user.IdentityToken,
) (*user.IdentityToken, error) {
	jwt, err := im.client.RefreshToken(ctx, token.RefreshToken)
	if err != nil {
		return nil, err
	}

	freshToken := user.TokenFromJWT(token.UserID, jwt)

	if err := im.cache.Upsert(ctx, freshToken); err != nil {
		im.log.Info().Err(err).
			Str("user_id", token.UserID.String()).
			Msg("failed to store/update identity token in cache")
	}

	return freshToken, nil
}

// GetToken authenticates a user using their username and password.
// It returns a JWT containing access and refresh tokens, or an error on failure.
func (im *KeycloakCachedManager) GetToken(ctx context.Context, usr *user.User) (*user.IdentityToken, error) {
	cachedToken, err := im.cache.Get(ctx, usr.ID)
	if err == nil && cachedToken.IsValid() {
		im.log.Info().
			Str("user_id", usr.ID.String()).
			Msg("got identity token from cache!")

		return cachedToken, nil
	}

	if err != nil {
		im.log.Info().Err(err).
			Str("user_id", usr.ID.String()).
			Msg("failed to get identity token from cache")
	}

	jwt, err := im.client.Login(ctx, usr)
	if err != nil {
		return nil, err
	}

	token := user.TokenFromJWT(usr.ID, jwt)

	if err := im.cache.Upsert(ctx, token); err != nil {
		im.log.Info().Err(err).
			Str("user_id", usr.ID.String()).
			Msg("failed to store/update identity token in cache")
	}

	return token, nil
}

// CreateUser registers a new user in the identity provider with the given credentials.
// It uses an admin token to authorize the operation and returns the new user's ID.
func (im *KeycloakCachedManager) CreateUser(ctx context.Context, usr *user.User) (string, error) {
	token, err := im.adminToken(ctx)
	if err != nil {
		return "", err
	}

	usrID, err := im.client.CreateUser(ctx, usr, token)
	if err != nil {
		return "", err
	}

	_, err = im.GetToken(ctx, usr)
	if err != nil {
		return "", err
	}

	return usrID, nil
}

// DeleteUser removes a user by ID from the identity provider.
// It uses a cached or freshly obtained admin token to authorize the deletion.
func (im *KeycloakCachedManager) DeleteUser(ctx context.Context, usr *user.User) error {
	if err := im.cache.Delete(ctx, usr.ID); err != nil {
		im.log.Info().Err(err).
			Str("user_id", usr.ID.String()).
			Msg("failed to delete identity token from cache")
	}

	token, err := im.adminToken(ctx)
	if err != nil {
		return err
	}

	return im.client.DeleteUser(ctx, usr.IdentityID, token)
}
