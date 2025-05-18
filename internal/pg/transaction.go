package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// query function wrapper into pg database transaction.
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
					err = rollbackErr

					return
				}
			}
		}()

		if fnErr := queryfn(queries.WithTx(trx)); fnErr != nil {
			return fnErr
		}

		err = trx.Commit(ctx)

		return
	}
}
