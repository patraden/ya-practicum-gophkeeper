package auth

import (
	"context"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func StoreTokenInGRPCHeader(ctx context.Context, token string, log zerolog.Logger) error {
	err := grpc.SetHeader(ctx, metadata.Pairs("authorization", "Bearer "+token))
	if err != nil {
		log.Error().Err(err).
			Msg("failed to set response metadata")

		return status.Errorf(codes.Internal, "Internal Server Error")
	}

	return nil
}
