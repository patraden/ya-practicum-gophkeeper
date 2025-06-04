package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/retry"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/pg"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/s3"
	"github.com/rs/zerolog"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *user.User) (*user.User, error)
	// ValidateUser(ctx context.Context, username, password string) (*user.User, error)
	// ChangeUserPassword(ctx context.Context, username, oldPassword, newPassword string) error
}

type UserRepo struct {
	s3client s3.Client
	connPool pg.ConnenctionPool
	queries  *pg.Queries
	log      *zerolog.Logger
}

func NewUserRepo(pool pg.ConnenctionPool, s3client s3.Client, log *zerolog.Logger) *UserRepo {
	return &UserRepo{
		connPool: pool,
		queries:  pg.New(pool),
		s3client: s3client,
		log:      log,
	}
}

func (repo *UserRepo) retryDB(ctx context.Context, dbOp func() error) error {
	return retry.PG(ctx, backoff.NewExponentialBackOff(), repo.log, dbOp)
}

// CreateUser creates a new user record in the database and provisions a dedicated
// S3 bucket for storing the user's secrets.
//
// The operation is transactional in spirit: it first attempts to create the
// S3 bucket, then inserts the user into the database. If the database insertion
// fails, the previously created bucket is removed as a best-effort compensation.
//
// It returns the created user with sensitive fields like Password zeroed out.
// If the user already exists, it returns ErrUserExists. For all other failures,
// it returns ErrServerInternal.
//
// This method retries transient database failures using an exponential backoff strategy.
//
//nolint:funlen // reason: method includes user models mapping.
func (repo *UserRepo) CreateUser(ctx context.Context, usr *user.User) (*user.User, error) {
	bucketName := usr.ID.String()
	logCtx := repo.log.With().
		Str("username", usr.Username).
		Str("user_id", usr.ID.String()).
		Str("user_role", usr.Role.String()).Logger()

	// Trying to create user bucket first.
	// It will be compensated later if db operations is not succesfful
	bucketErr := repo.s3client.MakeBucket(ctx, bucketName, map[string]string{
		"user_id":   usr.ID.String(),
		"user_role": usr.Role.String(),
	})

	if bucketErr != nil {
		logCtx.Error().Err(bucketErr).
			Msg("failed to create user bucket")

		return nil, e.ErrServerInternal
	}

	var dbUser *user.User

	queryFn := func(queries *pg.Queries) error {
		sqlUser, err := queries.CreateUser(ctx, pg.CreateUserParams{
			ID:        usr.ID,
			Username:  usr.Username,
			Role:      usr.Role,
			Password:  usr.Password,
			Salt:      usr.Salt,
			CreatedAt: usr.CreatedAt,
			UpdatedAt: usr.UpdatedAt,
		})

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return e.ErrUserExists
		}

		if err != nil {
			return fmt.Errorf("error creating user: %w", err)
		}

		dbUser = &user.User{
			ID:        sqlUser.ID,
			Username:  sqlUser.Username,
			Role:      sqlUser.Role,
			CreatedAt: sqlUser.CreatedAt,
			UpdatedAt: sqlUser.UpdatedAt,
			Password:  []byte{},
			Salt:      sqlUser.Salt,
		}

		if dbUser.ID != usr.ID {
			return e.ErrUserExists
		}

		return nil
	}

	dbOp := func() error { return queryFn(repo.queries) }

	dbErr := repo.retryDB(ctx, dbOp)
	if dbErr == nil {
		return dbUser, nil
	}

	logCtx.Error().Err(dbErr).
		Msg("Repo: failed to create user")

	// Compensate and remove created bucket.
	// Considering bucket removal compensation as "best-effort" operation.
	// It is non-critical and can be later compensated by background reconciliation.
	repo.compensateBucket(ctx, bucketName, "user_creation_failed", logCtx)

	return nil, e.ErrServerInternal
}

func (repo *UserRepo) compensateBucket(ctx context.Context, bucketName, reason string, logCtx zerolog.Logger) {
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
