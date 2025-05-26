package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/patraden/ya-practicum-gophkeeper/internal/domain/errors"
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
		log.Fatal(errors.ErrConfigEnvParse)
	}
}

func (b *builder) loadFlags() {
	flag.StringVar(&b.cfg.S3Endpoint, "s3-endpoint", b.cfg.S3Endpoint, "s3 endpoint {host}:{port}")
	flag.StringVar(&b.cfg.S3TLSCertPath, "s3-tls-cert", b.cfg.S3TLSCertPath, "s3 tls cert file path")
	flag.Parse()
}

func (b *builder) getConfig() *Config {
	b.loadFlags()
	b.loadEnv()

	return b.cfg
}
