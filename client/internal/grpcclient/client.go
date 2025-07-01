package grpcclient

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/patraden/ya-practicum-gophkeeper/client/internal/config"
	clientinfo "github.com/patraden/ya-practicum-gophkeeper/client/internal/systeminfo"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/dto"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	pb "github.com/patraden/ya-practicum-gophkeeper/pkg/proto/gophkeeper/v1"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Client wraps the gRPC connection and service clients.
type Client struct {
	Conn          *grpc.ClientConn
	UserService   pb.UserServiceClient
	SecretService pb.SecretServiceClient
	cfg           *config.Config
	log           zerolog.Logger
}

// New creates a new gRPC client with TLS credentials.
func New(cfg *config.Config, log zerolog.Logger) (*Client, error) {
	logCtx := log.With().
		Str("host", cfg.ServerHost).
		Int("port", cfg.ServerPort).
		Str("ca_path", cfg.ServerTLSCertPath).
		Logger()

	caCert, err := os.ReadFile(cfg.ServerTLSCertPath)
	if err != nil {
		logCtx.Error().Err(err).Msg("Failed to read ca certificate")
		return nil, fmt.Errorf("[%w]CA cert", e.ErrRead)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		logCtx.Error().Msg("Invalid ca certificate")
		return nil, fmt.Errorf("[%w]CA cert", e.ErrInvalidInput)
	}

	creds := credentials.NewClientTLSFromCert(certPool, cfg.ServerHost)

	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort),
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		logCtx.Error().Err(err).Msg("server connection error")
		return nil, fmt.Errorf("[%w]connect to gRPC", e.ErrUnavailable)
	}

	return &Client{
		Conn:          conn,
		UserService:   pb.NewUserServiceClient(conn),
		SecretService: pb.NewSecretServiceClient(conn),
		log:           log,
		cfg:           cfg,
	}, nil
}

func (c *Client) Close() error {
	err := c.Conn.Close()
	if err != nil {
		return fmt.Errorf("[%w] gRPC connection", e.ErrClose)
	}

	return nil
}

func (c *Client) Register(ctx context.Context) (*pb.RegisterResponse, error) {
	req := &pb.RegisterRequest{
		Username: c.cfg.Username,
		Password: c.cfg.Password,
		Role:     pb.UserRole_USER_ROLE_USER,
	}

	return c.UserService.Register(ctx, req)
}

func (c *Client) SecretUpdateInitRequest(
	ctx context.Context,
	scrt *dto.Secret,
) (*pb.SecretUpdateInitResponse, error) {
	req := &pb.SecretUpdateInitRequest{
		UserId:          scrt.UserID,
		SecretId:        scrt.ID,
		SecretName:      scrt.SecretName,
		VersionId:       scrt.VersionID,
		ParentVersionId: scrt.ParentVersionID,
		ClientInfo:      clientinfo.GenerateClientInfo(),
		Size:            scrt.SecretSize,
		Hash:            scrt.SecretHash,
		EncryptedDek:    scrt.SecretDek,
		MetadataJson:    "{}",
	}

	return c.SecretService.SecretUpdateInit(ctx, req)
}
