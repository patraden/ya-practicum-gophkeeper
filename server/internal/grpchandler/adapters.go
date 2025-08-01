package grpchandler

import (
	"context"

	pb "github.com/patraden/ya-practicum-gophkeeper/pkg/proto/gophkeeper/v1"
)

type AdminServiceServer interface {
	Unseal(ctx context.Context, r *pb.UnsealRequest) (*pb.UnsealResponse, error)
}

type UserServiceServer interface {
	Login(ctx context.Context, r *pb.LoginRequest) (*pb.LoginResponse, error)
	Register(ctx context.Context, r *pb.RegisterRequest) (*pb.RegisterResponse, error)
}

type SecretServiceServer interface {
	SecretUpdateInit(ctx context.Context, req *pb.SecretUpdateInitRequest) (*pb.SecretUpdateInitResponse, error)
	SecretUpdateCommit(ctx context.Context, req *pb.SecretUpdateCommitRequest) (*pb.SecretUpdateCommitResponse, error)
}

type AdminServiceAdapter struct {
	impl AdminServiceServer
	pb.UnimplementedAdminServiceServer
}

func NewAdminServiceAdapter(impl AdminServiceServer) *AdminServiceAdapter {
	return &AdminServiceAdapter{
		impl: impl,
	}
}

func (a *AdminServiceAdapter) Unseal(ctx context.Context, req *pb.UnsealRequest) (*pb.UnsealResponse, error) {
	return a.impl.Unseal(ctx, req)
}

type UserServiceAdapter struct {
	impl UserServiceServer
	pb.UnimplementedUserServiceServer
}

func NewUserServiceAdapter(impl UserServiceServer) *UserServiceAdapter {
	return &UserServiceAdapter{
		impl: impl,
	}
}

func (u *UserServiceAdapter) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return u.impl.Login(ctx, req)
}

func (u *UserServiceAdapter) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return u.impl.Register(ctx, req)
}

type SecretServiceAdapter struct {
	impl SecretServiceServer
	pb.UnimplementedSecretServiceServer
}

func NewSecretServiceAdapter(impl SecretServiceServer) *SecretServiceAdapter {
	return &SecretServiceAdapter{
		impl: impl,
	}
}

func (s *SecretServiceAdapter) SecretUpdateInit(
	ctx context.Context,
	req *pb.SecretUpdateInitRequest,
) (*pb.SecretUpdateInitResponse, error) {
	return s.impl.SecretUpdateInit(ctx, req)
}

func (s *SecretServiceAdapter) SecretUpdateCommit(
	ctx context.Context,
	req *pb.SecretUpdateCommitRequest,
) (*pb.SecretUpdateCommitResponse, error) {
	return s.impl.SecretUpdateCommit(ctx, req)
}
