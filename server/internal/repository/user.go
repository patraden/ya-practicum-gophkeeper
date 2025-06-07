package repository

import (
	"context"
	"errors"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgx/v5"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/retry"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/pg"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/s3"
	"github.com/rs/zerolog"
)

// UserRepository defines user-related persistence operations.
type UserRepository interface {
	// CreateUser registers a regular user with KEK and an S3 bucket.
	CreateUser(ctx context.Context, usr *user.User, kek *user.Key) (*user.User, error)
	// CreateAdmin inserts an admin user without S3/KEK provisioning.
	CreateAdmin(ctx context.Context, usr *user.User) (*user.User, error)
	// Future: Validate credentials.
	// ValidateUser(ctx context.Context, username, password string) (*user.User, error)
}

// UserRepo implements UserRepository using PostgreSQL and S3.
type UserRepo struct {
	s3client s3.Client
	connPool pg.ConnenctionPool
	queries  *pg.Queries
	log      *zerolog.Logger
}

// NewUserRepo creates a new instance of UserRepo with the provided database, S3 client, and logger.
func NewUserRepo(db *pg.DB, s3client s3.Client, log *zerolog.Logger) *UserRepo {
	return &UserRepo{
		connPool: db.ConnPool,
		queries:  pg.New(db.ConnPool),
		s3client: s3client,
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
func (repo *UserRepo) withDBRetry(ctx context.Context, dbOp func() error) error {
	return retry.PG(ctx, backoff.NewExponentialBackOff(), repo.log, dbOp)
}

func (repo *UserRepo) logWithUserContext(u *user.User) zerolog.Logger {
	return repo.log.With().
		Str("username", u.Username).
		Str("user_id", u.ID.String()).
		Str("user_role", u.Role.String()).Logger()
}

// CreateAdmin inserts an admin user directly into the database.
// This method does not create an S3 bucket or KEK entry.
// Returns ErrExists if the user or username already exists.
func (repo *UserRepo) CreateAdmin(ctx context.Context, usr *user.User) (*user.User, error) {
	var dbUsr *user.User

	logCtx := repo.logWithUserContext(usr)
	queryFn := func(queries *pg.Queries) error {
		pgUser, err := queries.CreateUser(ctx, ToCreateUserParams(usr))

		if pg.IsUniqueViolation(err) {
			return e.ErrExists
		}

		if err != nil {
			return err
		}

		dbUsr = FromPGUser(pgUser)

		// The DB enforces uniqueness on ID and username, but only one constraint may trigger.
		// Here we ensure both values match what we intended to store.
		if dbUsr.ID != usr.ID {
			return e.ErrExists
		}

		return nil
	}

	dbErr := repo.withDBRetry(ctx, func() error { return queryFn(repo.queries) })
	if errors.Is(dbErr, e.ErrExists) {
		return nil, dbErr
	}

	if dbErr != nil {
		logCtx.Error().Err(dbErr).
			Str("operation", "CreateAdmin").
			Msg("failed to create user")

		return nil, e.InternalErr(dbErr)
	}

	return dbUsr, nil
}

// CreateUser creates a new user record in the database and provisions a dedicated
// S3 bucket for storing the user's secrets.
//
// The operation is transactional in spirit: it first attempts to create the
// S3 bucket, then inserts the user/key into the database. If the database insertion
// fails, the previously created bucket is removed as a best-effort compensation.
//
// It returns the created user with sensitive fields like Password zeroed out.
// If the user already exists, it returns ErrUserExists. For all other failures,
// it returns ErrServerInternal.
//
// This method retries transient database failures using an exponential backoff strategy.
func (repo *UserRepo) CreateUser(ctx context.Context, usr *user.User, key *user.Key) (*user.User, error) {
	logCtx := repo.logWithUserContext(usr)

	// Trying to create user bucket first.
	// It will be compensated later if DB operation fails.
	if err := repo.createBucket(ctx, usr, &logCtx); err != nil {
		return nil, e.InternalErr(err)
	}

	var dbUsr *user.User
	// Create pg user and it's key within trx.
	queryFn := pg.WithinTrx(ctx, repo.connPool, pgx.TxOptions{}, func(queries *pg.Queries) error {
		pgUser, err := queries.CreateUser(ctx, ToCreateUserParams(usr))

		if pg.IsUniqueViolation(err) {
			return e.ErrExists
		}

		if err != nil {
			return err
		}

		if err = queries.CreateUserKey(ctx, ToCreateUserKeyParams(key)); err != nil {
			return err
		}

		dbUsr = FromPGUser(pgUser)

		// The DB enforces uniqueness on ID and username, but only one constraint may trigger.
		// Here we ensure both values match what we intended to store.
		if dbUsr.ID != usr.ID {
			return e.ErrExists
		}

		return nil
	})

	dbErr := repo.withDBRetry(ctx, func() error { return queryFn(repo.queries) })
	if dbErr != nil {
		if errors.Is(dbErr, e.ErrExists) {
			return nil, dbErr
		}

		logCtx.Error().Err(dbErr).
			Str("operation", "CreateUser").
			Msg("Repo: failed to create user")

		// Compensate and remove created bucket.
		// Considering bucket removal compensation as "best-effort" operation.
		// It is non-critical and can be later compensated by background reconciliation.
		repo.compensateBucket(ctx, usr, "user_creation_failed", &logCtx)

		return nil, e.InternalErr(dbErr)
	}

	return dbUsr, nil
}

func (repo *UserRepo) createBucket(ctx context.Context, usr *user.User, logCtx *zerolog.Logger) error {
	bucketName := usr.ID.String()

	err := repo.s3client.MakeBucket(ctx, bucketName, map[string]string{
		"user_id":   usr.ID.String(),
		"user_role": usr.Role.String(),
	})
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to create user bucket")
		return err
	}

	return nil
}

func (repo *UserRepo) compensateBucket(ctx context.Context, usr *user.User, reason string, logCtx *zerolog.Logger) {
	bucketName := usr.ID.String()
	if err := repo.s3client.RemoveBucket(ctx, bucketName); err != nil {
		logCtx.Error().Err(err).
			Str("reason", reason).
			Bool("compensation", true).
			Str("bucket", bucketName).
			Msg("failed to remove user bucket during compensation")
	} else {
		logCtx.Info().
			Str("reason", reason).
			Bool("compensation", true).
			Str("bucket", bucketName).
			Msg("successfully removed user bucket as compensation")
	}
}
