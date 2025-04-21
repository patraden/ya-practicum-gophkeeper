package object_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/patraden/ya-practicum-gophkeeper/internal/server/config"
	"github.com/patraden/ya-practicum-gophkeeper/internal/server/model"
	"github.com/patraden/ya-practicum-gophkeeper/internal/server/object"
	"github.com/stretchr/testify/require"
)

const validCert = `-----BEGIN CERTIFICATE-----
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
			certBytes: []byte(validCert),
			wantErr:   nil,
		},
		{
			name:      "invalid certificate",
			certBytes: []byte("not a cert"),
			wantErr:   model.ErrMinioClientTransport,
		},
		{
			name:      "empty certificate",
			certBytes: []byte{},
			wantErr:   model.ErrMinioClientTransport,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := object.NewHTTPTransportBuilder("", testCase.certBytes)
			transport, err := builder.Build()

			if testCase.wantErr != nil {
				require.ErrorIs(t, err, testCase.wantErr)
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
	require.NoError(t, os.WriteFile(tempValidFile, []byte(validCert), 0o600))

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
			wantErr:  model.ErrMinioClientTransport,
		},
		{
			name:     "empty file path",
			certPath: "",
			wantErr:  model.ErrMinioClientTransport,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := object.NewHTTPTransportBuilder(testCase.certPath, nil)
			transport, err := builder.Build()

			if testCase.wantErr != nil {
				require.ErrorIs(t, err, testCase.wantErr)
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

func TestNewMinioClient_Invalid(t *testing.T) {
	t.Parallel()

	transport := &http.Transport{}
	cfg := &config.ObjectStorageConfig{
		Endpoint:  "http://invalid:1234",
		AccessKey: "key",
		SecretKey: "secret",
		Token:     "token",
	}

	client, err := object.NewMinioClient(cfg, transport)
	require.ErrorIs(t, err, model.ErrMinioClientCreate)
	require.Nil(t, client)
}
