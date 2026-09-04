package application

import (
	"context"

	"donarium/server/internal/identity"
)

type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context, db identity.DBExecutor) error) error
}
