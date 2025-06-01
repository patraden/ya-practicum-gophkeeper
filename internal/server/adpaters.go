//nolint:wrapcheck // reason: error wrapcheck for adapters is excessive
package server

import (
	"context"

	pb "github.com/patraden/ya-practicum-gophkeeper/internal/proto/gophkeeper/v1"
)

type AdminServiceServer interface {
	Unseal(ctx context.Context, r *pb.UnsealRequest) (*pb.UnsealResponse, error)
}

type UserServiceServer interface {
	Login(ctx context.Context, r *pb.LoginRequest) (*pb.LoginResponse, error)
	Register(ctx context.Context, r *pb.RegisterRequest) (*pb.RegisterResponse, error)
}

type AdminServiceAdapter struct {
	impl AdminServiceServer
	pb.UnimplementedAdminServiceServer
}

func (a *AdminServiceAdapter) Unseal(ctx context.Context, req *pb.UnsealRequest) (*pb.UnsealResponse, error) {
	return a.impl.Unseal(ctx, req)
}

type UserServiceAdapter struct {
	impl UserServiceServer
	pb.UnimplementedUserServiceServer
}

func (u *UserServiceAdapter) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return u.impl.Login(ctx, req)
}

func (u *UserServiceAdapter) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return u.impl.Register(ctx, req)
}
