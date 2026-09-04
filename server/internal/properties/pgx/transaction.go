package pgx

import (
	"context"
	"fmt"

	"donarium/server/internal/properties"
	"donarium/server/internal/properties/application"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionManager struct {
	pool *pgxpool.Pool
}

func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{pool: pool}
}

func (m *TransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context, db properties.DBExecutor) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	db := NewExecutorFromPool(tx)
	if err = fn(ctx, db); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

var _ application.TransactionManager = (*TransactionManager)(nil)
