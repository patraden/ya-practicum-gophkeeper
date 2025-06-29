package app

import (
	"context"
	"fmt"

	"github.com/patraden/ya-practicum-gophkeeper/client/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/grpcclient"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/infra/sqlite"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/repository"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/auth"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/dto"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
)

func RegisterUser(cfg *config.Config, log logger.Logger) error {
	zlog := log.GetZeroLog()

	client, err := grpcclient.New(cfg, zlog)
	if err != nil {
		return e.InternalErr(err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestsTimeout)
	defer cancel()

	zlog.Info().
		Msg("Sending user register request to server...")

	resp, err := client.Register(ctx)
	if err != nil {
		return e.InternalErr(err)
	}

	zlog.Info().
		Msg("Successfully registered user in server...")

	usr, err := user.NewWithID(resp.GetUserId(), cfg.Username, resp.GetRole())
	if err != nil {
		return e.InternalErr(err)
	}

	usr.Salt = resp.GetSalt()
	usr.BucketName = resp.GetBucketName()
	usr.Verifier = resp.GetVerifier()

	token := &dto.ServerToken{
		UserID: resp.GetUserId(),
		Token:  resp.GetToken(),
		TTL:    resp.GetTokenTtlSeconds(),
	}

	if ok := auth.VerifyVerifier(cfg.Password, usr.Salt, usr.Verifier); !ok {
		return fmt.Errorf("[%w] wrong user verifier", e.ErrInternal)
	}

	zlog.Info().
		Msg("Creating registered user locally...")

	db, err := sqlite.NewDB(fmt.Sprintf("%s/%s", cfg.InstallDir, cfg.DatabaseFileName))
	if err != nil {
		return err
	}

	defer db.Close()

	repo := repository.NewUserRepo(db, zlog)
	if err := repo.CreateUser(ctx, usr, token); err != nil {
		return err
	}

	zlog.Info().
		Msg("Successfully registered user!")

	return nil
}
