package app

import (
	"context"
	"errors"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/dto"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/auth"
	repository "github.com/patraden/ya-practicum-gophkeeper/server/internal/repository"
	"github.com/rs/zerolog"
)

type UserUseCase interface {
	RegisterUser(ctx context.Context, creds *dto.UserCredentials) (*user.User, error)
	ValidateUser(ctx context.Context, creds *dto.UserCredentials) (*user.User, error)
}

type UserUC struct {
	repo repository.UserRepository
	log  *zerolog.Logger
}

func NewUserUC(repo repository.UserRepository, log *zerolog.Logger) *UserUC {
	return &UserUC{
		repo: repo,
		log:  log,
	}
}

func (u *UserUC) RegisterUser(ctx context.Context, creds *dto.UserCredentials) (*user.User, error) {
	_, claims, err := auth.FromContext(ctx)
	if err != nil {
		u.log.Error().Err(err).
			Str("username", creds.Username).
			Msg("failed to get auth user info")

		return nil, e.ErrInternal
	}

	authUserRole := claims.Role
	// Only admins are allowed to create other admin users
	if creds.Role == string(user.RoleAdmin) && authUserRole != string(user.RoleAdmin) {
		u.log.Error().
			Str("username", creds.Username).
			Msg("user is not authorised to register admin user")

		return nil, e.ErrUnauthorized
	}

	usr := user.New(creds.Username, user.Role(creds.Role))

	if err := usr.SetPassword(creds.Password); err != nil {
		u.log.Error().Err(err).
			Str("username", creds.Username).
			Msg("user password generation error")

		return nil, err
	}

	repoUser, err := u.repo.CreateUser(ctx, usr)

	if errors.Is(err, e.ErrExists) {
		return nil, err
	}

	if err != nil {
		u.log.Error().Err(err).
			Str("username", creds.Username).
			Msg("user registration error")

		return nil, e.ErrInternal
	}

	return repoUser, nil
}
