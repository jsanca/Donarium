package properties

import "context"

type Repository interface {
	Create(ctx context.Context, db DBExecutor, p Property) error
	FindByID(ctx context.Context, db DBExecutor, id PropertyID) (Property, error)
	FindAccessibleByUser(ctx context.Context, db DBExecutor, userID string) ([]Property, error)
}

type StakeholderRepository interface {
	Create(ctx context.Context, db DBExecutor, s PropertyStakeholder) error
	FindByProperty(ctx context.Context, db DBExecutor, propertyID PropertyID) ([]PropertyStakeholder, error)
	FindAccessiblePropertyIDs(ctx context.Context, db DBExecutor, userID string) ([]PropertyID, error)
	HasAccess(ctx context.Context, db DBExecutor, propertyID PropertyID, userID string) (bool, error)
}
