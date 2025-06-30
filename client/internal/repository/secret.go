package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/patraden/ya-practicum-gophkeeper/client/internal/config"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/infra/sqlite"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/dto"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/rs/zerolog"
)

// SecretRepository defines the interface for secret persistence operations.
type SecretRepository interface {
	CreateSecret(ctx context.Context, secret *dto.Secret) error
	GetSecret(ctx context.Context, userName, secretName string) (*dto.Secret, error)
}

// SecretRepo is a SQLite-backed implementation of SecretRepository.
type SecretRepo struct {
	SecretRepository
	queries *sqlite.Queries
	conn    *sql.DB
	cfg     *config.Config
	log     zerolog.Logger
}

// NewSecretRepo creates and initializes a new SecretRepo instance.
func NewSecretRepo(db *sqlite.DB, cfg *config.Config, log zerolog.Logger) *SecretRepo {
	return &SecretRepo{
		queries: db.Queries,
		conn:    db.Conn,
		cfg:     cfg,
		log:     log,
	}
}

// GetSecret returns a secret for the specified user and secret name.
// Returns ErrNotFound if no such secret exists.
func (repo *SecretRepo) GetSecret(ctx context.Context, userName, secretName string) (*dto.Secret, error) {
	dbSecret, err := repo.queries.GetSecret(ctx, sqlite.GetSecretParams{
		Username:   userName,
		SecretName: secretName,
	})

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("[%w] db secret", e.ErrNotFound)
	}

	if err != nil {
		return nil, e.InternalErr(err)
	}

	return &dto.Secret{
		ID:              dbSecret.SecretID,
		UserID:          dbSecret.UserID,
		SecretName:      dbSecret.SecretName,
		VersionID:       dbSecret.VersionID,
		ParentVersionID: dbSecret.ParentVersionID,
		FilePath:        dbSecret.FilePath,
		SecretSize:      dbSecret.SecretSize,
		SecretHash:      dbSecret.SecretHash,
		SecretDek:       dbSecret.SecretDek,
		CreatedAt:       dbSecret.CreatedAt,
		UpdatedAt:       dbSecret.UpdatedAt,
		InSync:          dbSecret.InSync > 0,
	}, nil
}

// CreateSecret attempts to insert a new secret into the database.
// Returns ErrExists if a conflict on (user_id, secret_id) or (user_id, secret_name) occurs.
func (repo *SecretRepo) CreateSecret(ctx context.Context, scrt *dto.Secret) error {
	err := repo.queries.CreateSecret(ctx, sqlite.CreateSecretParams{
		SecretID:        scrt.ID,
		UserID:          scrt.UserID,
		SecretName:      scrt.SecretName,
		VersionID:       scrt.VersionID,
		ParentVersionID: scrt.ParentVersionID,
		FilePath:        scrt.FilePath,
		SecretSize:      scrt.SecretSize,
		SecretHash:      scrt.SecretHash,
		SecretDek:       scrt.SecretDek,
		CreatedAt:       scrt.CreatedAt,
		UpdatedAt:       scrt.UpdatedAt,
		InSync:          0,
	})

	if sqlite.IsUniqueViolation(err) {
		return fmt.Errorf("[%w] db secret", e.ErrExists)
	}

	if err != nil {
		return e.InternalErr(err)
	}

	return nil
}
