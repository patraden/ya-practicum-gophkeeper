package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

// TokenExtractor extracts jwt token from context.
type TokenExtractor func(ctx context.Context) string

// MetaDataTokenExtractor extracts the JWT token from a metadat in the GRPC request.
func MetaDataTokenExtractor(ctx context.Context) string {
	md, exists := metadata.FromIncomingContext(ctx)
	if !exists {
		return ""
	}

	authHeader, exists := md["authorization"]
	if !exists || len(authHeader) == 0 {
		return ""
	}

	tokenString := authHeader[0]
	if !strings.HasPrefix(tokenString, "Bearer ") {
		return ""
	}

	return strings.TrimPrefix(tokenString, "Bearer ")
}
