package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cenkalti/backoff/v4"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/secret"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/retry"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/s3"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/identity"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/pg"
	"github.com/rs/zerolog"
)

// SecretRepository defines secret-related persistence operations.
type SecretRepository interface {
	CreateSecretInitRequest(
		ctx context.Context,
		req *secret.InitRequest,
	) (*secret.InitRequest, error)
	CreateSecretCommitRequest(
		ctx context.Context,
		req *secret.CommitRequest,
	) (*secret.CommitRequest, error)
}

// SecretRepo implements SecretRepository using PostgreSQL and S3.
type SecretRepo struct {
	s3client s3.ServerOperator
	idClient identity.Manager
	connPool pg.ConnectionPool
	queries  *pg.Queries
	log      zerolog.Logger
}

// NewSecretRepo creates a new instance of SecretRepository with the provided database, S3 client, and logger.
func NewSecretRepo(db *pg.DB, s3client s3.ServerOperator, idClient identity.Manager, log zerolog.Logger) *SecretRepo {
	return &SecretRepo{
		s3client: s3client,
		idClient: idClient,
		connPool: db.ConnPool,
		queries:  pg.New(db.ConnPool),
		log:      log,
	}
}

// withDBRetry performs the database operation with retry logic for transient errors:
// - ConnectionException
// - ConnectionDoesNotExist
// - ConnectionFailure
// - CannotConnectNow
// - SQLClientUnableToEstablishSQLConnection
// - TransactionResolutionUnknown.
func (repo *SecretRepo) withDBRetry(ctx context.Context, dbOp func() error) error {
	return retry.PG(ctx, backoff.NewExponentialBackOff(), repo.log, dbOp)
}

func (repo *SecretRepo) logWithRequestContext(req *secret.InitRequest, op string) zerolog.Logger {
	return repo.log.With().
		Str("repo", "SecretRepo").
		Str("operation", op).
		Str("user_id", req.UserID.String()).
		Str("secret_id", req.SecretID.String()).
		Str("secretName", req.SecretName).
		Str("clientInfo", req.ClientInfo).Logger()
}

// CreateSecretInitRequest.
func (repo *SecretRepo) CreateSecretInitRequest(
	ctx context.Context,
	req *secret.InitRequest,
) (*secret.InitRequest, error) {
	logCtx := repo.logWithRequestContext(req, "CreateSecretInitRequest")

	var dbReq *secret.InitRequest

	queryFn := func(queries *pg.Queries) error {
		row, err := queries.CreateSecretInitRequest(ctx, ToCreateSecretInitRequestParams(req))
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("[%w] secret parent version", e.ErrInvalidInput)
		}

		if err != nil {
			return err
		}

		dbReq = FromCreateSecretInitRequestParams(row)
		if dbReq.CreatedAt.Before(req.CreatedAt) || dbReq.VersionID != req.VersionID {
			return fmt.Errorf("[%w] secret init request", e.ErrExists)
		}

		return nil
	}

	dbErr := repo.withDBRetry(ctx, func() error { return queryFn(repo.queries) })
	if errors.Is(dbErr, e.ErrInvalidInput) || errors.Is(dbErr, e.ErrExists) {
		return nil, dbErr
	}

	if dbErr != nil {
		logCtx.Error().Err(dbErr).Msg("failed to create secret init request")
		return nil, e.InternalErr(dbErr)
	}

	// trying to generate s3 credentials now.
	creds, err := repo.getS3Credentials(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("[%w] create s3 credentials", e.ErrInternal)
	}

	dbReq.S3Creds = creds

	return dbReq, nil
}

// getS3Credentials TBD.
func (repo *SecretRepo) getS3Credentials(
	ctx context.Context,
	req *secret.InitRequest,
) (*s3.TemporaryCredentials, error) {
	idToken, err := repo.idClient.GetToken(ctx, req.User)
	if err != nil {
		return nil, err
	}

	creds, err := repo.s3client.AssumeRole(ctx, idToken.AccessToken, req.UploadDuration())
	if err != nil {
		return nil, err
	}

	return creds, nil
}
