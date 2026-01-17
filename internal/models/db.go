package models

import (
	"context"

	"github.com/cdriehuys/stuff2/internal/models/queries"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB interface {
	queries.DBTX

	Begin(context.Context) (Transaction, error)
}

type Transaction interface {
	queries.DBTX

	Commit(context.Context) error
	Rollback(context.Context) error
}

type PoolWrapper struct {
	*pgxpool.Pool
}

func (w PoolWrapper) Begin(ctx context.Context) (Transaction, error) {
	return w.Pool.Begin(ctx)
}
