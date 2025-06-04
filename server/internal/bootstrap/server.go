package bootstrap

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/grpchandler"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/pg"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/server"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/version"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// Server returns server appl function as fx.App.
func Server(config *config.Config) *fx.App {
	logLevel := zerolog.ErrorLevel
	if config.DebugMode {
		logLevel = zerolog.DebugLevel
	}

	appLogger := logger.Stdout(logLevel)

	if config.InstallMode {
		return fx.New(
			fx.StartTimeout(time.Minute),
			fx.StopTimeout(time.Minute),
			fx.Supply(appLogger),
			fx.Provide(func(l *logger.Logger) *zerolog.Logger { return l.GetZeroLog() }),
			fx.Provide(version.New),
			fx.Supply(config),
			fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
			fx.Invoke(fxServerInstallInvoke),
		)
	}

	return fx.New(
		fx.StartTimeout(time.Minute),
		fx.StopTimeout(time.Minute),
		fx.Supply(appLogger),
		fx.Provide(func(l *logger.Logger) *zerolog.Logger { return l.GetZeroLog() }),
		fx.Supply(config),
		fx.Provide(fx.Annotate(grpchandler.NewAdminServer, fx.As(new(grpchandler.AdminServiceServer)))),
		fx.Provide(fx.Annotate(grpchandler.NewUserServer, fx.As(new(grpchandler.UserServiceServer)))),
		fx.Provide(version.New),
		fx.Provide(server.New),
		fx.WithLogger(appLogger.GetFxLogger()),
		fx.Invoke(fxServerInvoke),
	)
}

func fxServerInvoke(
	lc fx.Lifecycle,
	log *zerolog.Logger,
	cfg *config.Config,
	shutdowner fx.Shutdowner,
	version *version.Version,
	server *server.GRPCServer,
) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			// handle extra signals.
			handleSignals(shutdowner, log)
			startServerAsync(shutdowner, server, log)

			version.Log()
			logStart(log, cfg)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logStop(log, cfg)

			return server.Shutdown(ctx)
		},
	})
}

func handleSignals(shutdowner fx.Shutdowner, log *zerolog.Logger) {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-stopChan
		log.Info().
			Str("Signal", sig.String()).
			Msg("App shutdown signal received")

		err := shutdowner.Shutdown()
		if err != nil {
			log.Error().Err(err).
				Str("Signal", sig.String()).
				Msg("Failed to shutdown the App")
		}
	}()
}

func startServerAsync(shutdowner fx.Shutdowner, server *server.GRPCServer, log *zerolog.Logger) {
	go func() {
		err := server.Run()
		if err != nil {
			log.Error().Err(err).
				Msg("Stopping App due to gRPC server error")

			err := shutdowner.Shutdown()
			if err != nil {
				log.Error().Err(err).
					Msg("Failed to shutdown the App")
			}
		}
	}()
}

func logStart(log *zerolog.Logger, config *config.Config) {
	log.Info().
		Str("SERVER_ADDRESS", config.ServerAddr).
		Str("SERVER_TLS_KEY_PATH", config.ServerTLSKeyPath).
		Str("SERVER_TLS_CERT_PATH", config.ServerTLSCertPath).
		Str("S3_ENDPOINT", config.S3Endpoint).
		Str("S3_TLS_CERT_PATH", config.S3TLSCertPath).
		Str("S3_ACCESS_KEY", config.S3AccessKey).
		Msg("App started")
}

func logStop(log *zerolog.Logger, config *config.Config) {
	log.Info().
		Str("SERVER_ADDRESS", config.ServerAddr).
		Str("SERVER_TLS_KEY_PATH", config.ServerTLSKeyPath).
		Str("SERVER_TLS_CERT_PATH", config.ServerTLSCertPath).
		Str("S3_ENDPOINT", config.S3Endpoint).
		Str("S3_TLS_CERT_PATH", config.S3TLSCertPath).
		Str("S3_ACCESS_KEY", config.S3AccessKey).
		Str("SERVER_ADDRESS", config.ServerAddr).
		Msg("App stopped")
}

func fxServerInstallInvoke(
	lc fx.Lifecycle,
	log *zerolog.Logger,
	mlog *logger.Logger,
	cfg *config.Config,
	version *version.Version,
	shutdowner fx.Shutdowner,
) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			log.Info().Msg("Starting installation up")

			version.Log()

			if err := pg.RunServerMigrations(cfg, mlog); err != nil {
				return err
			}

			err := shutdowner.Shutdown()
			if err != nil {
				log.Error().Err(err).
					Msg("Failed to shutdown the App")
			}

			return nil
		},
		OnStop: func(_ context.Context) error {
			log.Info().Msg("Wrapping installation down")

			return nil
		},
	})
}
