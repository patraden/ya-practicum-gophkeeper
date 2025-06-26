package grpchandler

import (
	"context"
	"encoding/base64"
	"strings"

	pb "github.com/patraden/ya-practicum-gophkeeper/pkg/proto/gophkeeper/v1"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/app"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/crypto/shamir"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdminServer struct {
	usecase app.AdminUseCase
	config  *config.Config
	log     zerolog.Logger
	pb.UnimplementedAdminServiceServer
}

func NewAdminServer(config *config.Config, usecase app.AdminUseCase, log zerolog.Logger) *AdminServer {
	return &AdminServer{
		usecase: usecase,
		config:  config,
		log:     log,
	}
}

func (s *AdminServer) Unseal(ctx context.Context, req *pb.UnsealRequest) (*pb.UnsealResponse, error) {
	if err := req.Validate(); err != nil {
		s.log.Error().Err(err).
			Str("operation", "Unseal").
			Msg("invalid grpc request")

		return nil, status.Error(codes.InvalidArgument, "Bad Request: invalid params")
	}

	encoded := strings.TrimSpace(req.GetKeyPiece())

	share, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		s.log.Error().
			Err(err).
			Str("operation", "Unseal").
			Msg("failed to decode base64 key piece")

		return nil, status.Errorf(codes.InvalidArgument, "Bad Request: invalid key piece encoding: %v", err)
	}

	if len(share) != shamir.ShareLength {
		s.log.Error().
			Str("operation", "Unseal").
			Int("decoded_length", len(share)).
			Msg("decoded share length mismatch")

		return nil, status.Error(codes.InvalidArgument, "Bad Request: invalid key piece length")
	}

	s.log.Info().
		Str("key_piece_prefix", encoded[:8]).
		Msg("attempting unseal with provided key piece")

	statusCode, message := s.usecase.Unseal(ctx, share)

	s.log.Info().
		Str("operation", "Unseal").
		Int32("status", int32(statusCode)).
		Str("message", message).
		Msg("responding with seal status")

	return &pb.UnsealResponse{
		Status:  statusCode,
		Message: message,
	}, nil
}
