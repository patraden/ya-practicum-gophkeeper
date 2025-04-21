package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/patraden/ya-practicum-gophkeeper/internal/server/model"
)

type builder struct {
	cfg *Config
}

func newBuilder() *builder {
	return &builder{
		cfg: DefaultConfig(),
	}
}

func (b *builder) loadEnv() {
	if err := env.Parse(b.cfg); err != nil {
		log.Fatal(model.ErrConfigEnvParse)
	}
}

func (b *builder) loadFlags() {
	flag.StringVar(
		&b.cfg.ObjectStorage.Endpoint,
		"object-storage-endpoint",
		b.cfg.ObjectStorage.Endpoint,
		"object storage endpoint {host}:{port}",
	)
	flag.StringVar(
		&b.cfg.ObjectStorage.CertPath,
		"object-storage-cert",
		b.cfg.ObjectStorage.CertPath,
		"object storage certificate file path",
	)

	flag.Parse()
}

func (b *builder) getConfig() *Config {
	b.loadFlags()
	b.loadEnv()

	return b.cfg
}
