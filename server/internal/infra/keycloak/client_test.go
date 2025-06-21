package keycloak_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-resty/resty/v2"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/testutil/http/roundtrip"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/keycloak"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T, httpClient *http.Client) *keycloak.Client {
	t.Helper()

	cfg := &config.Config{
		IdentityTLSCertPath:  "",
		IdentityEndpoint:     "http://fake-keycloak",
		IdentityRealm:        "test-realm",
		IdentityClientID:     "test-client",
		IdentityClientSecret: "test-secret",
	}

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()
	client, err := keycloak.NewClient(cfg, log)
	require.NoError(t, err)

	restyClient := resty.NewWithClient(httpClient)
	client.SetRestyClient(restyClient)

	return client
}

func TestLoginClientSuccess(t *testing.T) {
	t.Parallel()

	mockResponse := `{
		"access_token": "mocked-access-token",
		"expires_in": 3600,
		"refresh_expires_in": 1800,
		"token_type": "Bearer",
		"not-before-policy": 0,
		"scope": "profile"
	}`

	httpClient := roundtrip.NewTestHTTPClient(func(req *http.Request) *http.Response {
		assert.Equal(t, http.MethodPost, req.Method)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}
	})

	client := newTestClient(t, httpClient)
	token, err := client.LoginClient(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "mocked-access-token", token.AccessToken)
}

func TestLoginClientFailure(t *testing.T) {
	t.Parallel()

	httpClient := roundtrip.NewTestHTTPClient(func(_ *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       io.NopCloser(bytes.NewBufferString("unauthorized")),
		}
	})

	client := newTestClient(t, httpClient)
	token, err := client.LoginClient(context.Background())
	require.Error(t, err)
	require.Nil(t, token)
}

func TestNewClientUnsecured(t *testing.T) {
	t.Parallel()

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()
	cfg := &config.Config{
		IdentityTLSCertPath:  "",
		IdentityEndpoint:     "http://keycloak",
		IdentityRealm:        "realm",
		IdentityClientID:     "id",
		IdentityClientSecret: "secret",
	}

	client, err := keycloak.NewClient(cfg, log)
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestLoginSuccess(t *testing.T) {
	t.Parallel()

	mockResponse := `{"access_token": "user-token", "expires_in": 3600}`
	httpClient := roundtrip.NewTestHTTPClient(func(_ *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}
	})

	usr := user.New("user", user.RoleUser)
	client := newTestClient(t, httpClient)
	token, err := client.Login(context.Background(), usr)
	require.NoError(t, err)
	assert.Equal(t, "user-token", token.AccessToken)
}

func TestRefreshTokenSuccess(t *testing.T) {
	t.Parallel()

	mockResponse := `{"access_token": "refreshed-token", "expires_in": 3600}`
	httpClient := roundtrip.NewTestHTTPClient(func(_ *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}
	})

	client := newTestClient(t, httpClient)
	token, err := client.RefreshToken(context.Background(), "some-refresh-token")
	require.NoError(t, err)
	assert.Equal(t, "refreshed-token", token.AccessToken)
}

func TestCreateUserUserExists(t *testing.T) {
	t.Parallel()

	existingUsers, err := json.Marshal([]gocloak.User{{Username: gocloak.StringP("existing-user")}})
	require.NoError(t, err)

	httpClient := roundtrip.NewTestHTTPClient(func(req *http.Request) *http.Response {
		if req.URL.Path == "/admin/realms/test-realm/users" && req.Method == http.MethodGet {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(existingUsers)),
			}
		}

		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString("unexpected")),
		}
	})

	usr := user.New("existing-user", user.RoleUser)
	client := newTestClient(t, httpClient)
	userID, err := client.CreateUser(context.Background(), usr, "admin-token")
	require.Error(t, err)
	assert.Equal(t, "", userID)
}

func TestDeleteUserSuccess(t *testing.T) {
	t.Parallel()

	httpClient := roundtrip.NewTestHTTPClient(func(req *http.Request) *http.Response {
		assert.Equal(t, http.MethodDelete, req.Method)

		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(bytes.NewBufferString("")),
		}
	})

	client := newTestClient(t, httpClient)
	err := client.DeleteUser(context.Background(), "some-id", "token")
	assert.NoError(t, err)
}

func TestClientIntegrationCreateDeleteUser(t *testing.T) {
	t.Parallel()
	t.Skip("Disabled during development; enable when Keycloak container is running")

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()
	cfg := config.DefaultConfig()

	client, err := keycloak.NewClient(cfg, log)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token, err := client.LoginClient(ctx, "minio-admin-api-access", "minio-authorization")
	require.NoError(t, err)

	usr := user.New("test_user", user.RoleUser)

	usrID, err := client.CreateUser(ctx, usr, token.AccessToken)
	require.NoError(t, err)
	require.NotEqual(t, "", usrID)

	usr.SetIdentityID(usrID)

	log.Info().
		Str("user_id", usr.ID.String()).
		Str("identity_id", usr.IdentityID).
		Str("identity_pass", usr.IdentityPassword()).
		Msg("test user")

	err = client.DeleteUser(ctx, usrID, token.AccessToken)
	require.NoError(t, err)
}
