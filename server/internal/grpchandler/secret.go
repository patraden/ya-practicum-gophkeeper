package grpchandler

import (
	"context"
	"errors"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/dto"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	pb "github.com/patraden/ya-practicum-gophkeeper/pkg/proto/gophkeeper/v1"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/app"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SecretServer struct {
	app    app.SecretUseCase
	config *config.Config
	log    zerolog.Logger
	pb.UnimplementedSecretServiceServer
}

func NewSecretServer(config *config.Config, app app.SecretUseCase, log zerolog.Logger) *SecretServer {
	return &SecretServer{
		config: config,
		app:    app,
		log:    log,
	}
}

func (s *SecretServer) SecretUpdateInit(
	ctx context.Context,
	req *pb.SecretUpdateInitRequest,
) (*pb.SecretUpdateInitResponse, error) {
	if req == nil {
		return nil, status.Error(codes.Internal, "nil request")
	}

	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	initReq, err := dto.SecretUploadInitRequestFromProto(req).ToDomain()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.app.InitUploadRequest(ctx, initReq)
	if errors.Is(err, e.ErrNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if errors.Is(err, e.ErrExists) {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}

	if errors.Is(err, e.ErrInvalidInput) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp.ToProto(), nil
}
