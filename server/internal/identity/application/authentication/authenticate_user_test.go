package authentication_test

import (
	"context"
	"errors"
	"testing"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/application"
	"donarium/server/internal/identity/application/authentication"
)

type fakeUserRepo struct {
	findByEmailFn func(ctx context.Context, db identity.DBExecutor, email string) (identity.User, error)
	findByIDFn    func(ctx context.Context, db identity.DBExecutor, id identity.UserID) (identity.User, error)
}

func (r *fakeUserRepo) FindByEmail(ctx context.Context, db identity.DBExecutor, email string) (identity.User, error) {
	return r.findByEmailFn(ctx, db, email)
}
func (r *fakeUserRepo) Create(ctx context.Context, db identity.DBExecutor, user identity.User) error { return nil }
func (r *fakeUserRepo) FindByID(ctx context.Context, db identity.DBExecutor, id identity.UserID) (identity.User, error) {
	if r.findByIDFn != nil {
		return r.findByIDFn(ctx, db, id)
	}
	return identity.User{}, nil
}
func (r *fakeUserRepo) ExistsByEmail(ctx context.Context, db identity.DBExecutor, email string) (bool, error) {
	return false, nil
}

type fakeCredentialRepo struct {
	findByUserIDFn func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.Credential, error)
}

func (r *fakeCredentialRepo) FindByUserID(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.Credential, error) {
	return r.findByUserIDFn(ctx, db, userID)
}
func (r *fakeCredentialRepo) Create(ctx context.Context, db identity.DBExecutor, cred identity.Credential) error {
	return nil
}

type fakeGrantRepo struct {
	findByUserFn func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.PlatformGrant, error)
}

func (r *fakeGrantRepo) FindByUser(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.PlatformGrant, error) {
	return r.findByUserFn(ctx, db, userID)
}
func (r *fakeGrantRepo) Create(ctx context.Context, db identity.DBExecutor, g identity.PlatformGrant) error {
	return nil
}

type fakeMemRepo struct {
	findByUserFn func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) ([]identity.Membership, error)
}

func (r *fakeMemRepo) FindByUser(ctx context.Context, db identity.DBExecutor, userID identity.UserID) ([]identity.Membership, error) {
	return r.findByUserFn(ctx, db, userID)
}
func (r *fakeMemRepo) Create(ctx context.Context, db identity.DBExecutor, m identity.Membership) error { return nil }
func (r *fakeMemRepo) FindByUserAndOrg(ctx context.Context, db identity.DBExecutor, userID identity.UserID, orgID identity.OrganizationID) (identity.Membership, error) {
	return identity.Membership{}, nil
}

type fakeOrgRepo struct {
	findByIDFn func(ctx context.Context, db identity.DBExecutor, id identity.OrganizationID) (identity.Organization, error)
}

func (r *fakeOrgRepo) FindByID(ctx context.Context, db identity.DBExecutor, id identity.OrganizationID) (identity.Organization, error) {
	return r.findByIDFn(ctx, db, id)
}
func (r *fakeOrgRepo) Create(ctx context.Context, db identity.DBExecutor, org identity.Organization) error { return nil }
func (r *fakeOrgRepo) FindBySlug(ctx context.Context, db identity.DBExecutor, slug string) (identity.Organization, error) {
	return identity.Organization{}, nil
}
func (r *fakeOrgRepo) ExistsAny(ctx context.Context, db identity.DBExecutor) (bool, error) { return false, nil }

type fakeHasher struct {
	verifyFn func(password []byte, encodedHash identity.PasswordHash) error
}

func (h *fakeHasher) Hash(password []byte) (identity.PasswordHash, error) { return "hash", nil }
func (h *fakeHasher) Verify(password []byte, encodedHash identity.PasswordHash) error {
	return h.verifyFn(password, encodedHash)
}

type fakeNormalizer struct {
	normalizeFn func(email string) (string, error)
}

func (n *fakeNormalizer) Normalize(email string) (string, error) {
	return n.normalizeFn(email)
}

type fakeSessionIssuer struct {
	issueFn func(sub string) (string, error)
}

func (s *fakeSessionIssuer) Issue(sub string) (string, error) {
	return s.issueFn(sub)
}

type fakeDB struct{}

