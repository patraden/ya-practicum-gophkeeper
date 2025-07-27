package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/infra/sqlite"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/repository"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/keys"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/md5"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/stream"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/secret"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/dto"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/rs/zerolog"
)

//nolint:cyclop,funlen //reason: to refactor
func CreateSecret(cfg *config.Config, secretType, secretName, secretValue string, log logger.Logger) error {
	zlog := log.GetZeroLog()

	db, err := sqlite.NewDB(fmt.Sprintf("%s/%s", cfg.InstallDir, cfg.DatabaseFileName))
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to connect to db")
		return err
	}

	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestsTimeout)
	defer cancel()

	userRepo := repository.NewUserRepo(db, cfg, zlog)
	secretRepo := repository.NewSecretRepo(db, zlog)

	zlog.Info().Msg("Validating user...")

	usr, err := userRepo.ValidateUser(ctx, &dto.UserCredentials{Username: cfg.Username, Password: cfg.Password})
	if err != nil {
		return err
	}

	zlog.Info().Msg("User is valid!...")

	kek, err := keys.KEK(usr, cfg.Password)
	if err != nil {
		zlog.Error().Err(err).
			Msg("Failed generate keys encryption key")

		return err
	}

	_, err = secretRepo.GetSecret(ctx, cfg.Username, secretName)
	if err == nil {
		zlog.Error().
			Str("username", cfg.Username).
			Str("secret_name", secretName).
			Msg("Secret already exists")

		return fmt.Errorf("[%w] db secret", e.ErrExists)
	}

	if !errors.Is(err, e.ErrNotFound) {
		return err
	}

	var (
		scrt      *secret.Secret
		secretErr error
	)

	switch secretType {
	case "creds":
		secretErr = e.ErrNotImplemented
	case "binary":
		scrt, secretErr = createBinarySecret(cfg, kek, secretValue, secretName, usr, zlog)
	case "card":
		secretErr = e.ErrNotImplemented
	default:
		secretErr = e.ErrUnsupported
	}

	if secretErr != nil {
		return secretErr
	}

	zlog.Info().Msg("Storing secret in db...")

	err = secretRepo.CreateSecret(ctx, &dto.Secret{
		ID:              scrt.ID.String(),
		UserID:          scrt.UserID.String(),
		SecretName:      scrt.Name,
		VersionID:       scrt.CurrentVersionID.String(),
		ParentVersionID: scrt.CurrentVersion.ParentID.String(),
		FilePath:        scrt.CurrentVersion.S3URL,
		SecretSize:      scrt.CurrentVersion.SecretSize,
		SecretHash:      scrt.CurrentVersion.SecretHash,
		SecretDek:       scrt.CurrentVersion.SecretDEK,
		CreatedAt:       scrt.CreatedAt,
		UpdatedAt:       scrt.UpdatedAt,
		InSync:          false,
	})
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to create secret in db")
		return err
	}

	zlog.Info().Msg("successfully stored secret in db!")

	return nil
}

//nolint:funlen //reason: to refactor
func createBinarySecret(
	cfg *config.Config,
	kek []byte,
	filePath, secretName string,
	usr *user.User,
	log zerolog.Logger,
) (*secret.Secret, error) {
	log.Info().Msg("Creating secret...")

	srcFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("[%w] secret file", e.ErrRead)
	}
	defer srcFile.Close()

	dek, err := keys.DEK()
	if err != nil {
		log.Error().Err(err).
			Msg("Failed generate data encryption key")

		return nil, err
	}

	scrt := secret.NewSecret(usr.ID, secretName)
	destPath := filepath.Join(
		cfg.InstallDir,
		usr.BucketName,
		fmt.Sprintf("%s_%s.secret", secretName, scrt.ID.String()),
	)

	destFile, err := os.Create(destPath)
	if err != nil {
		log.Error().Err(err).
			Str("path", destPath).
			Msg("failed to create secret file")

		return nil, fmt.Errorf("[%w] secret file", e.ErrOpen)
	}
	defer destFile.Close()

	log.Info().Msg("Encrypting secret...")

	encryptReader, err := stream.EncryptSecretStream(srcFile, dek, log)
	if err != nil {
		return nil, err
	}

	secretSize, err := io.Copy(destFile, encryptReader)
	if err != nil {
		log.Error().Err(err).
			Str("path", destPath).
			Msg("failed to write encrypted secret")

		return nil, fmt.Errorf("[%w] write encrypted stream", e.ErrWrite)
	}

	log.Info().Msg("Encrypted secret successfully!")
	log.Info().Msg("Getting secret md5 hash...")

	secretMD5Hash, err := md5.GetFileMD5(destPath)
	if err != nil {
		log.Error().Err(err).
			Str("path", destPath).
			Msg("failed to get md5 hash for secret")

		return nil, err
	}

	encryptedDEK, err := keys.WrapDEK(kek, dek)
	if err != nil {
		log.Error().Err(err).
			Msg("failed to encrypt data encryption key")

		return nil, err
	}

	scrt = scrt.SetCurrentVersion(
		uuid.Nil,
		destPath,
		secretSize,
		secretMD5Hash,
		encryptedDEK,
	)

	log.Info().Msg("Secret created successfully!")

	return scrt, nil
}
