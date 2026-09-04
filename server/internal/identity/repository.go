package identity

import "context"

type UserRepository interface {
	Create(ctx context.Context, db DBExecutor, user User) error
	FindByID(ctx context.Context, db DBExecutor, id UserID) (User, error)
	FindByEmail(ctx context.Context, db DBExecutor, email string) (User, error)
	ExistsByEmail(ctx context.Context, db DBExecutor, email string) (bool, error)
}

type CredentialRepository interface {
	Create(ctx context.Context, db DBExecutor, cred Credential) error
	FindByUserID(ctx context.Context, db DBExecutor, userID UserID) (Credential, error)
}

type OrganizationRepository interface {
	Create(ctx context.Context, db DBExecutor, org Organization) error
	FindByID(ctx context.Context, db DBExecutor, id OrganizationID) (Organization, error)
	FindBySlug(ctx context.Context, db DBExecutor, slug string) (Organization, error)
	ExistsAny(ctx context.Context, db DBExecutor) (bool, error)
}

type MembershipRepository interface {
	Create(ctx context.Context, db DBExecutor, m Membership) error
	FindByUserAndOrg(ctx context.Context, db DBExecutor, userID UserID, orgID OrganizationID) (Membership, error)
	FindByUser(ctx context.Context, db DBExecutor, userID UserID) ([]Membership, error)
}

type PlatformGrantRepository interface {
	Create(ctx context.Context, db DBExecutor, grant PlatformGrant) error
	FindByUser(ctx context.Context, db DBExecutor, userID UserID) (PlatformGrant, error)
}
