package s3

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/rs/zerolog"
)

// HTTPTransportBuilder is responsible for building a custom HTTP transport
// with TLS configuration using provided certificate bytes or file path.
type HTTPTransportBuilder struct {
	CertBytes []byte
	CertPath  string
	log       *zerolog.Logger
}

// NewHTTPTransportBuilder creates a builder for HTTP transport with optional certificate input.
func NewHTTPTransportBuilder(certPath string, certBytes []byte, log *zerolog.Logger) *HTTPTransportBuilder {
	return &HTTPTransportBuilder{
		CertBytes: certBytes,
		CertPath:  certPath,
		log:       log,
	}
}

// Build constructs a new *http.Transport with TLS configuration using the
// provided certificate. It reads the certificate from CertPath if CertBytes
// is not set. Returns an error if certificate loading or TLS config fails.
func (b *HTTPTransportBuilder) Build() (*http.Transport, error) {
	if len(b.CertBytes) == 0 {
		if b.CertPath == "" {
			b.log.Error().
				Msg("path to certificate file is empty")

			return nil, errors.ErrMinioClientTransport
		}

		certData, err := os.ReadFile(b.CertPath)
		if err != nil {
			b.log.Error().
				Str("file_path", b.CertPath).
				Msg("failed to read certificate from file")

			return nil, errors.ErrMinioClientTransport
		}

		b.CertBytes = certData
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(b.CertBytes) {
		b.log.Error().
			Str("certificate", string(b.CertBytes)).
			Msg("failed to add certificate to the pool")

		return nil, errors.ErrMinioClientTransport
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
