package identity_test

import (
	"context"
	"testing"
	"time"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/identity"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/keycloak"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestKeycloakManagerIntegration(t *testing.T) {
	t.Parallel()
	t.Skip("Disabled during development; enable when Keycloak container is running")

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()
	cfg := config.DefaultConfig()
	usr := user.New("test_user_manager2", user.RoleUser)
	client, err := keycloak.NewClient(cfg, log)
	require.NoError(t, err)

	cache := identity.NewInMemoryIdentityTokenCache()
	manager := identity.NewCachedManager(client, cache, log)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	iuid, err := manager.CreateUser(ctx, usr)
	require.NoError(t, err)

	usr.SetIdentityID(iuid)

	token, err := manager.GetToken(ctx, usr)
	require.NoError(t, err)
	require.NotNil(t, token)

	freshToken, err := manager.RefreshToken(ctx, token)
	require.NoError(t, err)
	require.NotNil(t, freshToken)

	log.Info().
		Str("user_id", usr.ID.String()).
		Str("identity_id", usr.IdentityID).
		Str("identity_pass", usr.IdentityPassword()).
		Msg("test user")

	err = manager.DeleteUser(ctx, usr)
	require.NoError(t, err)
}
