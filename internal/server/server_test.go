package server_test

import (
	"context"
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/patraden/ya-practicum-gophkeeper/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/internal/logger"
	"github.com/patraden/ya-practicum-gophkeeper/internal/mock"
	pb "github.com/patraden/ya-practicum-gophkeeper/internal/proto/gophkeeper/v1"
	"github.com/patraden/ya-practicum-gophkeeper/internal/server"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/utils/certgen"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func generateTestCertificates(
	t *testing.T,
	dir string,
	log *zerolog.Logger,
) (string, string, string, string) {
	t.Helper()

	caCertPath := filepath.Join(dir, "ca-public.crt")
	caKeyPath := filepath.Join(dir, "ca-private.key")
	serverCertPath := filepath.Join(dir, "server.crt")
	serverKeyPath := filepath.Join(dir, "server.key")

	// Generate CA cert
	require.NoError(t, certgen.GenerateCertificate(certgen.Config{
		OrgName:    "TestCA",
		CommonName: "TestCA",
		IsCA:       true,
		ECDSACurve: "P256",
		ValidFrom:  time.Now(),
		ValidFor:   365 * 24 * time.Hour,
		Host:       "localhost",
	}, log))

	require.NoError(t, os.Rename("ca-public.crt", caCertPath))
	require.NoError(t, os.Rename("ca-private.key", caKeyPath))

	// Generate server cert signed by CA
	require.NoError(t, certgen.GenerateCertificate(certgen.Config{
		OrgName:    "TestServer",
		CommonName: "localhost",
		ECDSACurve: "P256",
		CACertPath: caCertPath,
		CAKeyPath:  caKeyPath,
		Host:       "localhost,127.0.0.1",
		ValidFrom:  time.Now(),
		ValidFor:   365 * 24 * time.Hour,
	}, log))

	require.NoError(t, os.Rename("public.crt", serverCertPath))
	require.NoError(t, os.Rename("private.key", serverKeyPath))

	return caCertPath, caKeyPath, serverCertPath, serverKeyPath
}

func TestGRPCServerWithTLS(t *testing.T) {
	t.Parallel()

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()
	tmpDir := t.TempDir()

	caCertPath, _, serverCertPath, serverKeyPath := generateTestCertificates(t, tmpDir, log)

	cfg := &config.Config{
		ServerAddr:        "127.0.0.1:50055",
		ServerTLSCertPath: serverCertPath,
		ServerTLSKeyPath:  serverKeyPath,
	}

	ctrl := gomock.NewController(t)
	adminSrv := mock.NewMockAdminServiceServer(ctrl)
	userSrv := mock.NewMockUserServiceServer(ctrl)

	adminSrv.EXPECT().
		Unseal(gomock.Any(), gomock.Any()).
		Return(&pb.UnsealResponse{
			Message:  "success",
			Unsealed: true,
			Status:   pb.SealStatus_SEAL_STATUS_UNSEALED,
		}, nil)

	server, err := server.New(cfg, adminSrv, userSrv, log)
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
	require.True(t, resp.GetUnsealed())
	require.Equal(t, pb.SealStatus_SEAL_STATUS_UNSEALED, resp.GetStatus())

	err = server.Shutdown(ctx)
	require.NoError(t, err)
	require.NoError(t, <-runErrCh)
}
