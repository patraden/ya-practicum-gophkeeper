package bootstrap

import (
	"context"
	"time"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/crypto"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/grpchandler"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/pg"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/s3"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/repository"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/server"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/version"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// Server returns server appl function as fx.App.
func Server(cfg *config.Config) *fx.App {
	logLevel := zerolog.ErrorLevel
	if cfg.DebugMode {
		logLevel = zerolog.DebugLevel
	}

	appLogger := logger.Stdout(logLevel)

	if cfg.InstallMode {
		return fx.New(
			fx.StartTimeout(time.Minute),
			fx.StopTimeout(time.Minute),
			fx.Supply(appLogger),
			fx.Provide(func(l *logger.Logger) *zerolog.Logger { return l.GetZeroLog() }),
			fx.Provide(func() (*pg.DB, error) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				return pg.NewDB(ctx, cfg.DatabaseDSN)
			}),
			fx.Provide(version.New),
			fx.Provide(fx.Annotate(s3.NewDummyS3, fx.As(new(s3.Client)))),
			fx.Provide(fx.Annotate(repository.NewUserRepo, fx.As(new(repository.UserRepository)))),
			fx.Provide(fx.Annotate(repository.NewREKRepo, fx.As(new(repository.REKRepository)))),
			fx.Provide(crypto.NewSplitter),
			fx.Supply(cfg),
			fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
			fx.Invoke(fxServerInstallInvoke),
		)
	}

	return fx.New(
		fx.StartTimeout(time.Minute),
		fx.StopTimeout(time.Minute),
		fx.Supply(appLogger),
		fx.Provide(func(l *logger.Logger) *zerolog.Logger { return l.GetZeroLog() }),
		fx.Supply(cfg),
		fx.Provide(fx.Annotate(grpchandler.NewAdminServer, fx.As(new(grpchandler.AdminServiceServer)))),
		fx.Provide(fx.Annotate(grpchandler.NewUserServer, fx.As(new(grpchandler.UserServiceServer)))),
		fx.Provide(version.New),
		fx.Provide(server.New),
		fx.WithLogger(appLogger.GetFxLogger()),
		fx.Invoke(fxServerInvoke),
	)
}
