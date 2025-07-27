package app

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/mailru/easyjson"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/infra/sqlite"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/rs/zerolog"
)

const (
	appDirPermissions = 0o700
	caCertFilename    = "ca.cert"
)

//nolint:funlen //reason: logging.
func SetupLocal(cfg *config.Config, log logger.Logger) error {
	basePath := cfg.InstallDir
	zlog := log.GetZeroLog()

	zlog.Info().Msg("Starting application installation")

	configFilePath := filepath.Join(basePath, config.ConfigFileName)
	if _, err := os.Stat(configFilePath); err == nil {
		zlog.Info().
			Str("path", configFilePath).
			Msg("Config file already exists, skipping installation")

		return nil
	}

	if err := os.MkdirAll(basePath, appDirPermissions); err != nil {
		zlog.Fatal().Err(err).
			Str("path", basePath).
			Int("permissions", appDirPermissions).
			Msg("Failed to create installation directory")

		return e.InternalErr(err)
	}

	if err := CopyCACertToInstallDir(cfg, zlog); err != nil {
		return err
	}

	dbPath := filepath.Join(basePath, cfg.DatabaseFileName)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		zlog.Fatal().Err(err).
			Str("path", dbPath).
			Str("db", "sqlite3").
			Msg("Failed to create application db")

		return e.InternalErr(err)
	}

	defer db.Close()

	if err := sqlite.RunClientMigrations(cfg, log); err != nil {
		zlog.Fatal().Err(err).
			Str("path", dbPath).
			Str("db", "sqlite3").
			Msg("Failed to apply migrations")

		return e.InternalErr(err)
	}

	if err := SaveToFile(cfg, zlog); err != nil {
		zlog.Error().Err(err).
			Str("path", configFilePath).
			Msg("Failed to save config to file")

		return e.InternalErr(err)
	}

	zlog.Info().Msg("Application installed successfully!")

	return nil
}

func SaveToFile(cfg *config.Config, log zerolog.Logger) error {
	basePath := cfg.InstallDir
	filePath := filepath.Join(basePath, config.ConfigFileName)

	log.Info().
		Str("path", filePath).
		Msg("Saving installation config to file")

	file, err := os.Create(filePath)
	if err != nil {
		log.Error().Err(err).
			Str("path", filePath).
			Msg("Faild to open config file for writing")

		return e.InternalErr(err)
	}
	defer file.Close()

	_, err = easyjson.MarshalToWriter(cfg, file)
	if err != nil {
		log.Error().Err(err).
			Str("path", filePath).
			Msg("Faild to marhcal config to file")

		return e.InternalErr(err)
	}

	return nil
}

func CopyCACertToInstallDir(cfg *config.Config, log zerolog.Logger) error {
	srcPath := cfg.ServerTLSCertPath
	dstPath := filepath.Join(cfg.InstallDir, caCertFilename)

	log.Info().
		Str("src", srcPath).
		Str("dst", dstPath).
		Msg("Copying CA certificate to install directory")

	src, err := os.Open(srcPath)
	if err != nil {
		log.Error().Err(err).
			Str("src", srcPath).
			Msg("Failed to open source CA cert for copying")

		return e.InternalErr(err)
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		log.Error().Err(err).
			Str("dst", dstPath).
			Msg("Failed to create destination CA cert file")

		return e.InternalErr(err)
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(src); err != nil {
		log.Error().Err(err).
			Str("dst", dstPath).
			Msg("Failed to copy CA cert")

		return e.InternalErr(err)
	}

	cfg.ServerTLSCertPath = dstPath

	log.Info().
		Str("dst", dstPath).
		Msg("CA certificate copied successfully")

	return nil
}
