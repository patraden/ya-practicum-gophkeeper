package grpchandler

import (
	"context"
	"errors"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/dto"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	pb "github.com/patraden/ya-practicum-gophkeeper/pkg/proto/gophkeeper/v1"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/app"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/auth"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServer struct {
	config *config.Config
	auth   *auth.Auth
	app    app.UserUseCase
	log    *zerolog.Logger
	pb.UnimplementedUserServiceServer
}

func NewUserServer(config *config.Config, auth *auth.Auth, app app.UserUseCase, log *zerolog.Logger) *UserServer {
	return &UserServer{
		config: config,
		auth:   auth,
		app:    app,
		log:    log,
	}
}

func (s *UserServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if err := req.Validate(); err != nil {
		s.log.Error().Err(err).
			Str("operation", "Login").
			Msg("invalid grpc request")

		return nil, status.Error(codes.InvalidArgument, "Bad Request: invalid params")
	}

	creds := &dto.UserCredentials{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
	}

	usr, err := s.app.ValidateUser(ctx, creds)
	if errors.Is(err, e.ErrValidation) {
		return nil, status.Error(codes.Internal, "Unauthorized: invalid user credentials")
	}

	if errors.Is(err, e.ErrNotFound) {
		return nil, status.Error(codes.Internal, "Unauthorized: user not found")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, "Internal Server Error: user validation")
	}

	tokenEnc := s.auth.Encoder()

	token, err := tokenEnc(usr)
	if err != nil {
		s.log.Error().Err(err).
			Msg("failed to generate token")

		return nil, status.Error(codes.Internal, "Internal Server Error: token creation")
	}

	if err := auth.StoreTokenInGRPCHeader(ctx, token, s.log); err != nil {
		s.log.Error().Err(err).
			Msg("failed to inject token to headers")

		return nil, status.Error(codes.Internal, "Internal Server Error: token injection")
	}

	return &pb.LoginResponse{
		UserId: usr.ID.String(),
		Token:  token,
		Role:   usr.Role,
	}, nil
}

func (s *UserServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if err := req.Validate(); err != nil {
		s.log.Error().Err(err).
			Str("operation", "Login").
			Msg("invalid grpc request")

		return nil, status.Error(codes.InvalidArgument, "Bad Request: invalid params")
	}

	creds := &dto.RegisterUserCredentials{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		Role:     req.GetRole(),
	}

	usr, err := s.app.RegisterUser(ctx, creds)
	if errors.Is(err, e.ErrExists) {
		return nil, status.Error(codes.AlreadyExists, "User exists")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, "Internal Server Error: user registration")
	}

	tokenEnc := s.auth.Encoder()

	token, err := tokenEnc(usr)
	if err != nil {
		s.log.Error().Err(err).
			Msg("failed to generate token")

		return nil, status.Error(codes.Internal, "Internal Server Error: token creation")
	}

	if err := auth.StoreTokenInGRPCHeader(ctx, token, s.log); err != nil {
		s.log.Error().Err(err).
			Msg("failed to inject token to headers")

		return nil, status.Error(codes.Internal, "Internal Server Error: token injection")
	}

	return &pb.RegisterResponse{
		UserId: usr.ID.String(),
		Token:  token,
		Role:   usr.Role,
	}, nil
}
