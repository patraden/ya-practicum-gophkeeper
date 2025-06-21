package identity

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
)

// KeycloakManager is a concrete implementation of IdentityManager.
// It manages user accounts and token lifecycles using a Keycloak-compatible identity provider.
type KeycloakManager struct {
	client   Client
	admToken *user.IdentityToken
	mu       sync.Mutex
}

func NewKeycloakManager(client Client) *KeycloakManager {
	return &KeycloakManager{
		client: client,
	}
}

// adminToken ensures a valid cached admin token is available.
// It refreshes the token if it's close to expiring or absent.
func (im *KeycloakManager) adminToken(ctx context.Context) (string, error) {
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
func (im *KeycloakManager) RefreshToken(ctx context.Context, token *user.IdentityToken) (*user.IdentityToken, error) {
	jwt, err := im.client.RefreshToken(ctx, token.RefreshToken)
	if err != nil {
		return nil, err
	}

	freshToken := user.TokenFromJWT(token.UserID, jwt)

	return freshToken, nil
}

// GetToken authenticates a user using their username and password.
// It returns a JWT containing access and refresh tokens, or an error on failure.
func (im *KeycloakManager) GetToken(ctx context.Context, usr *user.User) (*user.IdentityToken, error) {
	jwt, err := im.client.Login(ctx, usr)
	if err != nil {
		return nil, err
	}

	token := user.TokenFromJWT(usr.ID, jwt)

	return token, nil
}

// CreateUser registers a new user in the identity provider with the given credentials.
// It uses an admin token to authorize the operation and returns the new user's ID.
func (im *KeycloakManager) CreateUser(ctx context.Context, usr *user.User) (string, error) {
	token, err := im.adminToken(ctx)
	if err != nil {
		return "", err
	}

	return im.client.CreateUser(ctx, usr, token)
}

// DeleteUser removes a user by ID from the identity provider.
// It uses a cached or freshly obtained admin token to authorize the deletion.
func (im *KeycloakManager) DeleteUser(ctx context.Context, usr *user.User) error {
	token, err := im.adminToken(ctx)
	if err != nil {
		return err
	}

	return im.client.DeleteUser(ctx, usr.IdentityID, token)
}
