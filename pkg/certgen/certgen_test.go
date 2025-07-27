//nolint:testpackage,funlen // reason: testing internal logic and long test functions are acceptable
package certgen

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestCreateTemplateUsage(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name          string
		cfg           Config
		wantKeyUsage  x509.KeyUsage
		wantExtUsages []x509.ExtKeyUsage
	}{
		{
			name: "basic server cert with DNS",
			cfg: Config{
				OrgName:   "TestOrg",
				Host:      "example.com",
				ValidFrom: now,
				ValidFor:  time.Hour,
			},
			wantKeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
			wantExtUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		},
		{
			name: "client cert with email",
			cfg: Config{
				OrgName:    "TestOrg",
				Host:       "test@example.com",
				IsClient:   true,
				CommonName: "client-user",
				ValidFrom:  now,
				ValidFor:   time.Hour,
			},
			wantKeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
			wantExtUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageEmailProtection, x509.ExtKeyUsageClientAuth},
		},
		{
			name: "CA cert",
			cfg: Config{
				OrgName:   "TestOrg",
				Host:      "example.com",
				IsCA:      true,
				ValidFrom: now,
				ValidFor:  time.Hour,
			},
			wantKeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
			wantExtUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		},
		{
			name: "server cert with URI",
			cfg: Config{
				OrgName:   "TestOrg",
				Host:      "spiffe://service.example.com",
				ValidFrom: now,
				ValidFor:  time.Hour,
			},
			wantKeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
			wantExtUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		},
		{
			name: "ip and dns cert",
			cfg: Config{
				Host:      "127.0.0.1,localhost",
				ValidFrom: now,
				ValidFor:  time.Hour,
			},
			wantKeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
			wantExtUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		},
		{
			name: "empty host",
			cfg: Config{
				OrgName:   "TestOrg",
				Host:      "",
				ValidFrom: now,
				ValidFor:  time.Hour,
			},
			wantKeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
			wantExtUsages: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpl, err := createTemplate(tt.cfg)
			if err != nil {
				t.Fatalf("createTemplate() error = %v", err)
			}

			if tmpl.KeyUsage != tt.wantKeyUsage {
				t.Errorf("KeyUsage = %v, want %v", tmpl.KeyUsage, tt.wantKeyUsage)
			}

			if !reflect.DeepEqual(tmpl.ExtKeyUsage, tt.wantExtUsages) {
				t.Errorf("ExtKeyUsage = %v, want %v", tmpl.ExtKeyUsage, tt.wantExtUsages)
			}
		})
	}
}

func TestGenerateAndVerifyCASignedCertificate(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()

	caCertPath := filepath.Join(tmpDir, "ca-public.crt")
	caKeyPath := filepath.Join(tmpDir, "ca-private.key")

	caCfg := Config{
		OrgName:    "TestCA",
		ECDSACurve: "P256",
		CommonName: "TestCA Root",
		ValidFrom:  time.Now(),
		ValidFor:   365 * 24 * time.Hour,
		IsCA:       true,
		Host:       "TestCA",
	}

	err := GenerateCertificate(caCfg, log)
	require.NoError(t, err, "generate CA certificate should not error")

	err = os.Rename("ca-public.crt", caCertPath)
	require.NoError(t, err)
	err = os.Rename("ca-private.key", caKeyPath)
	require.NoError(t, err)

	serverCertPath := filepath.Join(tmpDir, "public.crt")
	serverKeyPath := filepath.Join(tmpDir, "private.key")
	serverCfg := Config{
		OrgName:    "TestServer",
		ECDSACurve: "P256",
		CommonName: "test.local",
		ValidFrom:  time.Now(),
		ValidFor:   365 * 24 * time.Hour,
		Host:       "localhost,127.0.0.1",
		CACertPath: caCertPath,
		CAKeyPath:  caKeyPath,
	}

	err = GenerateCertificate(serverCfg, log)
	require.NoError(t, err, "generate server certificate should not error")

	err = os.Rename("public.crt", serverCertPath)
	require.NoError(t, err)
	err = os.Rename("private.key", serverKeyPath)
	require.NoError(t, err)

	caCertPEM, err := os.ReadFile(caCertPath)
	require.NoError(t, err, "read CA certificate should not error")

	caBlock, _ := pem.Decode(caCertPEM)
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	require.NoError(t, err, "parse CA certificate should not error")

	serverCertPEM, err := os.ReadFile(serverCertPath)
	require.NoError(t, err, "read server certificate should not error")

	serverBlock, _ := pem.Decode(serverCertPEM)
	serverCert, err := x509.ParseCertificate(serverBlock.Bytes)
	require.NoError(t, err, "parse server certificate should not error")

	roots := x509.NewCertPool()
	roots.AddCert(caCert)

	opts := x509.VerifyOptions{
		Roots:   roots,
		DNSName: "localhost",
	}

	_, err = serverCert.Verify(opts)
	require.NoError(t, err, "certificate verification by CA should pass")
}
