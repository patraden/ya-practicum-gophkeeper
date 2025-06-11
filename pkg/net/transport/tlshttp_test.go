package transport_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/net/transport"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

const ValidCert = `-----BEGIN CERTIFICATE-----
MIIBaTCCARCgAwIBAgIQeqzCsR4c4UHsmG3Z2i88ijAKBggqhkjOPQQDAjAVMRMw
EQYDVQQKEwpHb3BoS2VlcGVyMB4XDTI1MDYwNzIwMzc0M1oXDTI2MDYwNzIwMzc0
M1owFTETMBEGA1UEChMKR29waEtlZXBlcjBZMBMGByqGSM49AgEGCCqGSM49AwEH
A0IABFSc129n05cdh9Rg7OUZd78I6qhuSKFxox0eJer/svRHZD5tK0fZ/Bu5bag+
8cf0Q2dzzDASBMvMirXS5Xj41jujQjBAMA4GA1UdDwEB/wQEAwICpDAPBgNVHRMB
Af8EBTADAQH/MB0GA1UdDgQWBBSHXnSKZ2lONP+CdJfHwrcH7Qw3PzAKBggqhkjO
PQQDAgNHADBEAiAva0+si2q9Zt4enx7Nwvsxt4UGnjnjWgKV2js4g6RtuQIgYf1t
yq7w5y8cgRgkww2wWIPufY/M7mBXWpsu1nUh0UM=
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
			wantErr:   errors.ErrInvalidInput,
		},
		{
			name:      "empty certificate",
			certBytes: []byte{},
			wantErr:   errors.ErrEmptyInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			log := zerolog.Nop()
			builder := transport.NewHTTPTransportBuilder("", tt.certBytes, &log)
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
			wantErr:  errors.ErrRead,
		},
		{
			name:     "empty file path",
			certPath: "",
			wantErr:  errors.ErrEmptyInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			log := zerolog.Nop()
			builder := transport.NewHTTPTransportBuilder(tt.certPath, nil, &log)
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
