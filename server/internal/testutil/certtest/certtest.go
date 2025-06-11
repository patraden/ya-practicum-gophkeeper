package certtest

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/certgen"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

// lock for parallel testing to keep generateTestCertificates helper func simple.
var certGenMu sync.Mutex

func GenerateTestCertificates(
	t *testing.T,
	dir string,
	log *zerolog.Logger,
) (string, string, string) {
	t.Helper()

	certGenMu.Lock()
	defer certGenMu.Unlock()

	caCertPath := filepath.Join(dir, "ca-public.crt")
	caKeyPath := filepath.Join(dir, "ca-private.key")
	serverCertPath := filepath.Join(dir, "server.crt")
	serverKeyPath := filepath.Join(dir, "server.key")

	// Generate CA cert
	require.NoError(t, certgen.GenerateCertificate(certgen.Config{
		OrgName:    "TestCA",
		CommonName: "TestCA",
		IsCA:       true,
		ECDSACurve: "P256",
		ValidFrom:  time.Now(),
		ValidFor:   365 * 24 * time.Hour,
		Host:       "localhost",
	}, log))

	require.NoError(t, os.Rename("ca-public.crt", caCertPath))
	require.NoError(t, os.Rename("ca-private.key", caKeyPath))

	// Generate server cert signed by CA
	require.NoError(t, certgen.GenerateCertificate(certgen.Config{
		OrgName:    "TestServer",
		CommonName: "localhost",
		ECDSACurve: "P256",
		CACertPath: caCertPath,
		CAKeyPath:  caKeyPath,
		Host:       "localhost,127.0.0.1",
		ValidFrom:  time.Now(),
		ValidFor:   365 * 24 * time.Hour,
	}, log))

	require.NoError(t, os.Rename("public.crt", serverCertPath))
	require.NoError(t, os.Rename("private.key", serverKeyPath))

	return caCertPath, serverCertPath, serverKeyPath
}
