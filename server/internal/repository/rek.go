package repository

import (
	"context"
	"errors"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/retry"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/pg"
	"github.com/rs/zerolog"
)

// REKRepository defines interface for storing and retrieving Root Encryption Key (REK) hash.
type REKRepository interface {
	// StoreHash inserts the REK hash into the database.
	// Returns ErrExists if the hash already exists.
	StoreHash(ctx context.Context, hash []byte) error

	// GetHash retrieves the stored REK hash from the database.
	GetHash(ctx context.Context) ([]byte, error)
}

// REKRepo implements REKRepository backed by PostgreSQL.
type REKRepo struct {
	connPool pg.ConnenctionPool
	queries  *pg.Queries
	log      *zerolog.Logger
}

// NewREKRepo creates a new REKRepo instance.
func NewREKRepo(db *pg.DB, log *zerolog.Logger) *REKRepo {
	return &REKRepo{
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
func (repo *REKRepo) withDBRetry(ctx context.Context, dbOp func() error) error {
	return retry.PG(ctx, backoff.NewExponentialBackOff(), repo.log, dbOp)
}

// StoreHash saves the given REK hash in the database.
// Returns e.ErrExists if the hash already exists (unique constraint violation).
func (repo *REKRepo) StoreHash(ctx context.Context, hash []byte) error {
	queryFn := func(queries *pg.Queries) error {
		var pgErr *pgconn.PgError

		err := queries.CreateREKHash(ctx, hash)
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return e.ErrExists
		}

		return err
	}

	dbErr := repo.withDBRetry(ctx, func() error { return queryFn(repo.queries) })

	if errors.Is(dbErr, e.ErrExists) {
		return dbErr
	}

	if dbErr != nil {
		repo.log.Error().Err(dbErr).
			Str("operation", "StoreHash").
			Msg("failed to create rek hash")

		return e.InternalErr(dbErr)
	}

	return nil
}

// GetHash retrieves the stored REK hash from the database.
// Returns e.ErrInternal if retrieval or retry fails.
func (repo *REKRepo) GetHash(ctx context.Context) ([]byte, error) {
	var hash []byte

	queryFn := func(queries *pg.Queries) error {
		row, err := queries.GetREKHash(ctx)
		if err != nil {
			return err
		}

		hash = row.RekHash

		return nil
	}

	dbErr := repo.withDBRetry(ctx, func() error { return queryFn(repo.queries) })
	if dbErr != nil {
		repo.log.Error().Err(dbErr).
			Str("operation", "GetHash").
			Msg("failed to get rek hash")

		return nil, e.InternalErr(dbErr)
	}

	return hash, nil
}
