package identity

import (
	"context"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
)

// IdentityManager defines the interface for interacting with the identity provider.
// It abstracts user management and token lifecycle operations.
type Manager interface {
	// CreateUser registers a new user with the given username and password.
	// Returns user ID and an error if the creation fails.
	CreateUser(ctx context.Context, usr *user.User) (string, error)

	// DeleteUser removes the user from the identity provider using the given user ID.
	// Returns an error if the deletion fails or the user does not exist.
	DeleteUser(ctx context.Context, usr *user.User) error

	// GetToken authenticates the user with the provided credentials and returns a JWT token.
	// The returned token includes both access and refresh tokens.
	// Returns an error if authentication fails.
	GetToken(ctx context.Context, usr *user.User) (*user.IdentityToken, error)

	// RefreshToken exchanges a valid refresh token for a new JWT token.
	// The new token includes updated access and refresh tokens.
	// Returns an error if the refresh fails.
	RefreshToken(ctx context.Context, token *user.IdentityToken) (*user.IdentityToken, error)
}

// IdentityClient defines the low-level interface for direct communication
// with the identity provider (e.g., Keycloak). This is typically used by
// Manager implementations to perform raw identity operations.
type Client interface {
	// LoginClient performs a client credentials login with optional scopes.
	LoginClient(ctx context.Context, scopes ...string) (*user.JWT, error)
	// RefreshToken exchanges a refresh token for a new JWT.
	RefreshToken(ctx context.Context, refreshToken string) (*user.JWT, error)
	// Login authenticates a user with a username and password.
	Login(ctx context.Context, user *user.User) (*user.JWT, error)
	// CreateUser registers a new user in the identity provider using a domain model and admin token.
	CreateUser(ctx context.Context, user *user.User, token string) (string, error)
	// DeleteUser removes a user by ID using an admin access token.
	DeleteUser(ctx context.Context, userID, token string) error
}
