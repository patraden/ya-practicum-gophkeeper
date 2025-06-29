package keystore

import (
	"context"

	pb "github.com/patraden/ya-practicum-gophkeeper/pkg/proto/gophkeeper/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServerStatusValidator returns a gRPC unary interceptor that blocks all requests
// if the keystore is sealed (i.e., not yet loaded). This ensures that secrets are not
// accessible until the keystore is explicitly unsealed.
//
// It makes an exception for specific methods such as AdminService.Unseal and
// UserService.Login, which are allowed even when the keystore is sealed.
// All other RPCs will return a gRPC Unavailable error until the keystore is unsealed.
func GRPCServerStatusValidator(kstore Keystore) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		switch info.FullMethod {
		case
			pb.AdminService_Unseal_FullMethodName,
			pb.UserService_Login_FullMethodName:
			return handler(ctx, req)
		}

		if kstore.IsLoaded() {
			return handler(ctx, req)
		}

		return nil, status.Errorf(codes.Unavailable, "server is sealed")
	}
}
