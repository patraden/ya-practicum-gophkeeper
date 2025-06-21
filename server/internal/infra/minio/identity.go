package minio

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strconv"

	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/s3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// AssumeRoleWithWebIdentityResponse represents the STS response from MinIO.
type AssumeRoleWithWebIdentityResponse struct {
	XMLName xml.Name `xml:"AssumeRoleWithWebIdentityResponse"`
	Result  struct {
		Credentials s3.TemporaryCredentials `xml:"Credentials"`
	} `xml:"AssumeRoleWithWebIdentityResult"`
}

// STS request form keys.
const (
	paramAction          = "Action"
	paramVersion         = "Version"
	paramToken           = "WebIdentityToken"
	paramDurationSeconds = "DurationSeconds"

	actionAssumeRoleWithWebIdentity = "AssumeRoleWithWebIdentity"
	apiVersion                      = "2011-06-15"
	contentTypeURLEncoded           = "application/x-www-form-urlencoded"
)

// MinioWebIdentityClient performs AssumeRoleWithWebIdentity calls to MinIO.
type WebIdentityClient struct {
	webURL string
	client *http.Client
	log    *zerolog.Logger
}

// NewMinioWebIdentityClient constructs a new client.
func NewMinioWebIdentityClient(
	webURL string,
	httpClient *http.Client,
	httpTransport *http.Transport,
	log *zerolog.Logger,
) *WebIdentityClient {
	if httpClient != nil {
		return &WebIdentityClient{
			webURL: webURL,
			client: httpClient,
			log:    log,
		}
	}

	return &WebIdentityClient{
		webURL: webURL,
		client: &http.Client{
			Transport: httpTransport,
		},
	}
}

// AssumeRole performs the web identity authentication and returns temporary credentials.
//
//nolint:funlen // reason : logging.
func (c *WebIdentityClient) AssumeRole(
	ctx context.Context,
	identityToken string,
	durationSeconds int,
) (*s3.TemporaryCredentials, error) {
	form := url.Values{}
	form.Set(paramAction, actionAssumeRoleWithWebIdentity)
	form.Set(paramVersion, apiVersion)
	form.Set(paramToken, identityToken)
	form.Set(paramDurationSeconds, strconv.Itoa(durationSeconds))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		log.Error().Err(err).
			Str("method", http.MethodPost).
			Str("webURL", c.webURL).
			Msg("failed to create request")

		return nil, e.ErrGenerate
	}

	req.Header.Set("Content-Type", contentTypeURLEncoded)

	resp, err := c.client.Do(req)
	if err != nil {
		log.Error().Err(err).
			Str("method", http.MethodPost).
			Str("webURL", c.webURL).
			Msg("failed to send request")

		return nil, e.ErrInternal
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).
			Msg("failed to read response body")

		return nil, e.ErrRead
	}

	if resp.StatusCode != http.StatusOK {
		log.Error().
			Int("status", resp.StatusCode).
			Str("body", string(body)).
			Msg("unexpected HTTP status")

		return nil, e.ErrValidation
	}

	var result AssumeRoleWithWebIdentityResponse
	if err := xml.Unmarshal(body, &result); err != nil {
		log.Error().Err(err).
			Str("body", string(body)).
			Msg("failed to unmarshal response")

		return nil, e.ErrUnmarshal
	}

	creds := result.Result.Credentials
	if creds.AccessKeyID == "" || creds.SecretAccessKey == "" || creds.SessionToken == "" {
		log.Error().Err(err).
			Str("AccessKeyID", creds.AccessKeyID).
			Str("SecretAccessKey", creds.SecretAccessKey).
			Str("SessionToken", creds.SessionToken).
			Msg("empty creds response")

		return nil, e.ErrValidation
	}

	return &creds, nil
}
