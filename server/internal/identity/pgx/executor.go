package pgx

import (
	"context"

	"donarium/server/internal/identity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type executor struct {
	execFn  func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	queryFn func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	rowFn   func(ctx context.Context, sql string, args ...any) pgx.Row
}

func (e *executor) Exec(ctx context.Context, sql string, arguments ...any) (int64, error) {
	tag, err := e.execFn(ctx, sql, arguments...)
	return tag.RowsAffected(), err
}

func (e *executor) Query(ctx context.Context, sql string, args ...any) (identity.Rows, error) {
	return e.queryFn(ctx, sql, args...)
}

func (e *executor) QueryRow(ctx context.Context, sql string, args ...any) identity.RowScanner {
	return e.rowFn(ctx, sql, args...)
}

// NewExecutorFromPool adapts pgx types (pool or tx) to the domain DBExecutor
// interface. This anti-corruption layer exists because Go does not support
// return-type covariance: QueryRow returning pgx.Row does not satisfy
// QueryRow returning identity.RowScanner even though pgx.Row implements it.
// This keeps repository code free of pgx import dependencies.

func NewExecutorFromPool(pool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}) identity.DBExecutor {
	return &executor{
		execFn:  pool.Exec,
		queryFn: pool.Query,
		rowFn:   pool.QueryRow,
	}
}
