package repository

import (
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/infra/sqlite"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
)

// FromSQLUser maps a sqlite.User (returned by sqlc) to a domain-level User model.
func FromSQLUser(usql sqlite.User) (*user.User, error) {
	usr, err := user.NewWithID(usql.ID, usql.Username, usql.Role)
	if err != nil {
		return nil, e.InternalErr(err)
	}

	usr.BucketName = usql.Bucketname
	usr.CreatedAt = usql.CreatedAt
	usr.UpdatedAt = usql.UpdatedAt
	usr.Salt = usql.Salt
	usr.Verifier = usql.Verifier

	return usr, nil
}
