package transport

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/rs/zerolog"
)

// TLSHTTPTransportBuilder is responsible for building a custom HTTP transport
// with TLS configuration using provided certificate bytes or file path.
type TLSHTTPTransportBuilder struct {
	CertBytes []byte
	CertPath  string
	log       zerolog.Logger
}

// NewHTTPTransportBuilder creates a builder for HTTP transport with optional certificate input.
func NewHTTPTransportBuilder(certPath string, certBytes []byte, log zerolog.Logger) *TLSHTTPTransportBuilder {
	return &TLSHTTPTransportBuilder{
		CertBytes: certBytes,
		CertPath:  certPath,
		log:       log,
	}
}

// Build constructs a new *http.Transport with TLS configuration using the
// provided certificate. It reads the certificate from CertPath if CertBytes
// is not set. Returns an error if certificate loading or TLS config fails.
func (b *TLSHTTPTransportBuilder) Build() (*http.Transport, error) {
	if len(b.CertBytes) == 0 {
		if b.CertPath == "" {
			b.log.Error().
				Msg("path to certificate file is empty")

			return nil, fmt.Errorf("[%w] tls certificate file path", e.ErrEmptyInput)
		}

		certData, err := os.ReadFile(b.CertPath)
		if err != nil {
			b.log.Error().Err(err).
				Str("file_path", b.CertPath).
				Msg("failed to read certificate from file")

			return nil, fmt.Errorf("[%w] tls certificate", e.ErrRead)
		}

		b.CertBytes = certData
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(b.CertBytes) {
		b.log.Error().
			Msg("failed to add certificate to the pool")

		return nil, fmt.Errorf("[%w] tls certificate data", e.ErrInvalidInput)
	}

	tlsConfig := &tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
	}

	base := &http.Transport{}
	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		base = transport.Clone()
	}

	base.TLSClientConfig = tlsConfig

	return base, nil
}
