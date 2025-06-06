package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/auth"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

const (
	jwtSecret = "secret"
	userName  = "testuser"
	userIDStr = "123e4567-e89b-12d3-a456-426614174000"
)

func setupTestUsers(t *testing.T) (*user.User, *user.User) {
	t.Helper()

	usr, err := user.NewWithID(userIDStr, userName, user.RoleUser)
	require.NoError(t, err)

	usrNil, err := user.NewWithID(uuid.Nil.String(), "", user.RoleUser)
	require.NoError(t, err)

	return usr, usrNil
}

func setupLogger(t *testing.T) *zerolog.Logger {
	t.Helper()

	return logger.Stdout(zerolog.DebugLevel).GetZeroLog()
}

func mockKeyFunc(_ *jwt.Token) (any, error) {
	return []byte(jwtSecret), nil
}

func ExpiredToken(t *testing.T) (string, error) {
	t.Helper()

	now := time.Now()
	userID, err := uuid.Parse(userIDStr)
	require.NoError(t, err)

	claims := auth.Claims{
		UserID:   userID.String(),
		Username: userName,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return ``, e.ErrInvalidInput
	}

	return signedToken, nil
}

func TestAuthEncoder(t *testing.T) {
	t.Parallel()

	logger := setupLogger(t)
	usr, _ := setupTestUsers(t)

	type testCase struct {
		name        string
		keyFunc     jwt.Keyfunc
		user        *user.User
		expectErr   error
		expectToken bool
	}

	tests := []testCase{
		{
			name: "successful token generation",
			keyFunc: func(_ *jwt.Token) (any, error) {
				return []byte("secret"), nil
			},
			user:        usr,
			expectErr:   nil,
			expectToken: true,
		},
		{
			name: "keyFunc returns error",
			keyFunc: func(_ *jwt.Token) (any, error) {
				return nil, e.ErrGenerate
			},
			user:        usr,
			expectErr:   e.ErrGenerate,
			expectToken: false,
		},
		{
			name: "signedString fails due to bad key type",
			keyFunc: func(_ *jwt.Token) (any, error) {
				return struct{}{}, nil // incompatible key type
			},
			user:        usr,
			expectErr:   e.ErrGenerate,
			expectToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			jwtauth := auth.New(tt.keyFunc, logger)
			encoder := jwtauth.Encoder()

			token, err := encoder(tt.user)

			if tt.expectErr != nil {
				require.ErrorIs(t, err, tt.expectErr)
				require.Empty(t, token)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, token)
			}
		})
	}
}

func TestAuthVerifyValid(t *testing.T) {
	t.Parallel()

	logger := setupLogger(t)
	usr, _ := setupTestUsers(t)
	jwtauth := auth.New(mockKeyFunc, logger)
	encoder := jwtauth.Encoder()

	tokenString, err := encoder(usr)
	require.NoError(t, err)

	token, err := jwtauth.Validate(tokenString)
	require.NoError(t, err, "Token verification should succeed")
	assert.NotNil(t, token, "Token should not be nil")
	assert.True(t, token.Valid, "Token should be valid")

	claims, ok := token.Claims.(*auth.Claims)
	assert.True(t, ok, "Claims should be of type *auth.Claims")
	assert.Equal(t, userName, claims.Username, "Username in claims should match")
	assert.Equal(t, usr.ID.String(), claims.UserID, "User id in claims should match")
}

func TestAuthVerifyInvalid(t *testing.T) {
	t.Parallel()

	logger := setupLogger(t)
	usr, usrNil := setupTestUsers(t)
	jwtauth := auth.New(mockKeyFunc, logger)

	token, err := jwtauth.Validate("invalid.token")
	require.ErrorIs(t, err, e.ErrInvalidInput, "Verification of an invalid token should error")
	assert.Nil(t, token)

	encoder := jwtauth.Encoder()
	tokenString, err := encoder(usrNil)
	require.NoError(t, err)

	token, err = jwtauth.Validate(tokenString)
	require.ErrorIs(t, err, e.ErrUnauthenticated, "Verification of a token with nil user_id should error")
	assert.Nil(t, token)

	expiredToken, err := ExpiredToken(t)
	require.NoError(t, err)
	token, err = jwtauth.Validate(expiredToken)
	require.ErrorIs(t, err, e.ErrInvalidInput, "Verification of an expired token should error")
	assert.Nil(t, token)

	// Create token with wrong key
	tempAuth := auth.New(func(_ *jwt.Token) (any, error) { return []byte("wrong_secret"), nil }, logger)
	encoder = tempAuth.Encoder()
	tokenString, err = encoder(usr)
	require.NoError(t, err)

	token, err = jwtauth.Validate(tokenString)
	require.ErrorIs(t, err, e.ErrInvalidInput, "Verification of a token with bad secret should error")
	assert.Nil(t, token)
}

func TestAuthVerifyContext(t *testing.T) {
	t.Parallel()

	logger := setupLogger(t)
	usr, _ := setupTestUsers(t)
	jwtauth := auth.New(mockKeyFunc, logger)
	encoder := jwtauth.Encoder()
	validToken, err := encoder(usr)

	require.NoError(t, err)

	tests := []struct {
		name        string
		metadata    map[string]string
		expectToken bool
		expectErr   error
	}{
		{"valid token", map[string]string{"authorization": "Bearer " + validToken}, true, nil},
		{"missing metadata", nil, false, e.ErrNotFound},
		{"no header", map[string]string{"x-custom-header": "value"}, false, e.ErrNotFound},
		{"header is empty", map[string]string{"authorization": ""}, false, e.ErrNotFound},
		{"no Bearer prefix", map[string]string{"authorization": "Token " + validToken}, false, e.ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var ctx context.Context
			if tt.metadata != nil {
				ctx = metadata.NewIncomingContext(context.Background(), metadata.New(tt.metadata))
			} else {
				ctx = context.Background()
			}

			token, err := jwtauth.VerifyContext(ctx, auth.MetaDataTokenExtractor)

			if tt.expectToken {
				require.NoError(t, err)
				require.NotNil(t, token)
			} else {
				require.ErrorIs(t, err, tt.expectErr)
				require.Nil(t, token)
			}
		})
	}
}
