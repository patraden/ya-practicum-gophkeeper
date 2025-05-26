package server

import (
	"github.com/patraden/ya-practicum-gophkeeper/internal/config"
	pb "github.com/patraden/ya-practicum-gophkeeper/internal/proto/gophkeeper/v1"
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
