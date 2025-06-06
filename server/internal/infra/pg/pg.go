package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
)

type QueryFunc = func(*Queries) error

type DB struct {
	connString string
	ConnPool   ConnenctionPool
}

func NewDB(connString string) *DB {
	return &DB{
		connString: connString,
		ConnPool:   nil,
	}
}

func (db *DB) WithPool(pool ConnenctionPool) *DB {
	db.ConnPool = pool

	return db
}

func (db *DB) Init(ctx context.Context) error {
	if db.ConnPool != nil {
		return nil
	}

	config, err := pgxpool.ParseConfig(db.connString)
	if err != nil {
		return errors.ErrParse
	}

	config.MaxConns = 30

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return errors.ErrInvalidInput
	}

	db.ConnPool = pool

	return nil
}

func (db *DB) Ping(ctx context.Context) error {
	if db.ConnPool == nil {
		return errors.ErrNotReady
	}

	if err := db.ConnPool.Ping(ctx); err != nil {
		return errors.ErrUnavailable
	}

	return nil
}

func (db *DB) Close() {
	if db.ConnPool == nil {
		return
	}

	db.ConnPool.Close()
	db.ConnPool = nil
}
