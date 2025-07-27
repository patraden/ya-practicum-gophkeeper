//nolint:funlen // reason: testing internal logic and long test functions are acceptable
package pg_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/pg"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name          string
	queryFn       pg.QueryFunc
	mockSetup     func(pool pgxmock.PgxPoolIface)
	expectedError error
}

func TestWithTransaction(t *testing.T) {
	t.Parallel()

	tests := []testCase{
		{
			name:          "success",
			queryFn:       func(*pg.Queries) error { return nil },
			mockSetup:     func(p pgxmock.PgxPoolIface) { p.ExpectBegin(); p.ExpectCommit() },
			expectedError: nil,
		},
		{
			name:          "query function fails",
			queryFn:       func(*pg.Queries) error { return e.ErrInternal },
			mockSetup:     func(p pgxmock.PgxPoolIface) { p.ExpectBegin(); p.ExpectRollback() },
			expectedError: e.ErrInternal,
		},
		{
			name:          "begin transaction fails",
			queryFn:       func(*pg.Queries) error { return nil },
			mockSetup:     func(p pgxmock.PgxPoolIface) { p.ExpectBegin().WillReturnError(e.ErrInternal) },
			expectedError: e.ErrInternal,
		},
		{
			name:    "commit transaction fails",
			queryFn: func(*pg.Queries) error { return nil },
			mockSetup: func(p pgxmock.PgxPoolIface) {
				p.ExpectBegin()
				p.ExpectCommit().WillReturnError(e.ErrInternal)
				p.ExpectRollback()
			},
			expectedError: e.ErrInternal,
		},
		{
			name:    "rollback transaction fails",
			queryFn: func(*pg.Queries) error { return e.ErrInternal },
			mockSetup: func(p pgxmock.PgxPoolIface) {
				p.ExpectBegin()
				p.ExpectRollback().WillReturnError(e.ErrInternal)
			},
			expectedError: e.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)

			tt.mockSetup(mockPool)
			defer mockPool.Close()

			trxQueryFn := pg.WithinTrx(context.Background(), mockPool, pgx.TxOptions{}, tt.queryFn)
			queries := pg.New(mockPool)

			err = trxQueryFn(queries)
			require.ErrorIs(t, err, tt.expectedError)
			err = mockPool.ExpectationsWereMet()
			require.NoError(t, err)
		})
	}
}
