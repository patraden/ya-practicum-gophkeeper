package bootstrap

import (
	"context"
	"errors"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/keys"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/crypto/shamir"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/pg"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/repository"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/version"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

const (
	defaultAdmin    = "Admin"
	defaultPassword = "Admin"
)

// fxServerInstallInvoke performs idempotent one-time installation steps:
// - Runs DB migrations
// - Generates and stores the REK hash
// - Creates a default admin user
// - Gracefully shuts down the app after setup.
func fxServerInstallInvoke(
	lc fx.Lifecycle,
	log zerolog.Logger,
	mlog logger.Logger,
	cfg *config.Config,
	version *version.Version,
	userRepo repository.UserRepository,
	rekRepo repository.REKRepository,
	splitter *shamir.Splitter,
	shutdowner fx.Shutdowner,
) {
	installLog := log.With().
		Str("component", "bootstrap").
		Str("phase", "install").Logger()

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			installLog.Info().
				Msg("starting installation")

			version.Log()

			installLog.Info().
				Msg("running database migrations")
			if err := pg.RunServerMigrations(cfg, mlog); err != nil {
				installLog.Error().Err(err).
					Msg("failed to apply database migrations")

				return err
			}

			installLog.Info().
				Msg("generating and storing REK")
			if err := generateREKShares(ctx, rekRepo, splitter, cfg.REKSharesPath, installLog); err != nil {
				return err
			}

			installLog.Info().Msg("creating default admin user")
			if err := createAdmin(ctx, userRepo, installLog); err != nil {
				return err
			}

			installLog.Info().
				Msg("installation completed successfully; shutting down")
			if err := shutdowner.Shutdown(); err != nil {
				installLog.Error().Err(err).
					Msg("failed to shutdown app after install")

				return e.InternalErr(err)
			}

			return nil
		},
		OnStop: func(_ context.Context) error {
			installLog.Info().
				Msg("installation lifecycle hook cleanup complete")
			return nil
		},
	})
}

func generateREKShares(
	ctx context.Context,
	rekRepo repository.REKRepository,
	splitter *shamir.Splitter,
	sharePath string,
	log zerolog.Logger,
) error {
	opLog := log.With().
		Str("operation", "generateREKShares").
		Logger()

	rek, err := keys.REK()
	if err != nil {
		opLog.Error().Err(err).
			Msg("failed to generate REK")

		return err
	}

	shares, err := splitter.Split(rek)
	if err != nil {
		opLog.Error().Err(err).
			Msg("failed to split REK into shares")

		return err
	}

	hash := keys.HashREK(rek)

	err = rekRepo.StoreHash(ctx, hash)

	switch {
	case err == nil:
		opLog.Info().
			Msg("REK hash stored successfully")
	case errors.Is(err, e.ErrExists):
		opLog.Info().
			Msg("REK hash already exists; skipping")
	default:
		opLog.Error().
			Err(err).Msg("failed to store REK hash")
		return err
	}

	if errors.Is(err, e.ErrExists) {
		opLog.Debug().
			Msg("skipping REK share output due to existing REK")
	} else {
		if err := WriteSharesFile(shares, sharePath, log); err != nil {
			opLog.Error().Err(err).
				Str("file", sharePath).
				Msg("failed to preserve shares to file")

			return err
		}
	}

	return nil
}

func createAdmin(ctx context.Context, userRepo repository.UserRepository, log zerolog.Logger) error {
	opLog := log.With().
		Str("operation", "createAdmin").
		Logger()

	adm := user.New(defaultAdmin, user.RoleAdmin)
	if err := adm.SetPassword(defaultPassword); err != nil {
		opLog.Error().Err(err).
			Msg("failed to set default admin password")

		return err
	}

	user, err := userRepo.CreateAdmin(ctx, adm)

	switch {
	case err == nil:
		opLog.Info().
			Str("username", user.Username).
			Msg("default admin user created")
	case errors.Is(err, e.ErrExists):
		opLog.Info().Str("username", defaultAdmin).
			Msg("default admin user already exists; skipping")
	default:
		opLog.Error().Err(err).
			Str("username", defaultAdmin).
			Msg("failed to create default admin user")

		return err
	}

	return nil
}
