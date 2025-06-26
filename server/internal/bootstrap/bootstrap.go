package bootstrap

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/s3"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/app"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/auth"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/crypto/keystore"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/crypto/shamir"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/grpchandler"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/identity"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/minio"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/pg"
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

	//nolint:unparam //reason: ignoring always nil error
	jwtKeyFunc := func(*jwt.Token) (any, error) { return []byte(cfg.JWTSecret), nil }
	appLogger := logger.Stdout(logLevel)
	pgDBFunc := func() (*pg.DB, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		return pg.NewDB(ctx, cfg.DatabaseDSN)
	}

	if cfg.InstallMode {
		return fx.New(
			fx.StartTimeout(time.Minute),
			fx.StopTimeout(time.Minute),
			fx.Supply(cfg),
			fx.Supply(appLogger),
			fx.Provide(version.New),
			fx.Provide(func(l logger.Logger) zerolog.Logger { return l.GetZeroLog() }),
			fx.Provide(pgDBFunc),
			fx.Provide(shamir.NewSplitter),
			fx.Provide(fx.Annotate(identity.KeycloakPGManager, fx.As(new(identity.Manager)))),
			fx.Provide(fx.Annotate(minio.NewClient, fx.As(new(s3.ServerOperator)))),
			fx.Provide(fx.Annotate(repository.NewUserRepo, fx.As(new(repository.UserRepository)))),
			fx.Provide(fx.Annotate(repository.NewREKRepo, fx.As(new(repository.REKRepository)))),
			fx.WithLogger(appLogger.GetFxLogger()),
			// fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
			fx.Invoke(fxServerInstallInvoke),
		)
	}

	return fx.New(
		fx.StartTimeout(time.Minute),
		fx.StopTimeout(time.Minute),
		fx.Supply(cfg),
		fx.Supply(server.PublicGRPCMethods),
		fx.Supply(appLogger),
		fx.Provide(version.New),
		fx.Provide(func(l logger.Logger) zerolog.Logger { return l.GetZeroLog() }),
		fx.Provide(func(l logger.Logger) *auth.Auth { return auth.New(jwtKeyFunc, l.GetZeroLog()) }),
		fx.Provide(pgDBFunc),
		fx.Provide(shamir.NewCollector),
		fx.Provide(fx.Annotate(identity.KeycloakPGManager, fx.As(new(identity.Manager)))),
		fx.Provide(fx.Annotate(minio.NewClient, fx.As(new(s3.ServerOperator)))),
		fx.Provide(fx.Annotate(keystore.NewInMemoryKeystore, fx.As(new(keystore.Keystore)))),
		fx.Provide(fx.Annotate(repository.NewREKRepo, fx.As(new(repository.REKRepository)))),
		fx.Provide(fx.Annotate(repository.NewUserRepo, fx.As(new(repository.UserRepository)))),
		fx.Provide(fx.Annotate(app.NewAdminUC, fx.As(new(app.AdminUseCase)))),
		fx.Provide(fx.Annotate(app.NewUserUC, fx.As(new(app.UserUseCase)))),
		fx.Provide(fx.Annotate(grpchandler.NewAdminServer, fx.As(new(grpchandler.AdminServiceServer)))),
		fx.Provide(fx.Annotate(grpchandler.NewUserServer, fx.As(new(grpchandler.UserServiceServer)))),
		fx.Provide(server.New),
		fx.WithLogger(fxevent.NopLogger),
		// fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
		fx.WithLogger(appLogger.GetFxLogger()),
		fx.Invoke(fxServerInvoke),
	)
}
