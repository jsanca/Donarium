package authentication_test

import (
	"context"
	"testing"
	"time"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/application/authentication"

	"github.com/google/uuid"
)

func TestResolve_DefaultContextIsEarliestMembership(t *testing.T) {
	userID := identity.NewUserID()
	org1ID := identity.NewOrganizationID()
	org2ID := identity.NewOrganizationID()
	org3ID := identity.NewOrganizationID()

	memberships := []identity.Membership{
		{UserID: userID, OrganizationID: org1ID, Role: identity.OrganizationRoleOwner, CreatedAt: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
		{UserID: userID, OrganizationID: org2ID, Role: identity.OrganizationRoleOwner, CreatedAt: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)},
		{UserID: userID, OrganizationID: org3ID, Role: identity.OrganizationRoleOwner, CreatedAt: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)},
	}

	userRepo := &fakeUserRepo{
		findByIDFn: func(ctx context.Context, db identity.DBExecutor, id identity.UserID) (identity.User, error) {
			return identity.User{ID: userID, DisplayName: "Test", Email: "test@donarium.test"}, nil
		},
	}

	grantRepo := &fakeGrantRepo{
		findByUserFn: func(ctx context.Context, db identity.DBExecutor, uid identity.UserID) (identity.PlatformGrant, error) {
			return identity.PlatformGrant{UserID: uid, Role: identity.PlatformRoleSuperAdmin}, nil
		},
	}

	memRepo := &fakeMemRepo{
		findByUserFn: func(ctx context.Context, db identity.DBExecutor, uid identity.UserID) ([]identity.Membership, error) {
			return memberships, nil
		},
	}

	orgRepo := &fakeOrgRepo{
		findByIDFn: func(ctx context.Context, db identity.DBExecutor, id identity.OrganizationID) (identity.Organization, error) {
			switch id {
			case org1ID:
				return identity.Organization{ID: org1ID, Name: "First Org", Slug: "first-org"}, nil
			case org2ID:
				return identity.Organization{ID: org2ID, Name: "Second Org", Slug: "second-org"}, nil
			case org3ID:
				return identity.Organization{ID: org3ID, Name: "Third Org", Slug: "third-org"}, nil
			}
			return identity.Organization{}, identity.ErrUserNotFound
		},
	}

	resolver := authentication.NewPrincipalResolverService(userRepo, grantRepo, memRepo, orgRepo)
	principal, err := resolver.Resolve(context.Background(), &fakeDB{}, userID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if principal.DefaultContext.Type != "organization" {
		t.Fatalf("expected default context type 'organization', got %q", principal.DefaultContext.Type)
	}

	if principal.DefaultContext.OrganizationID != uuid.UUID(org1ID).String() {
		t.Errorf("expected default context to be org1 (earliest), got %q", principal.DefaultContext.OrganizationID)
	}

	if principal.DefaultContext.Role != string(identity.OrganizationRoleOwner) {
		t.Errorf("expected default context role 'owner', got %q", principal.DefaultContext.Role)
	}

	if len(principal.OrganizationContexts) != 3 {
		t.Errorf("expected 3 organization contexts, got %d", len(principal.OrganizationContexts))
	}
}

func TestResolve_DefaultContextDeterministicWithSameCreationDate(t *testing.T) {
	userID := identity.NewUserID()
	sameTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	org1ID := identity.NewOrganizationID()
	org2ID := identity.NewOrganizationID()

	memberships := []identity.Membership{
		{UserID: userID, OrganizationID: org1ID, Role: identity.OrganizationRoleOwner, CreatedAt: sameTime},
		{UserID: userID, OrganizationID: org2ID, Role: identity.OrganizationRoleOwner, CreatedAt: sameTime},
	}

	userRepo := &fakeUserRepo{
		findByIDFn: func(ctx context.Context, db identity.DBExecutor, id identity.UserID) (identity.User, error) {
			return identity.User{ID: userID, DisplayName: "Test", Email: "test@donarium.test"}, nil
		},
	}

	grantRepo := &fakeGrantRepo{
		findByUserFn: func(ctx context.Context, db identity.DBExecutor, uid identity.UserID) (identity.PlatformGrant, error) {
			return identity.PlatformGrant{UserID: uid, Role: identity.PlatformRoleSuperAdmin}, nil
		},
	}

	memRepo := &fakeMemRepo{
		findByUserFn: func(ctx context.Context, db identity.DBExecutor, uid identity.UserID) ([]identity.Membership, error) {
			return memberships, nil
		},
	}

	orgRepo := &fakeOrgRepo{
		findByIDFn: func(ctx context.Context, db identity.DBExecutor, id identity.OrganizationID) (identity.Organization, error) {
			return identity.Organization{ID: id, Name: "Org", Slug: "org"}, nil
		},
	}

	resolver := authentication.NewPrincipalResolverService(userRepo, grantRepo, memRepo, orgRepo)

	var firstOrgID string
	for i := 0; i < 5; i++ {
		principal, err := resolver.Resolve(context.Background(), &fakeDB{}, userID)
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		if i == 0 {
			firstOrgID = principal.DefaultContext.OrganizationID
		} else if principal.DefaultContext.OrganizationID != firstOrgID {
			t.Errorf("iteration %d: expected default org %s, got %s", i, firstOrgID, principal.DefaultContext.OrganizationID)
		}
	}
}

