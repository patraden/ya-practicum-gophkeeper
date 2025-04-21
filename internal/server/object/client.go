package object

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/patraden/ya-practicum-gophkeeper/internal/server/config"
	"github.com/patraden/ya-practicum-gophkeeper/internal/server/model"
)

// HTTPTransportBuilder is responsible for building a custom HTTP transport
// with TLS configuration using provided certificate bytes or file path.
type HTTPTransportBuilder struct {
	CertBytes []byte
	CertPath  string
}

// NewHTTPTransportBuilder creates a builder for HTTP transport with optional certificate input.
func NewHTTPTransportBuilder(certPath string, certBytes []byte) *HTTPTransportBuilder {
	return &HTTPTransportBuilder{
		CertBytes: certBytes,
		CertPath:  certPath,
	}
}

// Build constructs a new *http.Transport with TLS configuration using the
// provided certificate. It reads the certificate from CertPath if CertBytes
// is not set. Returns an error if certificate loading or TLS config fails.
func (b *HTTPTransportBuilder) Build() (*http.Transport, error) {
	if len(b.CertBytes) == 0 {
		if b.CertPath == "" {
			return nil, model.ErrMinioClientTransport
		}

		certData, err := os.ReadFile(b.CertPath)
		if err != nil {
			return nil, model.ErrMinioClientTransport
		}

		b.CertBytes = certData
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(b.CertBytes) {
		return nil, model.ErrMinioClientTransport
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

// NewMinioClient initializes a new MinIO client using the provided configuration
// and a custom HTTP transport. Returns a configured *minio.Client or an error.
func NewMinioClient(cfg *config.ObjectStorageConfig, transport *http.Transport) (*minio.Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, cfg.Token),
		Secure:    true,
		Transport: transport,
	})
	if err != nil {
		return nil, model.ErrMinioClientCreate
	}

	return client, nil
}
