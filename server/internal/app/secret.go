package app

import (
	"context"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/secret"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/crypto/keystore"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/repository"
	"github.com/rs/zerolog"
)

// SecretUseCase defines the core operations related to user sercrets.
type SecretUseCase interface {
	InitUploadRequest(ctx context.Context) (*secret.InitRequest, error)
}

// SecretUC implements the SecretUseCase interface.
type SecretUC struct {
	SecretUseCase
	repo     repository.SecretRepository
	keyStore keystore.Keystore
	log      zerolog.Logger
}

func NewSecretUC(
	repo repository.SecretRepository,
	keyStore keystore.Keystore,
	log zerolog.Logger,
) *SecretUC {
	return &SecretUC{
		repo:     repo,
		keyStore: keyStore,
		log:      log,
	}
}
