package pg_test

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/pg"
	"github.com/stretchr/testify/require"
)

func TestDBInitAndPingSuccess(t *testing.T) {
	t.Parallel()

	dsn := "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable"
	ctx := context.Background()
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)

	mockPool.ExpectPing()

	db := pg.NewDB(dsn)
	defer db.Close()

	err = db.Init(ctx)
	require.NoError(t, err)

	db = db.WithPool(mockPool)
	err = db.Ping(ctx)
	require.NoError(t, err)
}

func TestDBInitFailureBadDSN(t *testing.T) {
	t.Parallel()

	dsn := "bad_dsn"
	ctx := context.Background()

	db := pg.NewDB(dsn)
	defer db.Close()

	err := db.Init(ctx)
	require.ErrorIs(t, err, e.ErrParse)

	err = db.Ping(ctx)
	require.ErrorIs(t, err, e.ErrNotReady)
}

func TestDBPingFailure(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	dsn := "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable"
	ctx := context.Background()

	require.NoError(t, err)

	mockPool.ExpectPing().WillReturnError(e.ErrInternal)

	db := pg.NewDB(dsn)
	defer db.Close()

	err = db.Init(ctx)
	require.NoError(t, err)

	db = db.WithPool(mockPool)
	err = db.Ping(ctx)
	require.ErrorIs(t, err, e.ErrUnavailable)
}

// Test replacing the connection pool using WithPool.
func TestDBWithPool(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	dsn := "postgres://fake-dsn"

	require.NoError(t, err)

	db := pg.NewDB(dsn)
	db = db.WithPool(mockPool)

	require.Equal(t, mockPool, db.ConnPool, "connection pool should be set correctly")
	db.Close()
	require.Nil(t, db.ConnPool, "connection pool should be nil after closing")
}

func TestDBCloseWithoutInit(t *testing.T) {
	t.Parallel()

	dsn := "postgres://fake-dsn"
	db := pg.NewDB(dsn)

	require.NotPanics(t, func() {
		db.Close()
	}, "calling Close on an uninitialized database should not panic")
}
