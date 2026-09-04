package application_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/application"
)

type fakeUserRepo struct {
	createFn        func(ctx context.Context, db identity.DBExecutor, user identity.User) error
	existsByEmailFn func(ctx context.Context, db identity.DBExecutor, email string) (bool, error)
}

func (r *fakeUserRepo) Create(ctx context.Context, db identity.DBExecutor, user identity.User) error {
	return r.createFn(ctx, db, user)
}
func (r *fakeUserRepo) FindByID(ctx context.Context, db identity.DBExecutor, id identity.UserID) (identity.User, error) {
	return identity.User{}, nil
}
func (r *fakeUserRepo) FindByEmail(ctx context.Context, db identity.DBExecutor, email string) (identity.User, error) {
	return identity.User{}, nil
}
func (r *fakeUserRepo) ExistsByEmail(ctx context.Context, db identity.DBExecutor, email string) (bool, error) {
	return r.existsByEmailFn(ctx, db, email)
}

type fakeCredentialRepo struct {
	createFn func(ctx context.Context, db identity.DBExecutor, cred identity.Credential) error
}

func (r *fakeCredentialRepo) Create(ctx context.Context, db identity.DBExecutor, cred identity.Credential) error {
	return r.createFn(ctx, db, cred)
}
func (r *fakeCredentialRepo) FindByUserID(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.Credential, error) {
	return identity.Credential{}, nil
}

type fakeOrgRepo struct {
	createFn    func(ctx context.Context, db identity.DBExecutor, org identity.Organization) error
	existsAnyFn func(ctx context.Context, db identity.DBExecutor) (bool, error)
}

func (r *fakeOrgRepo) Create(ctx context.Context, db identity.DBExecutor, org identity.Organization) error {
	return r.createFn(ctx, db, org)
}
func (r *fakeOrgRepo) FindByID(ctx context.Context, db identity.DBExecutor, id identity.OrganizationID) (identity.Organization, error) {
	return identity.Organization{}, nil
}
func (r *fakeOrgRepo) FindBySlug(ctx context.Context, db identity.DBExecutor, slug string) (identity.Organization, error) {
	return identity.Organization{}, nil
}
func (r *fakeOrgRepo) ExistsAny(ctx context.Context, db identity.DBExecutor) (bool, error) {
	return r.existsAnyFn(ctx, db)
}

type fakeMembershipRepo struct {
	createFn func(ctx context.Context, db identity.DBExecutor, m identity.Membership) error
}

func (r *fakeMembershipRepo) Create(ctx context.Context, db identity.DBExecutor, m identity.Membership) error {
	return r.createFn(ctx, db, m)
}
func (r *fakeMembershipRepo) FindByUserAndOrg(ctx context.Context, db identity.DBExecutor, userID identity.UserID, orgID identity.OrganizationID) (identity.Membership, error) {
	return identity.Membership{}, nil
}
func (r *fakeMembershipRepo) FindByUser(ctx context.Context, db identity.DBExecutor, userID identity.UserID) ([]identity.Membership, error) {
	return nil, nil
}

type fakePlatformGrantRepo struct {
	createFn func(ctx context.Context, db identity.DBExecutor, g identity.PlatformGrant) error
}

func (r *fakePlatformGrantRepo) Create(ctx context.Context, db identity.DBExecutor, g identity.PlatformGrant) error {
	return r.createFn(ctx, db, g)
}
func (r *fakePlatformGrantRepo) FindByUser(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.PlatformGrant, error) {
	return identity.PlatformGrant{}, nil
}

type fakeHasher struct {
	hashFn func(password []byte) (identity.PasswordHash, error)
}

func (h *fakeHasher) Hash(password []byte) (identity.PasswordHash, error) {
	return h.hashFn(password)
}
func (h *fakeHasher) Verify(password []byte, encodedHash identity.PasswordHash) error {
	return nil
}

type fakeNormalizer struct {
	normalizeFn func(email string) (string, error)
}

func (n *fakeNormalizer) Normalize(email string) (string, error) {
	return n.normalizeFn(email)
}

