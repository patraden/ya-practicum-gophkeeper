package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// contextKey is a value for use with context.WithValue. It's used as
	// a pointer so it fits in an interface{} without allocation. This technique
	// for defining context keys was copied from Go 1.7's new use of context in net/http.
	contextKey string
	// TokenEncoder encodes User into jwt token string.
	TokenEncoder func(*user.User) (string, error)
)

// JWT package constants.
const (
	TokenCtxKey      = contextKey("Token")
	ErrorCtxKey      = contextKey("TokenErr")
	maxTokenDuration = 365 * 24 * time.Hour
)

// FromContext gets token and claims from context added by GRPCServerVerifier interceptor.
func FromContext(ctx context.Context) (*jwt.Token, *Claims, error) {
	var claims *Claims

	token, _ := ctx.Value(TokenCtxKey).(*jwt.Token)
	err, _ := ctx.Value(ErrorCtxKey).(error)

	if token != nil {
		if c, ok := token.Claims.(*Claims); ok {
			claims = c
		} else {
			return token, nil, e.ErrInvalidInput
		}
	}

	return token, claims, err
}

// Auth is a struct that provides JWT-based authentication and authorization capabilities.
type Auth struct {
	keyFunc jwt.Keyfunc
	log     zerolog.Logger
}

// New creates a new object of type Auth.
func New(keyFunc jwt.Keyfunc, log zerolog.Logger) *Auth {
	return &Auth{
		keyFunc: keyFunc,
		log:     log,
	}
}

// Validate validates the JWT token tring and returns jwt.Token pointer if string is valid.
func (auth *Auth) Validate(tokenString string) (*jwt.Token, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, auth.keyFunc)
	if err != nil {
		auth.log.Error().Err(err).
			Msg("failed to parse JWT token")

		return nil, fmt.Errorf("[%w] auth jwt token", e.ErrInvalidInput)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || claims.Validate() != nil {
		auth.log.Error().
			Str("user_id", claims.UserID).
			Msg("invalid claims")

		return nil, fmt.Errorf("[%w] invalid auth jwt claims", e.ErrUnauthenticated)
	}

	return token, nil
}

// Encoder returns jwt TokenEncoder for User.
func (auth *Auth) Encoder() TokenEncoder {
	return func(user *user.User) (string, error) {
		now := time.Now()

		claims := &Claims{
			UserID:   user.ID.String(),
			Username: user.Username,
			Role:     user.Role.String(),
			RegisteredClaims: jwt.RegisteredClaims{
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(maxTokenDuration)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		signingKey, err := auth.keyFunc(token)
		if err != nil {
			auth.log.Error().Err(err).
				Str("user_id", user.ID.String()).
				Str("username", user.Username).
				Msg("failed to retrieve signing key")

			return "", fmt.Errorf("[%w] auth jwt signing key", e.ErrGenerate)
		}

		tokenString, err := token.SignedString(signingKey)
		if err != nil {
			auth.log.Error().Err(err).
				Str("user_id", user.ID.String()).
				Str("username", user.Username).
				Msg("failed to sign token")

			return "", fmt.Errorf("[%w] auth jwt token string", e.ErrGenerate)
		}

		auth.log.Info().
			Str("method", "HS256").
			Str("user_id", user.ID.String()).
			Str("username", user.Username).
			Msg("generated user token")

		return tokenString, nil
	}
}

// VerifyContext extracts token from request context using set of extractors.
func (auth *Auth) VerifyContext(ctx context.Context, extractors ...TokenExtractor) (*jwt.Token, error) {
	var tokenString string

	// Extract token string from the request by calling token find functions in
	// the order they where provided. Further extraction stops if a function
	// returns a non-empty string.
	for _, fn := range extractors {
		tokenString = fn(ctx)
		if tokenString != "" {
			break
		}
	}

	if tokenString == "" {
		auth.log.Info().
			Msg("token not found in request")

		return nil, fmt.Errorf("[%w] auth jwt token", e.ErrNotFound)
	}

	return auth.Validate(tokenString)
}

// VerifyGRPCUnaryServer return interceptor to extract and validate token string base on
// arbitrary set of context extractors. It produces and injects token and claims into context
// and unconditionally calls next handler.
func VerifyGRPCUnaryServer(auth *Auth, extractors ...TokenExtractor) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		token, err := auth.VerifyContext(ctx, extractors...)
		ctx = context.WithValue(ctx, TokenCtxKey, token)
		ctx = context.WithValue(ctx, ErrorCtxKey, err)

		return handler(ctx, req)
	}
}

// The GRPCServerVerifier always calls the next grpc server handler in sequence, which can either
// be the generic `Auth.GRPCServerAuthenticator` interceptor or any custom interceptor
// which checks the request context jwt token and error to prepare a custom response.
func GRPCServerVerifier(auth *Auth) grpc.UnaryServerInterceptor {
	return VerifyGRPCUnaryServer(auth, MetaDataTokenExtractor)
}

// GRPCServerAuthenticator is a default authentication interceptor to enforce access from the
// Verifier interceptor request context values. The GRPCServerAuthenticator sends a 401 Unauthorized
// response for any unverified tokens and passes the good ones through.
func GRPCServerAuthenticator(isPublicMethod func(method string) bool) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		token, claims, err := FromContext(ctx)
		if err != nil || token == nil || claims == nil {
			return nil, status.Errorf(codes.Unauthenticated, "Unauthorized")
		}

		return handler(ctx, req)
	}
}
