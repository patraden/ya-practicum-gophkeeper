package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/patraden/ya-practicum-gophkeeper/client/internal/infra/sqlite"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/dto"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/rs/zerolog"
)

// UserRepository defines user-related persistence operations.
type UserRepository interface {
	// CreateUser registers a regular user with KEK and an S3 bucket.
	CreateUser(ctx context.Context, usr *user.User, token *dto.ServerToken) error
	// GetUser get user by username.
	GetUser(ctx context.Context, username string) (*user.User, error)
	// GetUserByID get user by user id.
	// ValidateUser Validates user credentials on Login.
	ValidateUser(ctx context.Context, creds *dto.UserCredentials) (*user.User, error)
}

type UserRepo struct {
	queries *sqlite.Queries
	conn    *sql.DB
	log     zerolog.Logger
}

func NewUserRepo(db *sqlite.DB, log zerolog.Logger) *UserRepo {
	return &UserRepo{
		queries: db.Queries,
		conn:    db.Conn,
		log:     log,
	}
}

func (repo *UserRepo) logWithUserContext(usr *user.User, op string) zerolog.Logger {
	return repo.log.With().
		Str("repo", "UserRepo").
		Str("operation", op).
		Str("username", usr.Username).
		Str("user_id", usr.ID.String()).
		Str("user_role", usr.Role.String()).Logger()
}

func (repo *UserRepo) CreateUser(ctx context.Context, usr *user.User, token *dto.ServerToken) error {
	logCtx := repo.logWithUserContext(usr, "CreateUser")

	queryFn := sqlite.WithinTrx(ctx, repo.conn, &sql.TxOptions{}, func(queries *sqlite.Queries) error {
		err := queries.CreateUser(ctx, sqlite.CreateUserParams{
			ID:         usr.ID.String(),
			Username:   usr.Username,
			Verifier:   usr.Verifier,
			Role:       usr.Role,
			Salt:       usr.Salt,
			Bucketname: usr.BucketName,
			CreatedAt:  usr.CreatedAt,
			UpdatedAt:  usr.UpdatedAt,
		})
		if err != nil {
			return err
		}

		return queries.CreateUserToken(ctx, sqlite.CreateUserTokenParams{
			UserID: token.UserID,
			Token:  token.Token,
			Ttl:    int64(token.TTL),
		})
	})

	err := queryFn(repo.queries)

	if sqlite.IsUniqueViolation(err) {
		logCtx.Error().Err(err).Msg("Failed to create db user")
		return fmt.Errorf("[%w] already exists", e.ErrExists)
	}

	if err != nil {
		logCtx.Error().Err(err).Msg("Failed to create db user")
		return e.InternalErr(err)
	}

	return nil
}
