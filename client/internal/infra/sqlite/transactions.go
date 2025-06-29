package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

// WithinTrx wraps a query function in a db transaction using the provided connection and options.
// If the function returns an error, the transaction is rolled back.
// If the function succeeds, the transaction is committed.
// The original error is preserved even if rollback also fails.
func WithinTrx(
	ctx context.Context,
	db *sql.DB,
	trxOptions *sql.TxOptions,
	queryfn QueryFunc,
) QueryFunc {
	return func(queries *Queries) (err error) {
		trx, beginErr := db.BeginTx(ctx, trxOptions)
		if beginErr != nil {
			err = beginErr

			return
		}

		defer func() {
			if err != nil {
				if rollbackErr := trx.Rollback(); rollbackErr != nil {
					err = fmt.Errorf("query failed: %w, rollback failed: %w", err, rollbackErr)
				}
			}
		}()

		if err = queryfn(queries.WithTx(trx)); err != nil {
			return err
		}

		if commitErr := trx.Commit(); commitErr != nil {
			return fmt.Errorf("commit failed: %w", commitErr)
		}

		return nil
	}
}
