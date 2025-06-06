package config

import (
	"flag"
	"fmt"
	"log"

	"github.com/caarlos0/env/v6"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
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
		log.Fatal(fmt.Errorf("failed to parse env: %w", e.ErrInvalidInput))
	}
}

func (b *builder) loadFlags() {
	flag.StringVar(&b.cfg.S3Endpoint, "s3-endpoint", b.cfg.S3Endpoint, "s3 endpoint {host}:{port}")
	flag.StringVar(&b.cfg.S3TLSCertPath, "s3-tls-cert", b.cfg.S3TLSCertPath, "s3 tls cert file path")
	flag.StringVar(&b.cfg.DatabaseDSN, "dsn", b.cfg.DatabaseDSN, "databse dsn")
	flag.BoolVar(&b.cfg.InstallMode, "install", b.cfg.InstallMode, "install server application")
	flag.BoolVar(&b.cfg.DebugMode, "d", b.cfg.DebugMode, "debug")
	flag.Parse()
}

func (b *builder) getConfig() *Config {
	b.loadFlags()
	b.loadEnv()

	return b.cfg
}
