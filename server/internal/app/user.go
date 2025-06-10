package app

import (
	"context"
	"errors"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/crypto/keys"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/dto"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/auth"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/crypto/keystore"
	repository "github.com/patraden/ya-practicum-gophkeeper/server/internal/repository"
	"github.com/rs/zerolog"
)

// UserUseCase defines the core operations related to user authentication and registration.
type UserUseCase interface {
	// RegisterUser registers a user (admin or regular) with proper authorization and key wrapping.
	RegisterUser(ctx context.Context, creds *dto.RegisterUserCredentials) (*user.User, error)
	// ValidateUser checks user credentials against stored values.
	ValidateUser(ctx context.Context, creds *dto.UserCredentials) (*user.User, error)
}

// UserUC implements the UserUseCase interface and coordinates user auth logic.
type UserUC struct {
	repo     repository.UserRepository
	keyStore keystore.Keystore
	log      *zerolog.Logger
}

// NewUserUC returns a new instance of UserUC with dependencies injected.
func NewUserUC(repo repository.UserRepository, keyStore keystore.Keystore, log *zerolog.Logger) *UserUC {
	return &UserUC{
		repo:     repo,
		keyStore: keyStore,
		log:      log,
	}
}

// ValidateUser checks the given credentials and returns the user if valid.
func (u *UserUC) ValidateUser(ctx context.Context, creds *dto.UserCredentials) (*user.User, error) {
	return u.repo.ValidateUser(ctx, creds)
}

// RegisterUser registers a new user with the given credentials, wrapping their KEK with REK.
// Only admins can register admin users.
func (u *UserUC) RegisterUser(ctx context.Context, creds *dto.RegisterUserCredentials) (*user.User, error) {
	logCtx := u.log.With().Str("username", creds.Username).Logger()

	switch creds.Role {
	case user.RoleAdmin:
		return u.registerAdmin(ctx, creds, logCtx)
	case user.RoleUser:
		return u.registerUser(ctx, creds, logCtx)
	default:
		logCtx.Error().
			Str("role", creds.Role.String()).
			Msg("unsupported role provided in registration request")

		return nil, e.ErrInvalidInput
	}
}

// registerAdmin handles the creation of a new admin user.
// This requires the caller to already be authenticated as an admin.
func (u *UserUC) registerAdmin(
	ctx context.Context,
	creds *dto.RegisterUserCredentials,
	logCtx zerolog.Logger,
) (*user.User, error) {
	if creds.Role != user.RoleAdmin {
		return nil, e.ErrInvalidInput
	}

	_, claims, err := auth.FromContext(ctx)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to get auth user info")
		return nil, e.ErrUnauthorized
	}

	if claims.Role != string(user.RoleAdmin) {
		logCtx.Error().Msg("user is not authorised to register admin user")
		return nil, e.ErrUnauthorized
	}

	usr := user.New(creds.Username, creds.Role)
	if err := usr.SetPassword(creds.Password); err != nil {
		logCtx.Error().Err(err).Msg("user password generation error")
		return nil, e.InternalErr(err)
	}

	repoUser, err := u.repo.CreateAdmin(ctx, usr)
	if errors.Is(err, e.ErrExists) {
		return nil, err
	}

	if err != nil {
		return nil, e.InternalErr(err)
	}

	return repoUser, nil
}

// registerUser creates a new non-admin user with encrypted KEK stored in the database.
func (u *UserUC) registerUser(
	ctx context.Context,
	creds *dto.RegisterUserCredentials,
	logCtx zerolog.Logger,
) (*user.User, error) {
	if creds.Role != user.RoleUser {
		return nil, e.ErrInvalidInput
	}

	usr := user.New(creds.Username, creds.Role)
	if err := usr.SetPassword(creds.Password); err != nil {
		logCtx.Error().Err(err).Msg("user password generation error")
		return nil, e.InternalErr(err)
	}

	kek, err := keys.KEK(usr, creds.Password)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to generate kek for user")
		return nil, e.InternalErr(err)
	}

	rek, err := u.keyStore.Get()
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to get rek from keystore")
		return nil, e.InternalErr(err)
	}

	eKek, err := keys.WrapKEK(rek, kek)
	if err != nil {
		logCtx.Error().Err(err).Msg("failed to encrypt kek with rek")
		return nil, e.InternalErr(err)
	}

	key := user.NewKey(usr.ID, eKek, keys.EncryptionAlgo)
	repoUser, err := u.repo.CreateUser(ctx, usr, key)

	if errors.Is(err, e.ErrExists) {
		return nil, err
	}

	if err != nil {
		return nil, e.InternalErr(err)
	}

	return repoUser, nil
}
