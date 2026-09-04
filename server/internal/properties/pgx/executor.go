package pgx

import (
	"context"

	"donarium/server/internal/properties"

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

func (e *executor) Query(ctx context.Context, sql string, args ...any) (properties.Rows, error) {
	return e.queryFn(ctx, sql, args...)
}

func (e *executor) QueryRow(ctx context.Context, sql string, args ...any) properties.RowScanner {
	return e.rowFn(ctx, sql, args...)
}

func NewExecutorFromPool(pool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}) properties.DBExecutor {
	return &executor{
		execFn:  pool.Exec,
		queryFn: pool.Query,
		rowFn:   pool.QueryRow,
	}
}
