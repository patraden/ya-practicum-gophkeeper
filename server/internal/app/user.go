package app

import (
	"context"
	"errors"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/keys"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/dto"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/auth"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/keystore"
	repository "github.com/patraden/ya-practicum-gophkeeper/server/internal/repository"
	"github.com/rs/zerolog"
)

type UserUseCase interface {
	RegisterUser(ctx context.Context, creds *dto.UserCredentials) (*user.User, error)
	ValidateUser(ctx context.Context, creds *dto.UserCredentials) (*user.User, error)
}

type UserUC struct {
	repo     repository.UserRepository
	keyStore keystore.Keystore
	log      *zerolog.Logger
}

func NewUserUC(repo repository.UserRepository, keyStore keystore.Keystore, log *zerolog.Logger) *UserUC {
	return &UserUC{
		repo:     repo,
		keyStore: keyStore,
		log:      log,
	}
}

// RegisterUser registers a new user with the given credentials, wrapping their KEK with REK.
// Only admins can register admin users.
// It returns the created user or a domain-level error.
func (u *UserUC) RegisterUser(ctx context.Context, creds *dto.UserCredentials) (*user.User, error) {
	logCtx := u.log.With().Str("username", creds.Username).Logger()

	_, claims, err := auth.FromContext(ctx)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to get auth user info")
		return nil, err
	}

	authUserRole := claims.Role
	// Only admins are allowed to create other admin users
	if creds.Role == string(user.RoleAdmin) && authUserRole != string(user.RoleAdmin) {
		logCtx.Error().Msg("user is not authorised to register admin user")
		return nil, e.ErrUnauthorized
	}

	usr := user.New(creds.Username, user.Role(creds.Role))
	if err := usr.SetPassword(creds.Password); err != nil {
		logCtx.Error().Err(err).Msg("user password generation error")
		return nil, err
	}

	kek, err := keys.KEK(usr, creds.Password)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to generate kek for user")
		return nil, err
	}

	rek, err := u.keyStore.Get()
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to get rek from keystore")
		return nil, err
	}

	eKek, err := keys.WrapKEK(rek, kek)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed encrypt kek with rek")
		return nil, err
	}

	key := user.NewKey(usr.ID, eKek, keys.EncryptionAlgo)
	repoUser, err := u.repo.CreateUser(ctx, usr, key)

	if errors.Is(err, e.ErrExists) {
		return nil, err
	}

	if err != nil {
		logCtx.Error().Err(err).Msg("user registration error")
		return nil, err
	}

	return repoUser, nil
}
