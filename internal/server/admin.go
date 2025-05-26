package server

import (
	"github.com/patraden/ya-practicum-gophkeeper/internal/config"
	pb "github.com/patraden/ya-practicum-gophkeeper/internal/proto/gophkeeper/v1"
	"github.com/rs/zerolog"
)

type AdminServer struct {
	config *config.Config
	log    *zerolog.Logger
	pb.UnimplementedAdminServiceServer
}

func NewAdminServer(config *config.Config, log *zerolog.Logger) *AdminServer {
	return &AdminServer{
		config: config,
		log:    log,
	}
}
