package pgx_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/pgx"
	"donarium/server/internal/platform/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://donarium:donarium@localhost:5432/donarium?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		os.Exit(0)
	}
	if err := testPool.Ping(ctx); err != nil {
		testPool.Close()
		os.Exit(0)
	}
	initializeSchema(ctx)
	code := m.Run()
	testPool.Close()
	os.Exit(code)
}

func initializeSchema(ctx context.Context) {
	tables := []string{"platform_grants", "memberships", "credentials", "organizations", "users", "schema_migrations"}
	for _, t := range tables {
		_, _ = testPool.Exec(ctx, "DROP TABLE IF EXISTS "+t+" CASCADE")
	}
	if err := database.RunMigrations(ctx, testPool); err != nil {
		panic("migrations failed: " + err.Error())
	}
}

func db() identity.DBExecutor {
	return pgx.NewExecutorFromPool(testPool)
}

func cleanAll(t *testing.T) {
	t.Helper()
	tables := []string{"platform_grants", "memberships", "credentials", "organizations", "users"}
	for _, tbl := range tables {
		_, err := testPool.Exec(context.Background(), "DELETE FROM "+tbl)
		if err != nil {
			t.Fatalf("truncate %s: %v", tbl, err)
		}
	}
}

func mustCreateUser(t *testing.T, user identity.User) {
	t.Helper()
	if err := pgx.NewUserRepo().Create(context.Background(), db(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}
}

func mustCreateOrg(t *testing.T, org identity.Organization) {
	t.Helper()
	if err := pgx.NewOrganizationRepo().Create(context.Background(), db(), org); err != nil {
		t.Fatalf("create org: %v", err)
	}
}

func TestUserRepo_CreateAndFindByID(t *testing.T) {
	cleanAll(t)
	user, _ := identity.NewUser("a@b.com", "A")
	mustCreateUser(t, user)
	found, err := pgx.NewUserRepo().FindByID(context.Background(), db(), user.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Email != user.Email {
		t.Errorf("email mismatch: %q vs %q", user.Email, found.Email)
	}
}

func TestUserRepo_FindByEmail(t *testing.T) {
	cleanAll(t)
	user, _ := identity.NewUser("find@b.com", "X")
	mustCreateUser(t, user)
	found, err := pgx.NewUserRepo().FindByEmail(context.Background(), db(), "find@b.com")
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if found.ID != user.ID {
		t.Error("id mismatch")
	}
}

func TestUserRepo_DuplicateEmail(t *testing.T) {
	cleanAll(t)
	u1, _ := identity.NewUser("dup@c.com", "U1")
	u2, _ := identity.NewUser("dup@c.com", "U2")
	mustCreateUser(t, u1)
	err := pgx.NewUserRepo().Create(context.Background(), db(), u2)
	if !errors.Is(err, identity.ErrDuplicateEmail) {
		t.Errorf("expected ErrDuplicateEmail, got %v", err)
	}
}

func TestUserRepo_ExistsByEmail(t *testing.T) {
	cleanAll(t)
	user, _ := identity.NewUser("exists@d.com", "E")
	mustCreateUser(t, user)
	ok, _ := pgx.NewUserRepo().ExistsByEmail(context.Background(), db(), "exists@d.com")
	if !ok {
		t.Error("expected true")
	}
	ok, _ = pgx.NewUserRepo().ExistsByEmail(context.Background(), db(), "no@d.com")
	if ok {
		t.Error("expected false")
	}
}

func TestUserRepo_NotFound(t *testing.T) {
	cleanAll(t)
	_, err := pgx.NewUserRepo().FindByID(context.Background(), db(), identity.NewUserID())
	if !errors.Is(err, identity.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestCredentialRepo_CreateAndFind(t *testing.T) {
	cleanAll(t)
	user, _ := identity.NewUser("cred@e.com", "C")
	mustCreateUser(t, user)
	cred, _ := identity.NewCredential(user.ID, "hashed")
	_ = pgx.NewCredentialRepo().Create(context.Background(), db(), cred)
	found, err := pgx.NewCredentialRepo().FindByUserID(context.Background(), db(), user.ID)
	if err != nil {
		t.Fatalf("FindByUserID: %v", err)
	}
	if found.PasswordHash != "hashed" {
		t.Error("hash mismatch")
	}
}

func TestCredentialRepo_NotFound(t *testing.T) {
	cleanAll(t)
	_, err := pgx.NewCredentialRepo().FindByUserID(context.Background(), db(), identity.NewUserID())
	if !errors.Is(err, identity.ErrCredentialNotFound) {
		t.Errorf("expected ErrCredentialNotFound, got %v", err)
	}
}

func TestOrgRepo_CreateAndFind(t *testing.T) {
	cleanAll(t)
	user, _ := identity.NewUser("org@f.com", "O")
	mustCreateUser(t, user)
	org, _ := identity.NewOrganization("My Org", "my-org", user.ID)
	if err := pgx.NewOrganizationRepo().Create(context.Background(), db(), org); err != nil {
		t.Fatalf("Create: %v", err)
	}
	found, err := pgx.NewOrganizationRepo().FindByID(context.Background(), db(), org.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Name != "My Org" || found.Slug != "my-org" {
		t.Error("mismatch")
	}
}

func TestOrgRepo_DuplicateSlug(t *testing.T) {
	cleanAll(t)
	u1, _ := identity.NewUser("s1@g.com", "S1")
	u2, _ := identity.NewUser("s2@g.com", "S2")
	mustCreateUser(t, u1)
	mustCreateUser(t, u2)
	o1, _ := identity.NewOrganization("A", "same", u1.ID)
	o2, _ := identity.NewOrganization("B", "same", u2.ID)
	_ = pgx.NewOrganizationRepo().Create(context.Background(), db(), o1)
	err := pgx.NewOrganizationRepo().Create(context.Background(), db(), o2)
	if !errors.Is(err, identity.ErrDuplicateSlug) {
		t.Errorf("expected ErrDuplicateSlug, got %v", err)
	}
}

func TestOrgRepo_ExistsAny(t *testing.T) {
	cleanAll(t)
	ok, _ := pgx.NewOrganizationRepo().ExistsAny(context.Background(), db())
	if ok {
		t.Error("expected false on empty DB")
	}
	user, _ := identity.NewUser("any@h.com", "A")
	mustCreateUser(t, user)
	org, _ := identity.NewOrganization("X", "x-org", user.ID)
	_ = pgx.NewOrganizationRepo().Create(context.Background(), db(), org)
	ok, _ = pgx.NewOrganizationRepo().ExistsAny(context.Background(), db())
	if !ok {
		t.Error("expected true after insert")
	}
}

func TestOrgRepo_FindBySlug(t *testing.T) {
	cleanAll(t)
	user, _ := identity.NewUser("slug@i.com", "S")
	mustCreateUser(t, user)
	org, _ := identity.NewOrganization("Slug", "sluggy", user.ID)
	_ = pgx.NewOrganizationRepo().Create(context.Background(), db(), org)
	found, err := pgx.NewOrganizationRepo().FindBySlug(context.Background(), db(), "sluggy")
	if err != nil {
		t.Fatalf("FindBySlug: %v", err)
	}
	if found.Name != "Slug" {
		t.Error("name mismatch")
	}
}

func TestOrgRepo_DBRejectsInvalidSlug(t *testing.T) {
	cleanAll(t)
	user, _ := identity.NewUser("bads@j.com", "B")
	mustCreateUser(t, user)
	_, err := db().Exec(context.Background(),
		`INSERT INTO organizations (id, name, slug, created_by, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		uuid.UUID(identity.NewOrganizationID()), "Bad", "UPPER",
		uuid.UUID(identity.NewUserID()), time.Now(),
	)
	if err == nil {
		t.Fatal("expected DB constraint violation")
	}
}

func TestMembershipRepo_CreateAndFind(t *testing.T) {
	cleanAll(t)
	user, _ := identity.NewUser("mem@k.com", "M")
	mustCreateUser(t, user)
	org, _ := identity.NewOrganization("MOrg", "m-org", user.ID)
	mustCreateOrg(t, org)
	m, _ := identity.NewMembership(user.ID, org.ID, identity.OrganizationRoleOwner)
	_ = pgx.NewMembershipRepo().Create(context.Background(), db(), m)
	found, err := pgx.NewMembershipRepo().FindByUserAndOrg(context.Background(), db(), user.ID, org.ID)
	if err != nil {
		t.Fatalf("FindByUserAndOrg: %v", err)
	}
	if found.Role != identity.OrganizationRoleOwner {
		t.Error("role mismatch")
	}
}

func TestMembershipRepo_Duplicate(t *testing.T) {
	cleanAll(t)
	user, _ := identity.NewUser("dmem@l.com", "D")
	mustCreateUser(t, user)
	org, _ := identity.NewOrganization("DOrg", "d-org", user.ID)
	mustCreateOrg(t, org)
	m1, _ := identity.NewMembership(user.ID, org.ID, identity.OrganizationRoleOwner)
	m2, _ := identity.NewMembership(user.ID, org.ID, identity.OrganizationRoleOwner)
	_ = pgx.NewMembershipRepo().Create(context.Background(), db(), m1)
	err := pgx.NewMembershipRepo().Create(context.Background(), db(), m2)
	if err == nil {
		t.Fatal("expected duplicate key error")
	}
}

func TestPlatformGrantRepo_CreateAndFind(t *testing.T) {
	cleanAll(t)
	user, _ := identity.NewUser("admin@m.com", "Admin")
	mustCreateUser(t, user)
	grant, _ := identity.NewPlatformGrant(user.ID, identity.PlatformRoleSuperAdmin)
	_ = pgx.NewPlatformGrantRepo().Create(context.Background(), db(), grant)
	found, err := pgx.NewPlatformGrantRepo().FindByUser(context.Background(), db(), user.ID)
	if err != nil {
		t.Fatalf("FindByUser: %v", err)
	}
	if found.Role != identity.PlatformRoleSuperAdmin {
		t.Error("role mismatch")
	}
}

func TestMembershipRepo_FindByUser_TieBreakerOrdering(t *testing.T) {
	cleanAll(t)
	user, _ := identity.NewUser("tie@example.com", "Tie")
	mustCreateUser(t, user)

	orgA, _ := identity.NewOrganization("Org A", "org-a", user.ID)
	orgB, _ := identity.NewOrganization("Org B", "org-b", user.ID)
	mustCreateOrg(t, orgA)
	mustCreateOrg(t, orgB)

	now := time.Now().UTC()

	m1 := identity.Membership{UserID: user.ID, OrganizationID: orgA.ID, Role: identity.OrganizationRoleOwner, CreatedAt: now}
	m2 := identity.Membership{UserID: user.ID, OrganizationID: orgB.ID, Role: identity.OrganizationRoleOwner, CreatedAt: now}

	if err := pgx.NewMembershipRepo().Create(context.Background(), db(), m1); err != nil {
		t.Fatalf("create membership 1: %v", err)
	}
	if err := pgx.NewMembershipRepo().Create(context.Background(), db(), m2); err != nil {
		t.Fatalf("create membership 2: %v", err)
	}

	smallerID, largerID := orgA.ID, orgB.ID
	if uuid.UUID(orgB.ID).String() < uuid.UUID(orgA.ID).String() {
		smallerID, largerID = orgB.ID, orgA.ID
	}

	results, err := pgx.NewMembershipRepo().FindByUser(context.Background(), db(), user.ID)
	if err != nil {
		t.Fatalf("FindByUser: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 memberships, got %d", len(results))
	}

	if results[0].OrganizationID != smallerID {
		t.Errorf("expected smaller UUID (%v) first, got %v", uuid.UUID(smallerID), uuid.UUID(results[0].OrganizationID))
	}

	if results[1].OrganizationID != largerID {
		t.Errorf("expected larger UUID (%v) second, got %v", uuid.UUID(largerID), uuid.UUID(results[1].OrganizationID))
	}

	for i := 0; i < 3; i++ {
		rerun, err := pgx.NewMembershipRepo().FindByUser(context.Background(), db(), user.ID)
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		if rerun[0].OrganizationID != smallerID {
			t.Errorf("iteration %d: ordering changed — expected %v first, got %v",
				i, uuid.UUID(smallerID), uuid.UUID(rerun[0].OrganizationID))
		}
	}
}
