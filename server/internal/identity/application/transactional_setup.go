package application

import (
	"context"

	"donarium/server/internal/identity"
)

type TransactionalSetupService struct {
	canonical *CanonicalSetupService
	txManager TransactionManager
}

func NewTransactionalSetupService(canonical *CanonicalSetupService, txManager TransactionManager) *TransactionalSetupService {
	return &TransactionalSetupService{
		canonical: canonical,
		txManager: txManager,
	}
}

func (s *TransactionalSetupService) Execute(ctx context.Context, cmd InitialOwnerSetupCommand) (InitialOwnerSetupResult, error) {
	var result InitialOwnerSetupResult
	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context, db identity.DBExecutor) error {
		var innerErr error
		result, innerErr = s.canonical.Execute(ctx, db, cmd)
		return innerErr
	})
	return result, err
}