func (d *fakeDB) Exec(ctx context.Context, sql string, args ...any) (int64, error)    { return 0, nil }
func (d *fakeDB) Query(ctx context.Context, sql string, args ...any) (identity.Rows, error) {
	return nil, nil
}
func (d *fakeDB) QueryRow(ctx context.Context, sql string, args ...any) identity.RowScanner { return nil }

func buildService(
	user *fakeUserRepo,
	cred *fakeCredentialRepo,
	grant *fakeGrantRepo,
	mem *fakeMemRepo,
	org *fakeOrgRepo,
	hasher application.PasswordHasher,
	normalizer application.EmailNormalizer,
	issuer authentication.SessionIssuer,
) *authentication.AuthenticateUserService {
	if user == nil {
		user = &fakeUserRepo{}
	}
	if cred == nil {
		cred = &fakeCredentialRepo{}
	}
	if grant == nil {
		grant = &fakeGrantRepo{}
	}
	if mem == nil {
		mem = &fakeMemRepo{}
	}
	if org == nil {
		org = &fakeOrgRepo{}
	}
	if hasher == nil {
		hasher = &fakeHasher{verifyFn: func(p []byte, h identity.PasswordHash) error { return nil }}
	}
	if normalizer == nil {
		normalizer = &fakeNormalizer{normalizeFn: func(e string) (string, error) { return e, nil }}
	}
	if issuer == nil {
		issuer = &fakeSessionIssuer{issueFn: func(sub string) (string, error) { return "token", nil }}
	}
	resolver := authentication.NewPrincipalResolverService(user, grant, mem, org)
	return authentication.NewAuthenticateUserService(user, cred, hasher, normalizer, issuer, resolver)
}

func buildCmd() authentication.AuthenticateUserCommand {
	return authentication.AuthenticateUserCommand{Email: "owner@test.com", Password: "pass"}
}

func TestAuthenticate_Success(t *testing.T) {
	var createdUser identity.User
	userRepo := &fakeUserRepo{
		findByEmailFn: func(ctx context.Context, db identity.DBExecutor, email string) (identity.User, error) {
			user, _ := identity.NewUser(email, "Owner")
			createdUser = user
			return user, nil
		},
		findByIDFn: func(ctx context.Context, db identity.DBExecutor, id identity.UserID) (identity.User, error) {
			return createdUser, nil
		},
	}
	credRepo := &fakeCredentialRepo{findByUserIDFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.Credential, error) {
		cred, _ := identity.NewCredential(userID, "hash")
		return cred, nil
	}}
	grantRepo := &fakeGrantRepo{findByUserFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.PlatformGrant, error) {
		g, _ := identity.NewPlatformGrant(userID, identity.PlatformRoleSuperAdmin)
		return g, nil
	}}
	memRepo := &fakeMemRepo{findByUserFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) ([]identity.Membership, error) {
		orgID := identity.NewOrganizationID()
		m, _ := identity.NewMembership(userID, orgID, identity.OrganizationRoleOwner)
		return []identity.Membership{m}, nil
	}}
	orgRepo := &fakeOrgRepo{findByIDFn: func(ctx context.Context, db identity.DBExecutor, id identity.OrganizationID) (identity.Organization, error) {
		return identity.Organization{ID: id, Name: "Org", Slug: "org"}, nil
	}}

	svc := buildService(userRepo, credRepo, grantRepo, memRepo, orgRepo, nil, nil, nil)

	result, err := svc.Execute(context.Background(), &fakeDB{}, buildCmd())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SessionToken == "" {
		t.Error("expected session token")
	}
	if result.Email != "owner@test.com" {
		t.Errorf("email mismatch: %q", result.Email)
	}
	if result.DefaultContext.Type != "organization" {
		t.Errorf("expected default context type 'organization', got %q", result.DefaultContext.Type)
	}
}

