package identity

import "context"

type DBExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (int64, error)
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) RowScanner
}

type RowScanner interface {
	Scan(dest ...any) error
}

type Rows interface {
	RowScanner
	Next() bool
	Close()
	Err() error
}
