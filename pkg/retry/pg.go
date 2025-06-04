package retry

import (
	"context"
	"errors"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog"
)

func RetriablePG(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case
			pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure,
			pgerrcode.CannotConnectNow,
			pgerrcode.SQLClientUnableToEstablishSQLConnection,
			pgerrcode.TransactionResolutionUnknown:
			return true
		}
	}

	return false
}

func PG(
	ctx context.Context,
	boff backoff.BackOff,
	log *zerolog.Logger,
	op func() error,
) error {
	return WithRetry(ctx, boff, log, RetriablePG, op)
}