func TestAuthenticate_InvalidCredentials(t *testing.T) {
	userRepo := &fakeUserRepo{findByEmailFn: func(ctx context.Context, db identity.DBExecutor, email string) (identity.User, error) {
		return identity.User{}, identity.ErrUserNotFound
	}}

	svc := buildService(userRepo, nil, nil, nil, nil, nil, nil, nil)

	_, err := svc.Execute(context.Background(), &fakeDB{}, buildCmd())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAuthenticate_InvalidEmail(t *testing.T) {
	normalizer := &fakeNormalizer{normalizeFn: func(email string) (string, error) {
		return "", identity.ErrInvalidEmail
	}}
	svc := buildService(nil, nil, nil, nil, nil, nil, normalizer, nil)

	_, err := svc.Execute(context.Background(), &fakeDB{}, buildCmd())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAuthenticate_EmailNormalized(t *testing.T) {
	var receivedEmail, returnedEmail string
	normalizer := &fakeNormalizer{normalizeFn: func(email string) (string, error) {
		receivedEmail = email
		returnedEmail = "normalized@test.com"
		return returnedEmail, nil
	}}
	userRepo := &fakeUserRepo{findByEmailFn: func(ctx context.Context, db identity.DBExecutor, email string) (identity.User, error) {
		if email != "normalized@test.com" {
			t.Errorf("expected normalized email, got %q", email)
		}
		user, _ := identity.NewUser(email, "Owner")
		return user, nil
	}}
	credRepo := &fakeCredentialRepo{findByUserIDFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.Credential, error) {
		cred, _ := identity.NewCredential(userID, "hash")
		return cred, nil
	}}
	grantRepo := &fakeGrantRepo{findByUserFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.PlatformGrant, error) {
		g, _ := identity.NewPlatformGrant(userID, identity.PlatformRoleSuperAdmin)
		return g, nil
	}}
	memRepo := &fakeMemRepo{findByUserFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) ([]identity.Membership, error) {
		return nil, nil
	}}

	svc := buildService(userRepo, credRepo, grantRepo, memRepo, nil, nil, normalizer, nil)

	cmd := authentication.AuthenticateUserCommand{Email: "  Raw@Input.com  ", Password: "pass"}
	_, err := svc.Execute(context.Background(), &fakeDB{}, cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedEmail != "  Raw@Input.com  " {
		t.Errorf("normalizer should receive raw email, got %q", receivedEmail)
	}
}

func TestAuthenticate_SameErrorForUnknownEmailAndWrongPassword(t *testing.T) {
	hasher := &fakeHasher{verifyFn: func(p []byte, h identity.PasswordHash) error {
		return identity.ErrInvalidCredentials
	}}
	userRepo := &fakeUserRepo{findByEmailFn: func(ctx context.Context, db identity.DBExecutor, email string) (identity.User, error) {
		user, _ := identity.NewUser(email, "Owner")
		return user, nil
	}}
	credRepo := &fakeCredentialRepo{findByUserIDFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.Credential, error) {
		cred, _ := identity.NewCredential(userID, "hash")
		return cred, nil
	}}

	svc := buildService(userRepo, credRepo, nil, nil, nil, hasher, nil, nil)

	_, err := svc.Execute(context.Background(), &fakeDB{}, buildCmd())
	if err == nil {
		t.Fatal("expected error for wrong password")
	}

	userRepoNone := &fakeUserRepo{findByEmailFn: func(ctx context.Context, db identity.DBExecutor, email string) (identity.User, error) {
		return identity.User{}, identity.ErrUserNotFound
	}}

	svc2 := buildService(userRepoNone, nil, nil, nil, nil, nil, nil, nil)

	_, err2 := svc2.Execute(context.Background(), &fakeDB{}, buildCmd())
	if err2 == nil {
		t.Fatal("expected error for unknown email")
	}
}

func TestAuthenticate_OperationalErrorFindingUser(t *testing.T) {
	userRepo := &fakeUserRepo{findByEmailFn: func(ctx context.Context, db identity.DBExecutor, email string) (identity.User, error) {
		return identity.User{}, errors.New("db connection refused")
	}}
	svc := buildService(userRepo, nil, nil, nil, nil, nil, nil, nil)

	_, err := svc.Execute(context.Background(), &fakeDB{}, buildCmd())
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, identity.ErrInvalidCredentials) {
		t.Error("operational error should not be masked as invalid credentials")
	}
}

