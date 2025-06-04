package grpchandler

import (
	pb "github.com/patraden/ya-practicum-gophkeeper/pkg/proto/gophkeeper/v1"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
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
