package config

import (
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v6"
	easyjson "github.com/mailru/easyjson"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/rs/zerolog"
)

type builder struct {
	cfg *Config
	log zerolog.Logger
}

func newBuilder(dcfg *Config) *builder {
	return &builder{
		cfg: dcfg,
		log: logger.StdoutConsole(zerolog.InfoLevel).GetZeroLog(),
	}
}

func (b *builder) loadEnv() {
	if err := env.Parse(b.cfg); err != nil {
		b.log.Fatal().Err(err).
			Msg("Failed to parse config env")
	}
}

func (b *builder) loadFromFile() {
	path := filepath.Join(b.cfg.InstallDir, ConfigFileName)

	file, err := os.ReadFile(path)
	if err != nil {
		b.log.Fatal().Err(err).
			Str("file_path", path).
			Msg("Failed to open config file (run `gkcli install`?)")
	}

	if err := easyjson.Unmarshal(file, b.cfg); err != nil {
		b.log.Fatal().Err(err).
			Str("file_path", path).
			Msg("Failed to parse config file")
	}
}

func (b *builder) getConfig() *Config {
	b.loadFromFile()
	b.loadEnv()

	return b.cfg
}