type fakeDBExecutor struct{}

func (e *fakeDBExecutor) Exec(ctx context.Context, sql string, arguments ...any) (int64, error) {
	return 0, nil
}
func (e *fakeDBExecutor) Query(ctx context.Context, sql string, args ...any) (identity.Rows, error) {
	return nil, nil
}
func (e *fakeDBExecutor) QueryRow(ctx context.Context, sql string, args ...any) identity.RowScanner {
	return nil
}

type fakeTxManager struct {
	fn func(ctx context.Context, fn func(ctx context.Context, db identity.DBExecutor) error) error
}

func (m *fakeTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context, db identity.DBExecutor) error) error {
	return m.fn(ctx, fn)
}

func defaultServices() (*application.CanonicalSetupService, *fakeUserRepo, *fakeCredentialRepo, *fakeOrgRepo, *fakeMembershipRepo, *fakePlatformGrantRepo) {
	users := &fakeUserRepo{}
	creds := &fakeCredentialRepo{}
	orgs := &fakeOrgRepo{}
	memberships := &fakeMembershipRepo{}
	grants := &fakePlatformGrantRepo{}
	hasher := &fakeHasher{hashFn: func(p []byte) (identity.PasswordHash, error) {
		return "hashed:" + identity.PasswordHash(p), nil
	}}
	normalizer := &fakeNormalizer{normalizeFn: func(e string) (string, error) {
		return e, nil
	}}

	svc := application.NewCanonicalSetupService(users, creds, orgs, memberships, grants, hasher, normalizer)
	return svc, users, creds, orgs, memberships, grants
}

func buildCmd() application.InitialOwnerSetupCommand {
	return application.InitialOwnerSetupCommand{
		DisplayName:      "Owner",
		Email:            "owner@example.com",
		Password:         "ValidP@ss1",
		OrganizationName: "My Org",
		OrganizationSlug: "my-org",
	}
}

func TestSuccessfulInitialization(t *testing.T) {
	svc, users, creds, orgs, memberships, grants := defaultServices()

	users.existsByEmailFn = func(ctx context.Context, db identity.DBExecutor, email string) (bool, error) {
		return false, nil
	}
	orgs.existsAnyFn = func(ctx context.Context, db identity.DBExecutor) (bool, error) {
		return false, nil
	}

	userCreated := false
	users.createFn = func(ctx context.Context, db identity.DBExecutor, user identity.User) error {
		userCreated = true
		return nil
	}
	credCreated := false
	creds.createFn = func(ctx context.Context, db identity.DBExecutor, cred identity.Credential) error {
		credCreated = true
		return nil
	}
	orgCreated := false
	orgs.createFn = func(ctx context.Context, db identity.DBExecutor, org identity.Organization) error {
		orgCreated = true
		return nil
	}
	memCreated := false
	memberships.createFn = func(ctx context.Context, db identity.DBExecutor, m identity.Membership) error {
		memCreated = true
		return nil
	}
	grantCreated := false
	grants.createFn = func(ctx context.Context, db identity.DBExecutor, g identity.PlatformGrant) error {
		grantCreated = true
		return nil
	}

	result, err := svc.Execute(context.Background(), &fakeDBExecutor{}, buildCmd())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID.IsZero() || result.OrganizationID.IsZero() {
		t.Error("expected non-zero IDs in result")
	}
	if !userCreated || !credCreated || !orgCreated || !memCreated || !grantCreated {
		t.Errorf("not all entities created: user=%v cred=%v org=%v mem=%v grant=%v",
			userCreated, credCreated, orgCreated, memCreated, grantCreated)
	}
}

func TestAlreadyInitialized(t *testing.T) {
	svc, _, _, orgs, _, _ := defaultServices()
	orgs.existsAnyFn = func(ctx context.Context, db identity.DBExecutor) (bool, error) {
		return true, nil
	}

	_, err := svc.Execute(context.Background(), &fakeDBExecutor{}, buildCmd())
	if !errors.Is(err, identity.ErrAlreadyInitialized) {
		t.Errorf("expected ErrAlreadyInitialized, got %v", err)
	}
}

