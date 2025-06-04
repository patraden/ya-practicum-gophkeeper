package retry

import (
	"context"

	"github.com/cenkalti/backoff/v4"
	"github.com/rs/zerolog"
)

type Retryable func(err error) bool

func WithRetry(
	ctx context.Context,
	boff backoff.BackOff,
	log *zerolog.Logger,
	retriable Retryable,
	op func() error,
) error {
	operation := func() error {
		err := op()
		if err == nil {
			return nil
		}

		if retriable(err) {
			log.Info().Err(err).
				Msg("retrying after error")

			return err
		}

		return backoff.Permanent(err)
	}

	return backoff.Retry(operation, backoff.WithContext(boff, ctx))
}
