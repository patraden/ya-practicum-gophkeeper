package server

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/patraden/ya-practicum-gophkeeper/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/internal/domain/errors"
	pb "github.com/patraden/ya-practicum-gophkeeper/internal/proto/gophkeeper/v1"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Server represents the gRPC server for gophkeeper application.
type Server struct {
	grpcSrv  *grpc.Server
	config   *config.Config
	adminSrv AdminServiceServer
	userSrv  UserServiceServer
	log      *zerolog.Logger
}

// New creates instance of the application gRPC server.
func New(
	config *config.Config,
	adminSrv AdminServiceServer,
	userSrv UserServiceServer,
	log *zerolog.Logger,
) (*Server, error) {
	cert, err := tls.LoadX509KeyPair(config.ServerTLSCertPath, config.ServerTLSKeyPath)
	if err != nil {
		log.Error().Err(err).
			Msg("gRPC failed to load tls keypair")

		return nil, errors.ErrServerTLS
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
	grpcSrv := grpc.NewServer(grpc.Creds(creds))

	return &Server{
		grpcSrv:  grpcSrv,
		config:   config,
		adminSrv: adminSrv,
		userSrv:  userSrv,
		log:      log,
	}, nil
}

// Run starts the application gRPC server.
func (s *Server) Run() error {
	s.log.Info().
		Str("server_address", s.config.ServerAddr).
		Msgf("Starting gRPC server")

	listen, err := net.Listen("tcp", s.config.ServerAddr)
	if err != nil {
		s.log.Error().Err(err).
			Str("server_address", s.config.ServerAddr).
			Msg("failed to listen tcp address")

		return errors.ErrServerStart
	}

	pb.RegisterUserServiceServer(s.grpcSrv, &UserServiceAdapter{impl: s.userSrv})
	pb.RegisterAdminServiceServer(s.grpcSrv, &AdminServiceAdapter{impl: s.adminSrv})

	if err := s.grpcSrv.Serve(listen); err != nil {
		s.log.Error().Err(err).
			Str("server_address", s.config.ServerAddr).
			Msg("failed to server gRPC server")

		return errors.ErrServerStart
	}

	return nil
}

// Shutdown stops the application gRPC server.
func (s *Server) Shutdown(ctx context.Context) error {
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

		return errors.ErrServerShutdown
	case <-stopped:
		s.log.Info().
			Str("server_address", s.config.ServerAddr).
			Msg("gRPC server stopped gracefully")

		return nil
	}
}