func TestResolve_DefaultContextIsPlatformWhenNoOrganizations(t *testing.T) {
	userID := identity.NewUserID()

	userRepo := &fakeUserRepo{
		findByIDFn: func(ctx context.Context, db identity.DBExecutor, id identity.UserID) (identity.User, error) {
			return identity.User{ID: userID, DisplayName: "Test", Email: "test@donarium.test"}, nil
		},
	}

	grantRepo := &fakeGrantRepo{
		findByUserFn: func(ctx context.Context, db identity.DBExecutor, uid identity.UserID) (identity.PlatformGrant, error) {
			return identity.PlatformGrant{UserID: uid, Role: identity.PlatformRoleSuperAdmin}, nil
		},
	}

	memRepo := &fakeMemRepo{
		findByUserFn: func(ctx context.Context, db identity.DBExecutor, uid identity.UserID) ([]identity.Membership, error) {
			return []identity.Membership{}, nil
		},
	}

	orgRepo := &fakeOrgRepo{}

	resolver := authentication.NewPrincipalResolverService(userRepo, grantRepo, memRepo, orgRepo)
	principal, err := resolver.Resolve(context.Background(), &fakeDB{}, userID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if principal.DefaultContext.Type != "platform" {
		t.Errorf("expected default context type 'platform', got %q", principal.DefaultContext.Type)
	}

	if principal.DefaultContext.Role != string(identity.PlatformRoleSuperAdmin) {
		t.Errorf("expected role 'super_admin', got %q", principal.DefaultContext.Role)
	}
}

func TestResolve_DefaultContextIsFirstOrg_SingleMembership(t *testing.T) {
	userID := identity.NewUserID()
	orgID := identity.NewOrganizationID()

	memberships := []identity.Membership{
		{UserID: userID, OrganizationID: orgID, Role: identity.OrganizationRoleOwner, CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	userRepo := &fakeUserRepo{
		findByIDFn: func(ctx context.Context, db identity.DBExecutor, id identity.UserID) (identity.User, error) {
			return identity.User{ID: userID, DisplayName: "Test", Email: "test@donarium.test"}, nil
		},
	}

	grantRepo := &fakeGrantRepo{
		findByUserFn: func(ctx context.Context, db identity.DBExecutor, uid identity.UserID) (identity.PlatformGrant, error) {
			return identity.PlatformGrant{UserID: uid, Role: identity.PlatformRoleSuperAdmin}, nil
		},
	}

	memRepo := &fakeMemRepo{
		findByUserFn: func(ctx context.Context, db identity.DBExecutor, uid identity.UserID) ([]identity.Membership, error) {
			return memberships, nil
		},
	}

	orgRepo := &fakeOrgRepo{
		findByIDFn: func(ctx context.Context, db identity.DBExecutor, id identity.OrganizationID) (identity.Organization, error) {
			return identity.Organization{ID: orgID, Name: "Only Org", Slug: "only-org"}, nil
		},
	}

	resolver := authentication.NewPrincipalResolverService(userRepo, grantRepo, memRepo, orgRepo)
	principal, err := resolver.Resolve(context.Background(), &fakeDB{}, userID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if principal.DefaultContext.Type != "organization" {
		t.Fatalf("expected 'organization', got %q", principal.DefaultContext.Type)
	}

	if principal.DefaultContext.OrganizationID != uuid.UUID(orgID).String() {
		t.Errorf("expected org %s, got %s", uuid.UUID(orgID).String(), principal.DefaultContext.OrganizationID)
	}
}
