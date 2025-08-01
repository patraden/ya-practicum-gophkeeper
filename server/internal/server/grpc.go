package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	pb "github.com/patraden/ya-practicum-gophkeeper/pkg/proto/gophkeeper/v1"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/auth"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/crypto/keystore"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/grpchandler"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Server represents the gRPC server for gophkeeper application.
type GRPCServer struct {
	grpcSrv   *grpc.Server
	config    *config.Config
	adminSrv  grpchandler.AdminServiceServer
	userSrv   grpchandler.UserServiceServer
	secretSrv grpchandler.SecretServiceServer
	log       zerolog.Logger
}

// New creates instance of the application gRPC server.
func New(
	config *config.Config,
	adminSrv grpchandler.AdminServiceServer,
	userSrv grpchandler.UserServiceServer,
	secretSrv grpchandler.SecretServiceServer,
	authenticator *auth.Auth,
	kstore keystore.Keystore,
	isPublicMethod func(method string) bool,
	log zerolog.Logger,
) (*GRPCServer, error) {
	cert, err := tls.LoadX509KeyPair(config.ServerTLSCertPath, config.ServerTLSKeyPath)
	if err != nil {
		log.Error().Err(err).
			Msg("gRPC failed to load tls keypair")

		return nil, fmt.Errorf("[%w] gRPC tls keypair", e.ErrRead)
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
		PreferServerCipherSuites: true,
	}

	creds := credentials.NewTLS(tlsCfg)
	interceptors := grpc.ChainUnaryInterceptor(
		auth.GRPCServerVerifier(authenticator),
		auth.GRPCServerAuthenticator(isPublicMethod),
		keystore.GRPCServerStatusValidator(kstore),
	)

	grpcSrv := grpc.NewServer(
		grpc.Creds(creds),
		interceptors,
	)

	return &GRPCServer{
		grpcSrv:   grpcSrv,
		config:    config,
		adminSrv:  adminSrv,
		userSrv:   userSrv,
		secretSrv: secretSrv,
		log:       log,
	}, nil
}

// Run starts the application gRPC server.
func (s *GRPCServer) Run() error {
	s.log.Info().
		Str("server_address", s.config.ServerAddr).
		Msgf("Starting gRPC server")

	listen, err := net.Listen("tcp", s.config.ServerAddr)
	if err != nil {
		s.log.Error().Err(err).
			Str("server_address", s.config.ServerAddr).
			Msg("failed to listen tcp address")

		return fmt.Errorf("[%w] gRPC tcp address", e.ErrUnavailable)
	}

	pb.RegisterUserServiceServer(s.grpcSrv, grpchandler.NewUserServiceAdapter(s.userSrv))
	pb.RegisterAdminServiceServer(s.grpcSrv, grpchandler.NewAdminServiceAdapter(s.adminSrv))
	pb.RegisterSecretServiceServer(s.grpcSrv, grpchandler.NewSecretServiceAdapter(s.secretSrv))

	if err := s.grpcSrv.Serve(listen); err != nil {
		s.log.Error().Err(err).
			Str("server_address", s.config.ServerAddr).
			Msg("failed to serve gRPC server")

		return e.InternalErr(err)
	}

	return nil
}

// Shutdown stops the application gRPC server.
func (s *GRPCServer) Shutdown(ctx context.Context) error {
	stopped := make(chan struct{})

	go func() {
		s.grpcSrv.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		s.grpcSrv.Stop()
		s.log.Error().Err(ctx.Err()).
			Msg("Forced gRPC shutdown due to context cancel")

		return e.ErrTimeout
	case <-stopped:
		s.log.Info().
			Str("server_address", s.config.ServerAddr).
			Msg("gRPC server stopped gracefully")

		return nil
	}
}
