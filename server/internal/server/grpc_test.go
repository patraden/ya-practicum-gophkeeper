//nolint:funlen // reason: testing internal logic and long test functions are acceptable
package server_test

import (
	"context"
	"crypto/x509"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	pb "github.com/patraden/ya-practicum-gophkeeper/pkg/proto/gophkeeper/v1"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/testutil/certtest"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/auth"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/mock"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/server"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestGRPCServerWithTLS(t *testing.T) {
	t.Parallel()

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()
	tmpDir := t.TempDir()

	caCertPath, serverCertPath, serverKeyPath := certtest.GenerateTestCertificates(t, tmpDir, log)

	cfg := &config.Config{
		ServerAddr:        "127.0.0.1:50055",
		ServerTLSCertPath: serverCertPath,
		ServerTLSKeyPath:  serverKeyPath,
		JWTSecret:         "secret",
	}

	jwtKeyFunc := func(*jwt.Token) (any, error) { return []byte(cfg.JWTSecret), nil }
	authenticator := auth.New(jwtKeyFunc, log)
	isPublicMethod := func(method string) bool { return method == pb.AdminService_Unseal_FullMethodName }

	ctrl := gomock.NewController(t)
	adminSrv := mock.NewMockAdminServiceServer(ctrl)
	userSrv := mock.NewMockUserServiceServer(ctrl)

	adminSrv.EXPECT().
		Unseal(gomock.Any(), gomock.Any()).
		Return(&pb.UnsealResponse{
			Message: "success",
			Status:  pb.SealStatus_SEAL_STATUS_UNSEALED,
		}, nil)

	server, err := server.New(cfg, adminSrv, userSrv, authenticator, isPublicMethod, log)
	require.NoError(t, err)

	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- server.Run()
	}()

	caCert, err := os.ReadFile(caCertPath)
	require.NoError(t, err)

	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(caCert)
	require.True(t, ok)

	creds := credentials.NewClientTLSFromCert(certPool, "localhost")
	conn, err := grpc.NewClient(cfg.ServerAddr, grpc.WithTransportCredentials(creds))
	require.NoError(t, err)

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client := pb.NewAdminServiceClient(conn)
	resp, err := client.Unseal(ctx, &pb.UnsealRequest{})
	require.NoError(t, err)

	require.Equal(t, "success", resp.GetMessage())
	require.Equal(t, pb.SealStatus_SEAL_STATUS_UNSEALED, resp.GetStatus())

	err = server.Shutdown(ctx)
	require.NoError(t, err)
	require.NoError(t, <-runErrCh)
}
