// Copyright (c) 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.
// Inspired by: https://github.com/minio/certgen

package certgen

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	serialNumberSize = 128
	privateKeyPerm   = 0o600
)

var (
	errInvalidPEMCert       = errors.New("invalid pem certificate")
	errInvalidPEMKey        = errors.New("invalid pem private key")
	errUnknownEllipticCurve = errors.New("unrecognized or missing elliptic curve")
)

// GenerateCertificate generates a self-signed or CA-signed X.509 certificate and writes it to disk.
//
// The behavior is determined by the provided CertConfig:
//   - If IsCA is true, generates a certificate authority (CA) certificate.
//   - If CACertPath and CAKeyPath are set, the certificate will be signed by the CA.
//   - If Ed25519 is true, uses Ed25519 keys; otherwise uses ECDSA (with specified curve).
//
// Logging details are written using the provided zerolog.Logger.
func GenerateCertificate(cfg Config, log zerolog.Logger) error {
	priv, pkey, err := createKeyPair(cfg)
	if err != nil {
		return err
	}

	template, err := createTemplate(cfg)
	if err != nil {
		return err
	}

	logTemplateInfo(log, template)

	parent, signer, err := maybeLoadCA(cfg, template, priv)
	if err != nil {
		return err
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, parent, pkey, signer)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	keyFile := "private.key"
	certFile := "public.crt"

	if cfg.IsClient {
		keyFile = "client-" + keyFile
		certFile = "client-" + certFile
	}

	if cfg.IsCA {
		keyFile = "ca-" + keyFile
		certFile = "ca-" + certFile
	}

	if err := saveCertificate(certFile, derBytes); err != nil {
		return err
	}

	if err := savePrivateKey(keyFile, priv); err != nil {
		return err
	}

	warnSecondLevelWildcards(cfg, log)

	return nil
}

func createKeyPair(cfg Config) (any, any, error) {
	if cfg.Ed25519 && cfg.ECDSACurve == "" {
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("ed25519 key generation failed: %w", err)
		}

		return priv, pub, nil
	}

	var curve elliptic.Curve

	switch cfg.ECDSACurve {
	case "P224":
		curve = elliptic.P224()
	case "P256":
		curve = elliptic.P256()
	case "P384":
		curve = elliptic.P384()
	case "P521":
		curve = elliptic.P521()
	default:
		return nil, nil, errUnknownEllipticCurve
	}

	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("ecdsa key generation failed: %w", err)
	}

	return priv, &priv.PublicKey, nil
}

func maybeLoadCA(cfg Config, template *x509.Certificate, priv any) (*x509.Certificate, any, error) {
	var err error

	parent := template
	signer := priv

	if !cfg.IsCA && cfg.CACertPath != "" && cfg.CAKeyPath != "" {
		parent, err = loadPEMCert(cfg.CACertPath)
		if err != nil {
			return nil, nil, err
		}

		signer, err = loadPEMKey(cfg.CAKeyPath)
		if err != nil {
			return nil, nil, err
		}
	}

	return parent, signer, nil
}

func loadPEMKey(path string) (any, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading pem private key %s: %w", path, err)
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil || !strings.Contains(block.Type, "PRIVATE KEY") {
		return nil, errInvalidPEMKey
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing pem private key: %w", err)
	}

	return key, nil
}

func loadPEMCert(path string) (*x509.Certificate, error) {
	certBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading pem certificate %s: %w", path, err)
	}

	block, _ := pem.Decode(certBytes)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, errInvalidPEMCert
	}

	key, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing pem certificate: %w", err)
	}

	return key, nil
}

func createTemplate(cfg Config) (*x509.Certificate, error) {
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, err
	}

	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{Organization: []string{cfg.OrgName}},
		NotBefore:             cfg.ValidFrom,
		NotAfter:              cfg.ValidFrom.Add(cfg.ValidFor),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	populateSANs(cfg, template)
	populateUsage(cfg, template)

	return template, nil
}

func generateSerialNumber() (*big.Int, error) {
	limit := new(big.Int).Lsh(big.NewInt(1), serialNumberSize)

	serial, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	return serial, nil
}

func populateSANs(cfg Config, template *x509.Certificate) {
	for host := range strings.SplitSeq(cfg.Host, ",") {
		host = strings.TrimSpace(host)

		if host == "" {
			continue
		}

		if ip := net.ParseIP(host); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else if email, err := mail.ParseAddress(host); err == nil && email.Address == host {
			template.EmailAddresses = append(template.EmailAddresses, host)
		} else if uriName, err := url.Parse(host); err == nil && uriName.Scheme != "" && uriName.Host != "" {
			template.URIs = append(template.URIs, uriName)
		} else {
			template.DNSNames = append(template.DNSNames, host)
		}
	}
}

func populateUsage(cfg Config, template *x509.Certificate) {
	if len(template.IPAddresses) > 0 || len(template.DNSNames) > 0 || len(template.URIs) > 0 {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
	}

	if len(template.EmailAddresses) > 0 {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageEmailProtection)
	}

	if cfg.IsClient {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	}

	if cfg.IsCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	// Only set CommonName for client or CA certs; server certs should rely on SANs
	if cfg.CommonName != "" && (cfg.IsClient || cfg.IsCA) {
		template.Subject.CommonName = cfg.CommonName
	}
}

func savePrivateKey(keyFile string, priv any) error {
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, privateKeyPerm)
	if err != nil {
		return fmt.Errorf("error opening %s for writing: %w", keyFile, err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("error marshaling %s: %w", keyFile, err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("error writing data to %s: %w", keyFile, err)
	}

	if err := keyOut.Close(); err != nil {
		return fmt.Errorf("error closing %s: %w", keyFile, err)
	}

	return nil
}

func saveCertificate(certFile string, derBytes []byte) error {
	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("error opening %s for writing: %w", certFile, err)
	}

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("error writing data to %s: %w", certFile, err)
	}

	if err := certOut.Close(); err != nil {
		return fmt.Errorf("error closing %s: %w", certFile, err)
	}

	return nil
}

func logTemplateInfo(log zerolog.Logger, template *x509.Certificate) {
	log.Info().
		Str("NotBefore", template.NotBefore.Format(time.RFC3339)).
		Str("NotAfter", template.NotAfter.Format(time.RFC3339)).
		Str("SerialNumber", template.SerialNumber.String()).
		Strs("DNSNames", template.DNSNames).
		Strs("EmailAddresses", template.EmailAddresses).
		Interface("IPAddresses", template.IPAddresses).
		Interface("URIs", template.URIs).
		Bool("IsCA", template.IsCA).
		Str("Org", strings.Join(template.Subject.Organization, ",")).
		Str("CommonName", template.Subject.CommonName).
		Msg("Generated certificate template")
}

func warnSecondLevelWildcards(cfg Config, log zerolog.Logger) {
	secondLvlWildcardRegexp := regexp.MustCompile(`(?i)^\*\.[0-9a-z_-]+$`)

	for host := range strings.SplitSeq(cfg.Host, ",") {
		host = strings.TrimSpace(host)
		if host == "" {
			continue
		}

		log.Info().Msgf("%q", host)

		if secondLvlWildcardRegexp.MatchString(host) {
			log.Warn().
				Msgf("Many browsers don't support second-level wildcards like %q", host)
		}
	}

	for host := range strings.SplitSeq(cfg.Host, ",") {
		host = strings.TrimSpace(host)
		if host == "" {
			continue
		}

		if strings.HasPrefix(host, "*.") {
			log.Info().
				Msgf("X.509 wildcards only go one level deep, so this won't match a.b.%s", host[2:])
			break
		}
	}
}
