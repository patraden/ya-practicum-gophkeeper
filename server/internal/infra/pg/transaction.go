package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// WithinTrx wraps a query function in a db transaction using the provided connection pool and options.
// If the function returns an error, the transaction is rolled back.
// If the function succeeds, the transaction is committed.
// The original error is preserved even if rollback also fails.
func WithinTrx(
	ctx context.Context,
	connPool ConnenctionPool,
	trxOptions pgx.TxOptions,
	queryfn QueryFunc,
) QueryFunc {
	return func(queries *Queries) (err error) {
		trx, beginErr := connPool.BeginTx(ctx, trxOptions)
		if beginErr != nil {
			err = beginErr

			return
		}

		defer func() {
			if err != nil {
				if rollbackErr := trx.Rollback(ctx); rollbackErr != nil {
					err = fmt.Errorf("query failed: %w, rollback failed: %w", err, rollbackErr)
				}
			}
		}()

		if err = queryfn(queries.WithTx(trx)); err != nil {
			return err
		}

		if commitErr := trx.Commit(ctx); commitErr != nil {
			return fmt.Errorf("commit failed: %w", commitErr)
		}

		return nil
	}
}
