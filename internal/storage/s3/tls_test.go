package s3_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/patraden/ya-practicum-gophkeeper/internal/domain/errors"
	"github.com/patraden/ya-practicum-gophkeeper/internal/storage/s3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

const ValidCert = `-----BEGIN CERTIFICATE-----
MIIB5jCCAYygAwIBAgIRAPIBNlr5AgVg7LmwwHZ5AZowCgYIKoZIzj0EAwIwOjEc
MBoGA1UEChMTQ2VydGdlbiBEZXZlbG9wbWVudDEaMBgGA1UECwwRcm9vdEAxNjVl
OGYxOTM5OGEwHhcNMjUwNDEzMjAwMjExWhcNMjYwNDEzMjAwMjExWjA6MRwwGgYD
VQQKExNDZXJ0Z2VuIERldmVsb3BtZW50MRowGAYDVQQLDBFyb290QDE2NWU4ZjE5
Mzk4YTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABCLCGjyg3/DUExQ7byXOdF1W
vPyjzXlfAV5OwEJdbv8MZTtTz4rsJEZeUU/HSV2OpLYfs/R8j6pKYD3zcp5djiWj
czBxMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggrBgEFBQcDATAPBgNVHRMB
Af8EBTADAQH/MB0GA1UdDgQWBBSIRoY2U0W5CZQxMWBnYYk32+rGYzAaBgNVHREE
EzARgglsb2NhbGhvc3SHBH8AAAEwCgYIKoZIzj0EAwIDSAAwRQIgdMipjU+5qR2i
EmgAFSzd0YmFKzIA3VvMKcsLWTPmQ3sCIQCieiGaNkihZq9PFotCp6lVCC+F/Dfs
0bVOhAWuQTRKog==
-----END CERTIFICATE-----`

func TestHTTPTransportBuilder_FromBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		certBytes []byte
		wantErr   error
	}{
		{
			name:      "valid certificate",
			certBytes: []byte(ValidCert),
			wantErr:   nil,
		},
		{
			name:      "invalid certificate",
			certBytes: []byte("not a cert"),
			wantErr:   errors.ErrMinioClientTransport,
		},
		{
			name:      "empty certificate",
			certBytes: []byte{},
			wantErr:   errors.ErrMinioClientTransport,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			log := zerolog.Nop()
			builder := s3.NewHTTPTransportBuilder("", tt.certBytes, &log)
			transport, err := builder.Build()

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, transport)
			} else {
				require.NoError(t, err)
				require.NotNil(t, transport)
				require.NotNil(t, transport.TLSClientConfig)
				require.NotNil(t, transport.TLSClientConfig.RootCAs)
			}
		})
	}
}

func TestHTTPTransportBuilder_FromFile(t *testing.T) {
	t.Parallel()

	tempValidFile := filepath.Join(t.TempDir(), "valid-cert.pem")
	require.NoError(t, os.WriteFile(tempValidFile, []byte(ValidCert), 0o600))

	tests := []struct {
		name     string
		certPath string
		wantErr  error
	}{
		{
			name:     "valid certificate",
			certPath: tempValidFile,
			wantErr:  nil,
		},
		{
			name:     "non-existent file",
			certPath: filepath.Join(t.TempDir(), "no-such-file.pem"),
			wantErr:  errors.ErrMinioClientTransport,
		},
		{
			name:     "empty file path",
			certPath: "",
			wantErr:  errors.ErrMinioClientTransport,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			log := zerolog.Nop()
			builder := s3.NewHTTPTransportBuilder(tt.certPath, nil, &log)
			transport, err := builder.Build()

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, transport)
			} else {
				require.NoError(t, err)
				require.NotNil(t, transport)
				require.NotNil(t, transport.TLSClientConfig)
				require.NotNil(t, transport.TLSClientConfig.RootCAs)
			}
		})
	}
}