func TestAuthenticate_OperationalErrorFindingCredential(t *testing.T) {
	userRepo := &fakeUserRepo{findByEmailFn: func(ctx context.Context, db identity.DBExecutor, email string) (identity.User, error) {
		user, _ := identity.NewUser(email, "Owner")
		return user, nil
	}}
	credRepo := &fakeCredentialRepo{findByUserIDFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.Credential, error) {
		return identity.Credential{}, errors.New("db connection refused")
	}}
	svc := buildService(userRepo, credRepo, nil, nil, nil, nil, nil, nil)

	_, err := svc.Execute(context.Background(), &fakeDB{}, buildCmd())
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, identity.ErrInvalidCredentials) {
		t.Error("operational error should not be masked as invalid credentials")
	}
}

func TestAuthenticate_GrantNotFoundContinues(t *testing.T) {
	userRepo := &fakeUserRepo{findByEmailFn: func(ctx context.Context, db identity.DBExecutor, email string) (identity.User, error) {
		user, _ := identity.NewUser(email, "Owner")
		return user, nil
	}}
	credRepo := &fakeCredentialRepo{findByUserIDFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.Credential, error) {
		cred, _ := identity.NewCredential(userID, "hash")
		return cred, nil
	}}
	grantRepo := &fakeGrantRepo{findByUserFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.PlatformGrant, error) {
		return identity.PlatformGrant{}, identity.ErrMembershipNotFound
	}}
	memRepo := &fakeMemRepo{findByUserFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) ([]identity.Membership, error) {
		orgID := identity.NewOrganizationID()
		m, _ := identity.NewMembership(userID, orgID, identity.OrganizationRoleOwner)
		return []identity.Membership{m}, nil
	}}
	orgRepo := &fakeOrgRepo{findByIDFn: func(ctx context.Context, db identity.DBExecutor, id identity.OrganizationID) (identity.Organization, error) {
		return identity.Organization{ID: id, Name: "Org", Slug: "org"}, nil
	}}

	svc := buildService(userRepo, credRepo, grantRepo, memRepo, orgRepo, nil, nil, nil)

	result, err := svc.Execute(context.Background(), &fakeDB{}, buildCmd())
	if err != nil {
		t.Fatalf("expected success when grant is simply absent: %v", err)
	}
	if len(result.PlatformRoles) != 0 {
		t.Errorf("expected 0 platform roles, got %d", len(result.PlatformRoles))
	}
}

func TestAuthenticate_OperationalErrorFindingGrant(t *testing.T) {
	userRepo := &fakeUserRepo{findByEmailFn: func(ctx context.Context, db identity.DBExecutor, email string) (identity.User, error) {
		user, _ := identity.NewUser(email, "Owner")
		return user, nil
	}}
	credRepo := &fakeCredentialRepo{findByUserIDFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.Credential, error) {
		cred, _ := identity.NewCredential(userID, "hash")
		return cred, nil
	}}
	grantRepo := &fakeGrantRepo{findByUserFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.PlatformGrant, error) {
		return identity.PlatformGrant{}, errors.New("db connection refused")
	}}
	svc := buildService(userRepo, credRepo, grantRepo, nil, nil, nil, nil, nil)

	_, err := svc.Execute(context.Background(), &fakeDB{}, buildCmd())
	if err == nil {
		t.Fatal("expected error for operational grant failure")
	}
	if errors.Is(err, identity.ErrInvalidCredentials) {
		t.Error("operational error should not be masked as invalid credentials")
	}
}

func TestAuthenticate_OperationalErrorFindingMemberships(t *testing.T) {
	userRepo := &fakeUserRepo{findByEmailFn: func(ctx context.Context, db identity.DBExecutor, email string) (identity.User, error) {
		user, _ := identity.NewUser(email, "Owner")
		return user, nil
	}}
	credRepo := &fakeCredentialRepo{findByUserIDFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.Credential, error) {
		cred, _ := identity.NewCredential(userID, "hash")
		return cred, nil
	}}
	grantRepo := &fakeGrantRepo{findByUserFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.PlatformGrant, error) {
		return identity.PlatformGrant{}, identity.ErrMembershipNotFound
	}}
	memRepo := &fakeMemRepo{findByUserFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) ([]identity.Membership, error) {
		return nil, errors.New("db connection refused")
	}}
	svc := buildService(userRepo, credRepo, grantRepo, memRepo, nil, nil, nil, nil)

	_, err := svc.Execute(context.Background(), &fakeDB{}, buildCmd())
	if err == nil {
		t.Fatal("expected error for operational membership failure")
	}
}

