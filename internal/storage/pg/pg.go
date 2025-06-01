package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/patraden/ya-practicum-gophkeeper/internal/domain/errors"
	"github.com/rs/zerolog"
)

type QueryFunc = func(*Queries) error

type DB struct {
	connString string
	ConnPool   ConnenctionPool
	log        *zerolog.Logger
}

func NewDB(connString string, log *zerolog.Logger) *DB {
	return &DB{
		connString: connString,
		ConnPool:   nil,
		log:        log,
	}
}

func (db *DB) WithPool(pool ConnenctionPool) *DB {
	if db.ConnPool != nil {
		db.log.Info().
			Msg("database connection will be replaced")
	}

	db.ConnPool = pool

	return db
}

func (db *DB) Init(ctx context.Context) error {
	if db.ConnPool != nil {
		return nil
	}

	config, err := pgxpool.ParseConfig(db.connString)
	if err != nil {
		db.log.Info().
			Str("conn_string", db.connString).
			Msg("failed to parse pg conn string")

		return errors.ErrDBInit
	}

	config.MaxConns = 30

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		db.log.Info().
			Str("conn_string", config.ConnString()).
			Msg("failed to create connection pool")

		return errors.ErrDBInit
	}

	db.ConnPool = pool
	db.log.Info().
		Msg("database connections pool initialized")

	return nil
}

func (db *DB) Ping(ctx context.Context) error {
	if db.ConnPool == nil {
		db.log.Error().
			Msg("pg connection pool is empty")
		return errors.ErrDBInit
	}

	if err := db.ConnPool.Ping(ctx); err != nil {
		db.log.Error().
			Msg("pg is not reachibale")
		return errors.ErrDBConn
	}

	return nil
}

func (db *DB) Close() {
	if db.ConnPool == nil {
		return
	}

	db.ConnPool.Close()
	db.ConnPool = nil

	db.log.Info().Msg("disconnected from database pool")
}