func TestDuplicateEmail(t *testing.T) {
	svc, users, _, orgs, _, _ := defaultServices()
	orgs.existsAnyFn = func(ctx context.Context, db identity.DBExecutor) (bool, error) {
		return false, nil
	}
	users.existsByEmailFn = func(ctx context.Context, db identity.DBExecutor, email string) (bool, error) {
		return true, nil
	}

	_, err := svc.Execute(context.Background(), &fakeDBExecutor{}, buildCmd())
	if !errors.Is(err, identity.ErrDuplicateEmail) {
		t.Errorf("expected ErrDuplicateEmail, got %v", err)
	}
}

func TestInvalidPassword(t *testing.T) {
	svc, _, _, orgs, _, _ := defaultServices()
	orgs.existsAnyFn = func(ctx context.Context, db identity.DBExecutor) (bool, error) {
		return false, nil
	}

	cmd := buildCmd()
	cmd.Password = "short"

	_, err := svc.Execute(context.Background(), &fakeDBExecutor{}, cmd)
	if !errors.Is(err, identity.ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestInvalidEmail(t *testing.T) {
	_, _, _, orgs, _, _ := defaultServices()
	orgs.existsAnyFn = func(ctx context.Context, db identity.DBExecutor) (bool, error) {
		return false, nil
	}

	normalizer := &fakeNormalizer{normalizeFn: func(e string) (string, error) {
		return "", identity.ErrInvalidEmail
	}}

	svc2 := application.NewCanonicalSetupService(
		&fakeUserRepo{existsByEmailFn: func(ctx context.Context, db identity.DBExecutor, email string) (bool, error) { return false, nil }},
		&fakeCredentialRepo{},
		orgs,
		&fakeMembershipRepo{},
		&fakePlatformGrantRepo{},
		&fakeHasher{hashFn: func(p []byte) (identity.PasswordHash, error) { return "h", nil }},
		normalizer,
	)

	_, err := svc2.Execute(context.Background(), &fakeDBExecutor{}, buildCmd())
	if !errors.Is(err, identity.ErrInvalidEmail) {
		t.Errorf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestTransactionalSetup_CommitsOnSuccess(t *testing.T) {
	svc, users, creds, orgs, memberships, grants := defaultServices()

	users.existsByEmailFn = func(ctx context.Context, db identity.DBExecutor, email string) (bool, error) { return false, nil }
	orgs.existsAnyFn = func(ctx context.Context, db identity.DBExecutor) (bool, error) { return false, nil }

	var captured []string
	users.createFn = func(ctx context.Context, db identity.DBExecutor, user identity.User) error {
		captured = append(captured, "user")
		return nil
	}
	creds.createFn = func(ctx context.Context, db identity.DBExecutor, cred identity.Credential) error {
		captured = append(captured, "cred")
		return nil
	}
	orgs.createFn = func(ctx context.Context, db identity.DBExecutor, org identity.Organization) error {
		captured = append(captured, "org")
		return nil
	}
	memberships.createFn = func(ctx context.Context, db identity.DBExecutor, m identity.Membership) error {
		captured = append(captured, "membership")
		return nil
	}
	grants.createFn = func(ctx context.Context, db identity.DBExecutor, g identity.PlatformGrant) error {
		captured = append(captured, "grant")
		return nil
	}

	txManager := &fakeTxManager{fn: func(ctx context.Context, fn func(ctx context.Context, db identity.DBExecutor) error) error {
		return fn(ctx, &fakeDBExecutor{})
	}}
	txSvc := application.NewTransactionalSetupService(svc, txManager)

	result, err := txSvc.Execute(context.Background(), buildCmd())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID.IsZero() {
		t.Error("expected user ID")
	}
	if len(captured) != 5 {
		t.Errorf("expected 5 entities created, got %d: %v", len(captured), captured)
	}
}

func TestTransactionalSetup_RollbackOnFailure(t *testing.T) {
	_, users, creds, _, _, _ := defaultServices()

	users.existsByEmailFn = func(ctx context.Context, db identity.DBExecutor, email string) (bool, error) { return false, nil }
	orgs := &fakeOrgRepo{
		existsAnyFn: func(ctx context.Context, db identity.DBExecutor) (bool, error) { return false, nil },
		createFn: func(ctx context.Context, db identity.DBExecutor, org identity.Organization) error {
			return errors.New("org create failed")
		},
	}

	users.createFn = func(ctx context.Context, db identity.DBExecutor, user identity.User) error { return nil }
	creds.createFn = func(ctx context.Context, db identity.DBExecutor, cred identity.Credential) error { return nil }

	svc2 := application.NewCanonicalSetupService(
		users,
		creds,
		orgs,
		&fakeMembershipRepo{},
		&fakePlatformGrantRepo{},
		&fakeHasher{hashFn: func(p []byte) (identity.PasswordHash, error) { return "h", nil }},
		&fakeNormalizer{normalizeFn: func(e string) (string, error) { return e, nil }},
	)

	txManager := &fakeTxManager{fn: func(ctx context.Context, fn func(ctx context.Context, db identity.DBExecutor) error) error {
		return fn(ctx, &fakeDBExecutor{})
	}}
	txSvc := application.NewTransactionalSetupService(svc2, txManager)

	_, err := txSvc.Execute(context.Background(), buildCmd())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTransactionalSetup_TxManagerError(t *testing.T) {
	svc, _, _, _, _, _ := defaultServices()

	txManager := &fakeTxManager{fn: func(ctx context.Context, fn func(ctx context.Context, db identity.DBExecutor) error) error {
		return errors.New("tx begin failed")
	}}
	txSvc := application.NewTransactionalSetupService(svc, txManager)

	_, err := txSvc.Execute(context.Background(), buildCmd())
	if err == nil {
		t.Fatal("expected error from tx manager")
	}
}

func TestAtomicity_PartialWriteRollback(t *testing.T) {
	var mu sync.Mutex
	var created []string

	users := &fakeUserRepo{
		existsByEmailFn: func(ctx context.Context, db identity.DBExecutor, email string) (bool, error) { return false, nil },
		createFn: func(ctx context.Context, db identity.DBExecutor, user identity.User) error {
			mu.Lock()
			created = append(created, "user")
			mu.Unlock()
			return nil
		},
	}
	creds := &fakeCredentialRepo{
		createFn: func(ctx context.Context, db identity.DBExecutor, cred identity.Credential) error {
			mu.Lock()
			created = append(created, "cred")
			mu.Unlock()
			return nil
		},
	}
	orgs := &fakeOrgRepo{
		existsAnyFn: func(ctx context.Context, db identity.DBExecutor) (bool, error) { return false, nil },
		createFn: func(ctx context.Context, db identity.DBExecutor, org identity.Organization) error {
			mu.Lock()
			created = append(created, "org")
			mu.Unlock()
			return errors.New("org failure")
		},
	}
	memberships := &fakeMembershipRepo{
		createFn: func(ctx context.Context, db identity.DBExecutor, m identity.Membership) error {
			mu.Lock()
			created = append(created, "membership")
			mu.Unlock()
			return nil
		},
	}
	grants := &fakePlatformGrantRepo{
		createFn: func(ctx context.Context, db identity.DBExecutor, g identity.PlatformGrant) error {
			mu.Lock()
			created = append(created, "grant")
			mu.Unlock()
			return nil
		},
	}

	svc := application.NewCanonicalSetupService(
		users, creds, orgs, memberships, grants,
		&fakeHasher{hashFn: func(p []byte) (identity.PasswordHash, error) { return "h", nil }},
		&fakeNormalizer{normalizeFn: func(e string) (string, error) { return e, nil }},
	)

	_, err := svc.Execute(context.Background(), &fakeDBExecutor{}, buildCmd())
	if err == nil {
		t.Fatal("expected error")
	}

	if len(created) > 0 {
		for _, c := range created {
			if c == "membership" || c == "grant" {
				t.Errorf("entity %q should not have been created after org failure", c)
			}
		}
	}
}
