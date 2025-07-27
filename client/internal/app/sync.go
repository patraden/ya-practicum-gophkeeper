package app

import (
	"context"
	"fmt"

	"github.com/patraden/ya-practicum-gophkeeper/client/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/grpcclient"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/infra/minio"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/infra/sqlite"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/repository"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/dto"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/s3"
)

//nolint:funlen //reason: to refactor
func SyncSecrets(cfg *config.Config, secretName string, log logger.Logger) error {
	zlog := log.GetZeroLog()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestsTimeout)
	defer cancel()

	db, err := sqlite.NewDB(fmt.Sprintf("%s/%s", cfg.InstallDir, cfg.DatabaseFileName))
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to connect to db")
		return err
	}

	defer db.Close()

	userRepo := repository.NewUserRepo(db, cfg, zlog)
	secretRepo := repository.NewSecretRepo(db, zlog)

	zlog.Info().Msg("Validating user...")

	usr, err := userRepo.ValidateUser(ctx, &dto.UserCredentials{Username: cfg.Username, Password: cfg.Password})
	if err != nil {
		return err
	}

	zlog.Info().Msg("User is valid!...")
	zlog.Info().Msg("Validating user sercret...")

	scrt, err := secretRepo.GetSecret(ctx, usr.Username, secretName)
	if err != nil {
		return err
	}

	zlog.Info().Msg("User sercret is valid!...")
	zlog.Info().Msg("Sending sync request to server...")

	client, err := grpcclient.New(cfg, zlog)
	if err != nil {
		return e.InternalErr(err)
	}
	defer client.Close()

	resp, err := client.SecretUpdateInitRequest(ctx, scrt)
	if err != nil {
		return err
	}

	zlog.Info().Msg("Sync request confirmed by server!!")

	s3Config := &s3.ClientConfig{
		S3Endpoint:    cfg.S3Endpoint,
		S3TLSCertPath: cfg.ServerTLSCertPath,
		S3AccessKey:   resp.GetCredentials().GetAccessKeyId(),
		S3SecretKey:   resp.GetCredentials().GetSecretAccessKey(),
		S3Token:       resp.GetCredentials().GetSessionToken(),
		S3AccountID:   cfg.S3AccountID,
		S3Region:      cfg.S3Region,
	}

	minioClient, err := minio.NewClient(s3Config, zlog)
	if err != nil {
		return err
	}

	_, err = minioClient.PutObject(ctx, usr.BucketName, resp.GetS3Url(), scrt.FilePath, s3.PutObjectOptions{})
	if err != nil {
		return err
	}

	// here we need to commit upload transaction by sending relevant server request...

	return nil
}
