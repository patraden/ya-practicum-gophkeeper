// Package keycloak provides a wrapper around the GoCloak client for interacting
// with a Keycloak identity provider using custom domain models and enhanced logging.
package keycloak

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-resty/resty/v2"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/net/transport"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/rs/zerolog"
)

const (
	defaultUserMinioPolicy = "readwrite"
	defaultMinioUsersGroup = "minio_clients_group"
)

// Client wraps a GoCloak client with additional context-specific configuration and logging.
type Client struct {
	realm        string
	clientID     string
	clientSecret string
	client       *gocloak.GoCloak
	log          zerolog.Logger
}

// NewClient creates a Keycloak client. If `secured` is true, it loads TLS from cert path.
func NewClient(cfg *config.Config, log zerolog.Logger) (*Client, error) {
	client := gocloak.NewClient(cfg.IdentityEndpoint)
	secured := cfg.IdentityTLSCertPath != ""

	if secured {
		builder := transport.NewHTTPTransportBuilder(cfg.IdentityTLSCertPath, nil, log)

		httpTransport, err := builder.Build()
		if err != nil {
			return nil, err
		}

		restyClient := resty.NewWithClient(&http.Client{Transport: httpTransport})
		client.SetRestyClient(restyClient)
	}

	return &Client{
		realm:        cfg.IdentityRealm,
		clientID:     cfg.IdentityClientID,
		clientSecret: cfg.IdentityClientSecret,
		client:       client,
		log:          log,
	}, nil
}

// SetRestyClient overwrites the internal resty.
func (c *Client) SetRestyClient(restyClient *resty.Client) {
	c.client.SetRestyClient(restyClient)
}

// Login performs a login with user credentials and a client.
func (c *Client) Login(ctx context.Context, usr *user.User) (*user.JWT, error) {
	token, err := c.client.Login(ctx, c.clientID, c.clientSecret, c.realm, usr.Username, usr.IdentityPassword())
	if err != nil {
		c.log.Error().Err(err).
			Str("client_id", c.clientID).
			Str("realm", c.realm).
			Str("username", usr.Username).
			Msg("failed to login as user")

		return nil, e.InternalErr(err)
	}

	return token, nil
}

// LoginClient performs a login with client credentials.
func (c *Client) LoginClient(ctx context.Context, scopes ...string) (*user.JWT, error) {
	token, err := c.client.LoginClient(ctx, c.clientID, c.clientSecret, c.realm, scopes...)
	if err != nil {
		c.log.Error().Err(err).
			Str("client_id", c.clientID).
			Str("realm", c.realm).
			Msg("failed to login as client")

		return nil, fmt.Errorf("[%w] keyclock login client api", e.ErrUnavailable)
	}

	return token, nil
}

// RefreshToken exchanges a refresh token for a new JWT using the client credentials.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*user.JWT, error) {
	token, err := c.client.RefreshToken(ctx, refreshToken, c.clientID, c.clientSecret, c.realm)
	if err != nil {
		c.log.Error().Err(err).
			Str("client_id", c.clientID).
			Str("realm", c.realm).
			Msg("failed to refresh access token")

		return nil, fmt.Errorf("[%w] keyclock refresh token api", e.ErrUnavailable)
	}

	return token, nil
}

// CreateUser creates a new Keycloak user based on the provided domain user model.
// A default minio policy attribute `policy=readwrite` is also applied.
// Returns the created Keycloak user object.
//
//nolint:funlen //reason: logging
func (c *Client) CreateUser(ctx context.Context, usr *user.User, token string) (string, error) {
	gcUser := gocloak.User{
		Username:      gocloak.StringP(usr.Username),
		Enabled:       gocloak.BoolP(true),
		EmailVerified: gocloak.BoolP(true),
		Groups:        &[]string{defaultMinioUsersGroup},
		Attributes: &map[string][]string{
			"policy":    {defaultUserMinioPolicy},
			"appuserid": {usr.IDNoDash()},
		},
	}

	users, err := c.client.GetUsers(ctx, token, c.realm, gocloak.GetUsersParams{Username: gocloak.StringP(usr.Username)})
	if err != nil {
		c.log.Error().Err(err).
			Str("username", usr.Username).
			Msg("failed to fetch users")

		return "", e.InternalErr(err)
	}

	// manually filter by exact username match
	for _, u := range users {
		if u.Username != nil && *u.Username == usr.Username {
			c.log.Error().
				Str("client_id", c.clientID).
				Str("realm", c.realm).
				Str("username", usr.Username).
				Msg("user already exists")

			return "", fmt.Errorf("[%w] keycloak user", e.ErrExists)
		}
	}

	userID, err := c.client.CreateUser(ctx, token, c.realm, gcUser)
	if err != nil {
		c.log.Error().Err(err).
			Str("client_id", c.clientID).
			Str("realm", c.realm).
			Str("username", *gcUser.Username).
			Msg("failed to create user")

		return "", e.InternalErr(err)
	}

	err = c.client.SetPassword(ctx, token, userID, c.realm, usr.IdentityPassword(), false)
	if err != nil {
		c.log.Error().Err(err).
			Str("client_id", c.clientID).
			Str("realm", c.realm).
			Str("username", *gcUser.Username).
			Msg("failed to set user password")

		return "", e.InternalErr(err)
	}

	createdUser, err := c.client.GetUserByID(ctx, token, c.realm, userID)
	if err != nil {
		c.log.Error().Err(err).
			Str("client_id", c.clientID).
			Str("realm", c.realm).
			Str("username", *gcUser.Username).
			Msg("failed to fetch created user")

		return "", e.InternalErr(err)
	}

	return *createdUser.ID, nil
}

// DeleteUser removes the user with the given ID from Keycloak.
func (c *Client) DeleteUser(ctx context.Context, userID, token string) error {
	err := c.client.DeleteUser(ctx, token, c.realm, userID)
	if err != nil {
		c.log.Error().Err(err).
			Str("client_id", c.clientID).
			Str("realm", c.realm).
			Str("user_id", userID).
			Msg("failed to delete user")

		return e.InternalErr(err)
	}

	return nil
}
