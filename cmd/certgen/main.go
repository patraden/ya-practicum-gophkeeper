package main

import (
	"flag"
	"log"
	"time"

	"github.com/patraden/ya-practicum-gophkeeper/internal/logger"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/utils/certgen"
	"github.com/rs/zerolog"
)

func main() {
	timeFlag := certgen.TimeFlag{T: time.Now()}
	cfg := certgen.Config{}
	logger := logger.Stdout(zerolog.DebugLevel).GetZeroLog()

	flag.StringVar(
		&cfg.Host,
		"host",
		"",
		"Comma-separated hostnames and IPs to generate a certificate for",
	)
	flag.StringVar(
		&cfg.ECDSACurve,
		"ecdsa-curve",
		"P256",
		"ECDSA curve to use to generate a key.Valid values are P224, P256 (recommended), P384, P521",
	)
	flag.BoolVar(&cfg.Ed25519, "ed25519", false, "Generate Ed25519 key")
	flag.StringVar(&cfg.OrgName, "org-name", "Certgen Development", "Organization name used when generating the certs")
	flag.StringVar(&cfg.CommonName, "common-name", "", "Common name for client cert")
	flag.BoolVar(&cfg.IsClient, "client", false, "Whether it's a client cert")
	flag.BoolVar(&cfg.IsCA, "ca", false, "Generate CA certificate")
	flag.Var(&timeFlag, "start-date", "Creation date formatted as Jan 1 15:04:05 2011")
	flag.DurationVar(&cfg.ValidFor, "duration", 365*24*time.Hour, "Duration that certificate is valid for")
	flag.StringVar(&cfg.CACertPath, "ca-cert", "", "Path to CA certificate (for signing child certs)")
	flag.StringVar(&cfg.CAKeyPath, "ca-key", "", "Path to CA private key (for signing child certs)")
	flag.Parse()

	cfg.ValidFrom = timeFlag.T

	if err := certgen.GenerateCertificate(cfg, logger); err != nil {
		log.Fatalf("Failed to generate certificate: %v", err)
	}
}
