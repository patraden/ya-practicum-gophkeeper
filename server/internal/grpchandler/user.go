package grpchandler

import (
	pb "github.com/patraden/ya-practicum-gophkeeper/pkg/proto/gophkeeper/v1"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/rs/zerolog"
)

type UserServer struct {
	config *config.Config
	log    *zerolog.Logger
	pb.UnimplementedUserServiceServer
}

func NewUserServer(config *config.Config, log *zerolog.Logger) *UserServer {
	return &UserServer{
		config: config,
		log:    log,
	}
}
