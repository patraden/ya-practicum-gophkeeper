//nolint:funlen // reason: long test functions are acceptable
package pg_test

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/pg"
	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	t.Parallel()

	t.Run("Init and Ping Success", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockPool, err := pgxmock.NewPool()
		require.NoError(t, err)

		mockPool.ExpectPing()

		// Test config parsing (won't use pool from NewDB)
		db, err := pg.NewDB(ctx, "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable")
		require.NoError(t, err)
		defer db.Close()

		// Override with mock for Ping
		db, err = pg.DBWithPool(mockPool)
		require.NoError(t, err)

		err = db.Ping(ctx)
		require.NoError(t, err)
	})

	t.Run("Init Failure with Bad DSN", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db, err := pg.NewDB(ctx, "bad_dsn")
		require.ErrorIs(t, err, e.ErrParse)
		require.Nil(t, db)
	})

	t.Run("Ping Failure with Unreachable DB", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		mockPool, err := pgxmock.NewPool()
		require.NoError(t, err)

		mockPool.ExpectPing().WillReturnError(e.ErrInternal)

		db, err := pg.DBWithPool(mockPool)
		require.NoError(t, err)
		defer db.Close()

		err = db.Ping(ctx)
		require.ErrorIs(t, err, e.ErrUnavailable)
	})

	t.Run("WithPool replaces connection pool and Close clears it", func(t *testing.T) {
		t.Parallel()

		mockPool, err := pgxmock.NewPool()
		require.NoError(t, err)

		db, err := pg.DBWithPool(mockPool)
		require.NoError(t, err)
		require.Equal(t, mockPool, db.ConnPool)

		db.Close()
		require.Nil(t, db.ConnPool)
	})

	t.Run("WithPool with nil pool returns error", func(t *testing.T) {
		t.Parallel()

		db, err := pg.DBWithPool(nil)
		require.ErrorIs(t, err, e.ErrInvalidInput)
		require.Nil(t, db)
	})

	t.Run("Ping after db close returns error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		db, err := pg.NewDB(ctx, "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable")
		require.NoError(t, err)

		defer db.Close()

		db.Close()
		err = db.Ping(ctx)
		require.ErrorIs(t, err, e.ErrNotReady)
	})
}