func TestAuthenticate_OrganizationNotFoundFails(t *testing.T) {
	userRepo := &fakeUserRepo{findByEmailFn: func(ctx context.Context, db identity.DBExecutor, email string) (identity.User, error) {
		user, _ := identity.NewUser(email, "Owner")
		return user, nil
	}}
	credRepo := &fakeCredentialRepo{findByUserIDFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.Credential, error) {
		cred, _ := identity.NewCredential(userID, "hash")
		return cred, nil
	}}
	grantRepo := &fakeGrantRepo{findByUserFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.PlatformGrant, error) {
		return identity.PlatformGrant{}, identity.ErrMembershipNotFound
	}}
	memRepo := &fakeMemRepo{findByUserFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) ([]identity.Membership, error) {
		orgID := identity.NewOrganizationID()
		m, _ := identity.NewMembership(userID, orgID, identity.OrganizationRoleOwner)
		return []identity.Membership{m}, nil
	}}
	orgRepo := &fakeOrgRepo{findByIDFn: func(ctx context.Context, db identity.DBExecutor, id identity.OrganizationID) (identity.Organization, error) {
		return identity.Organization{}, identity.ErrOrganizationNotFound
	}}

	svc := buildService(userRepo, credRepo, grantRepo, memRepo, orgRepo, nil, nil, nil)

	_, err := svc.Execute(context.Background(), &fakeDB{}, buildCmd())
	if err == nil {
		t.Fatal("expected error when org referenced by membership does not exist")
	}
}

func TestAuthenticate_OperationalErrorFindingOrganization(t *testing.T) {
	userRepo := &fakeUserRepo{findByEmailFn: func(ctx context.Context, db identity.DBExecutor, email string) (identity.User, error) {
		user, _ := identity.NewUser(email, "Owner")
		return user, nil
	}}
	credRepo := &fakeCredentialRepo{findByUserIDFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.Credential, error) {
		cred, _ := identity.NewCredential(userID, "hash")
		return cred, nil
	}}
	grantRepo := &fakeGrantRepo{findByUserFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.PlatformGrant, error) {
		return identity.PlatformGrant{}, identity.ErrMembershipNotFound
	}}
	memRepo := &fakeMemRepo{findByUserFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) ([]identity.Membership, error) {
		orgID := identity.NewOrganizationID()
		m, _ := identity.NewMembership(userID, orgID, identity.OrganizationRoleOwner)
		return []identity.Membership{m}, nil
	}}
	orgRepo := &fakeOrgRepo{findByIDFn: func(ctx context.Context, db identity.DBExecutor, id identity.OrganizationID) (identity.Organization, error) {
		return identity.Organization{}, errors.New("db connection refused")
	}}

	svc := buildService(userRepo, credRepo, grantRepo, memRepo, orgRepo, nil, nil, nil)

	_, err := svc.Execute(context.Background(), &fakeDB{}, buildCmd())
	if err == nil {
		t.Fatal("expected error for operational org failure")
	}
}

func TestAuthenticate_NoGrantAndNoMembership(t *testing.T) {
	userRepo := &fakeUserRepo{findByEmailFn: func(ctx context.Context, db identity.DBExecutor, email string) (identity.User, error) {
		user, _ := identity.NewUser(email, "Owner")
		return user, nil
	}}
	credRepo := &fakeCredentialRepo{findByUserIDFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.Credential, error) {
		cred, _ := identity.NewCredential(userID, "hash")
		return cred, nil
	}}
	grantRepo := &fakeGrantRepo{findByUserFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.PlatformGrant, error) {
		return identity.PlatformGrant{}, identity.ErrMembershipNotFound
	}}
	memRepo := &fakeMemRepo{findByUserFn: func(ctx context.Context, db identity.DBExecutor, userID identity.UserID) ([]identity.Membership, error) {
		return nil, nil
	}}

	svc := buildService(userRepo, credRepo, grantRepo, memRepo, nil, nil, nil, nil)

	_, err := svc.Execute(context.Background(), &fakeDB{}, buildCmd())
	if err == nil {
		t.Fatal("expected error for user with no grant and no membership")
	}
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}
