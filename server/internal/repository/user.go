package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/dto"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/retry"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/s3"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/identity"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/pg"
	"github.com/rs/zerolog"
)

// UserRepository defines user-related persistence operations.
type UserRepository interface {
	// CreateUser registers a regular user with KEK and an S3 bucket.
	CreateUser(ctx context.Context, usr *user.User, kek *user.Key) (*user.User, error)
	// GetUser get user by username.
	GetUser(ctx context.Context, username string) (*user.User, error)
	// GetUserByID get user by user id.
	GetUserByID(ctx context.Context, uid uuid.UUID) (*user.User, error)
	// CreateAdmin inserts an admin user without S3/KEK provisioning.
	CreateAdmin(ctx context.Context, usr *user.User) (*user.User, error)
	// ValidateUser Validates user credentials on Login.
	ValidateUser(ctx context.Context, creds *dto.UserCredentials) (*user.User, error)
}

// UserRepo implements UserRepository using PostgreSQL and S3.
type UserRepo struct {
	s3client s3.ServerOperator
	idClient identity.Manager
	connPool pg.ConnectionPool
	queries  *pg.Queries
	log      zerolog.Logger
}

// NewUserRepo creates a new instance of UserRepo with the provided database, S3 client, and logger.
func NewUserRepo(db *pg.DB, s3client s3.ServerOperator, idClient identity.Manager, log zerolog.Logger) *UserRepo {
	return &UserRepo{
		connPool: db.ConnPool,
		queries:  pg.New(db.ConnPool),
		s3client: s3client,
		idClient: idClient,
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

func (repo *UserRepo) logWithUserContext(usr *user.User, op string) zerolog.Logger {
	return repo.log.With().
		Str("repo", "UserRepo").
		Str("operation", op).
		Str("username", usr.Username).
		Str("user_id", usr.ID.String()).
		Str("user_role", usr.Role.String()).Logger()
}

// CreateAdmin inserts an admin user directly into the database.
// This method does not create an S3 bucket or KEK entry.
// Returns ErrExists if the user or username already exists.
func (repo *UserRepo) CreateAdmin(ctx context.Context, usr *user.User) (*user.User, error) {
	logCtx := repo.logWithUserContext(usr, "CreateAdmin")

	if usr.Role != user.RoleAdmin {
		logCtx.Error().Msg("expected user with admin role")
		return nil, e.ErrInvalidInput
	}

	var dbUsr *user.User

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
	logCtx := repo.logWithUserContext(usr, "CreateUser")

	if usr.Role != user.RoleUser {
		logCtx.Error().Msg("expected user with user role")
		return nil, fmt.Errorf("[%w] bad user role", e.ErrInvalidInput)
	}

	// Trying to create user in identity first.
	if err := repo.createIdentityUser(ctx, usr); err != nil {
		return nil, e.InternalErr(err)
	}

	// Trying to create user bucket first.
	if err := repo.createBucket(ctx, usr, logCtx); err != nil {
		repo.compensateIdentityUser(ctx, usr, "bucket_creation_failed", logCtx)
		return nil, e.InternalErr(err)
	}

	var dbUsr *user.User
	// Create pg user and it's key within trx.
	queryFn := pg.WithinTrx(ctx, repo.connPool, pgx.TxOptions{}, func(queries *pg.Queries) error {
		pgUser, err := queries.CreateUser(ctx, ToCreateUserParams(usr))

		if pg.IsUniqueViolation(err) {
			return fmt.Errorf("[%w] user", e.ErrExists)
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
			return fmt.Errorf("username %w", e.ErrExists)
		}

		return nil
	})

	dbErr := repo.withDBRetry(ctx, func() error { return queryFn(repo.queries) })
	if errors.Is(dbErr, e.ErrExists) {
		return nil, dbErr
	}

	if dbErr != nil {
		logCtx.Error().Err(dbErr).Msg("ailed to create user")
		// Compensate and remove created bucket and identity user.
		// Considering bucket and identity user removal compensation as "best-effort" operation.
		// It is non-critical and can be later compensated by background reconciliation.
		repo.compensateBucket(ctx, usr, "user_creation_failed", logCtx)
		repo.compensateIdentityUser(ctx, usr, "user_creation_failed", logCtx)

		return nil, e.InternalErr(dbErr)
	}

	return dbUsr, nil
}

// GetUser retrieves a user by username from the database.
//
// It returns ErrNotFound if the user does not exist.
// For all other errors, it returns ErrInternal.
// The password is removed before returning.
func (repo *UserRepo) GetUser(ctx context.Context, username string) (*user.User, error) {
	var dbUsr *user.User

	logCtx := repo.log.With().
		Str("repo", "UserRepo").
		Str("operation", "GetUser").
		Str("username", username).
		Logger()

	queryFn := func(queries *pg.Queries) error {
		pgUser, err := queries.GetUser(ctx, username)
		if err != nil {
			return err
		}

		dbUsr = FromPGUser(pgUser)

		return nil
	}

	dbErr := repo.withDBRetry(ctx, func() error { return queryFn(repo.queries) })
	if errors.Is(dbErr, sql.ErrNoRows) {
		return nil, fmt.Errorf("[%w] user", e.ErrNotFound)
	}

	if dbErr != nil {
		logCtx.Error().Err(dbErr).Msg("failed to get user")
		return nil, e.InternalErr(dbErr)
	}

	return dbUsr, nil
}

// GetUserByID retrieves a user by id from the database.
//
// It returns ErrNotFound if the user does not exist.
// For all other errors, it returns ErrInternal.
// The password is removed before returning.
func (repo *UserRepo) GetUserByID(ctx context.Context, uid uuid.UUID) (*user.User, error) {
	var dbUsr *user.User

	logCtx := repo.log.With().
		Str("repo", "UserRepo").
		Str("operation", "GetUser").
		Str("user_id", uid.String()).
		Logger()

	queryFn := func(queries *pg.Queries) error {
		pgUser, err := queries.GetUserByID(ctx, uid)
		if err != nil {
			return err
		}

		dbUsr = FromPGUser(pgUser)

		return nil
	}

	dbErr := repo.withDBRetry(ctx, func() error { return queryFn(repo.queries) })
	if errors.Is(dbErr, sql.ErrNoRows) {
		return nil, fmt.Errorf("[%w] user", e.ErrNotFound)
	}

	if dbErr != nil {
		logCtx.Error().Err(dbErr).Msg("failed to get user")
		return nil, e.InternalErr(dbErr)
	}

	return dbUsr, nil
}

// ValidateUser authenticates a user based on provided credentials.
//
// It first fetches the user record by username and then verifies the password.
// Returns ErrValidation if the password is incorrect and ErrNotFound if the user does not exist.
// For all other errors, it returns ErrInternal.
func (repo *UserRepo) ValidateUser(ctx context.Context, creds *dto.UserCredentials) (*user.User, error) {
	logCtx := repo.log.With().
		Str("repo", "UserRepo").
		Str("operation", "GetUser").
		Str("username", creds.Username).
		Logger()

	user, err := repo.GetUser(ctx, creds.Username)
	if err != nil {
		return nil, err
	}

	if !user.CheckPassword(creds.Password) {
		logCtx.Info().Msg("invalid password attempt")
		return nil, fmt.Errorf("[%w] user password", e.ErrValidation)
	}

	return user, nil
}

// createIdentityUser attempts to create identity user.
func (repo *UserRepo) createIdentityUser(ctx context.Context, usr *user.User) error {
	iuid, err := repo.idClient.CreateUser(ctx, usr)
	if err != nil {
		return err
	}

	usr.SetIdentityID(iuid)

	return nil
}

// createBucket attempts to create a dedicated S3 bucket for a new user.
func (repo *UserRepo) createBucket(
	ctx context.Context,
	usr *user.User,
	logCtx zerolog.Logger,
) error {
	bucketName := usr.IDNoDash()
	attrs := map[string]string{
		"user_id":   usr.ID.String(),
		"user_role": usr.Role.String(),
	}

	if err := repo.s3client.MakeBucket(ctx, bucketName, attrs); err != nil {
		logCtx.Error().Err(err).Msg("failed to create user bucket")
		return err
	}

	usr.SetBucketName(bucketName)

	return nil
}

// compensateIdentityUser deletes the previously created identity user.
// This is a best-effort operation for ensuring consistency between the database and object store.
func (repo *UserRepo) compensateIdentityUser(
	ctx context.Context,
	usr *user.User,
	reason string,
	logCtx zerolog.Logger,
) {
	if err := repo.idClient.DeleteUser(ctx, usr); err != nil {
		logCtx.Error().Err(err).
			Str("reason", reason).
			Bool("compensation", true).
			Str("identity_user", usr.IdentityID).
			Msg("failed to remove identity user during compensation")
	} else {
		logCtx.Info().
			Str("reason", reason).
			Bool("compensation", true).
			Str("identity_user", usr.IdentityID).
			Msg("successfully removed identity user as compensation")
	}
}

// compensateBucket deletes the previously created S3 bucket in case of a failed user creation.
// This is a best-effort operation for ensuring consistency between the database and object store.
func (repo *UserRepo) compensateBucket(ctx context.Context, usr *user.User, reason string, logCtx zerolog.Logger) {
	bucketName := usr.BucketName
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
