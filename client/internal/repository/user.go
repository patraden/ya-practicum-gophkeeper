package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/patraden/ya-practicum-gophkeeper/client/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/infra/sqlite"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/dto"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	userDirPermissions = 0o700
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
	UserRepository
	queries *sqlite.Queries
	conn    *sql.DB
	cfg     *config.Config
	log     zerolog.Logger
}

func NewUserRepo(db *sqlite.DB, cfg *config.Config, log zerolog.Logger) *UserRepo {
	return &UserRepo{
		queries: db.Queries,
		conn:    db.Conn,
		cfg:     cfg,
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

	if err := repo.createUserDir(usr, logCtx); err != nil {
		logCtx.Error().Err(err).Msg("Failed to create user directory")
		repo.cleanupUserDir("creation failure", usr, logCtx)

		return err
	}

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

// createUserDir attempts to create a dedicated user directory.
func (repo *UserRepo) createUserDir(
	usr *user.User,
	logCtx zerolog.Logger,
) error {
	dirName := filepath.Join(repo.cfg.InstallDir, usr.BucketName)

	if err := os.MkdirAll(dirName, userDirPermissions); err != nil {
		logCtx.Error().Err(err).
			Str("path", dirName).
			Int("permissions", userDirPermissions).
			Msg("Failed to create user directory")

		return e.InternalErr(err)
	}

	return nil
}

// compensateUserDir deletes the user's local directory in case of failure.
func (repo *UserRepo) cleanupUserDir(
	reason string,
	usr *user.User,
	logCtx zerolog.Logger,
) {
	dirName := filepath.Join(repo.cfg.InstallDir, usr.BucketName)

	if err := os.RemoveAll(dirName); err != nil {
		logCtx.Error().Err(err).
			Str("reason", reason).
			Bool("compensation", true).
			Str("path", dirName).
			Msg("failed to remove user directory during compensation")
	} else {
		logCtx.Debug().
			Str("path", dirName).
			Str("reason", reason).
			Bool("compensation", true).
			Msg("successfully removed user bucket as compensation")
	}
}
