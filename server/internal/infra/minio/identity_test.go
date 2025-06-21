package minio_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/testutil/http/roundtrip"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/minio"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssumeRoleSuccess(t *testing.T) {
	t.Parallel()

	mockResponse := `<?xml version="1.0" encoding="UTF-8"?>
		<AssumeRoleWithWebIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
			<AssumeRoleWithWebIdentityResult>
				<Credentials>
					<AccessKeyId>AKIAEXAMPLE</AccessKeyId>
					<SecretAccessKey>secret123</SecretAccessKey>
					<SessionToken>token123</SessionToken>
					<Expiration>2025-06-19T12:00:00Z</Expiration>
				</Credentials>
			</AssumeRoleWithWebIdentityResult>
		</AssumeRoleWithWebIdentityResponse>`

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()

	mockHTTPClient := roundtrip.NewTestHTTPClient(func(req *http.Request) *http.Response {
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
		}
	})

	client := minio.NewMinioWebIdentityClient("http://localhost:9000", mockHTTPClient, nil, log)

	creds, err := client.AssumeRole(context.Background(), "dummy-token", 3600)
	require.NoError(t, err)

	assert.Equal(t, "AKIAEXAMPLE", creds.AccessKeyID)
	assert.Equal(t, "secret123", creds.SecretAccessKey)
	assert.Equal(t, "token123", creds.SessionToken)
	assert.Equal(t, "2025-06-19T12:00:00Z", creds.Expiration)
}

func TestAssumeRoleInvalidXML(t *testing.T) {
	t.Parallel()

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()

	mockHTTPClient := roundtrip.NewTestHTTPClient(func(_ *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("<invalid>")),
		}
	})

	client := minio.NewMinioWebIdentityClient("http://localhost:9000", mockHTTPClient, nil, log)

	_, err := client.AssumeRole(context.Background(), "token", 3600)
	require.ErrorIs(t, err, e.ErrUnmarshal)
}

func TestAssumeRoleHTTPError(t *testing.T) {
	t.Parallel()

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()

	mockHTTPClient := roundtrip.NewTestHTTPClient(func(_ *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       io.NopCloser(bytes.NewBufferString("unauthorized")),
		}
	})

	client := minio.NewMinioWebIdentityClient("http://localhost:9000", mockHTTPClient, nil, log)

	_, err := client.AssumeRole(context.Background(), "token", 3600)
	require.ErrorIs(t, err, e.ErrValidation)
}

func TestAssumeRoleEmptyCreds(t *testing.T) {
	t.Parallel()

	mockResponse := `<?xml version="1.0" encoding="UTF-8"?>
		<AssumeRoleWithWebIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
			<AssumeRoleWithWebIdentityResult>
				<Credentials>
					<AccessKeyId></AccessKeyId>
					<SecretAccessKey></SecretAccessKey>
					<SessionToken></SessionToken>
					<Expiration></Expiration>
				</Credentials>
			</AssumeRoleWithWebIdentityResult>
		</AssumeRoleWithWebIdentityResponse>`

	log := logger.Stdout(zerolog.DebugLevel).GetZeroLog()

	mockHTTPClient := roundtrip.NewTestHTTPClient(func(_ *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockResponse)),
		}
	})

	client := minio.NewMinioWebIdentityClient("http://localhost:9000", mockHTTPClient, nil, log)

	_, err := client.AssumeRole(context.Background(), "token", 3600)
	require.ErrorIs(t, err, e.ErrValidation)
	assert.Contains(t, err.Error(), "validation")
}
