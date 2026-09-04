package application_test

import (
	"context"
	"os"
	"testing"
	"time"

	"donarium/server/internal/identity"
	identitypgx "donarium/server/internal/identity/pgx"
	"donarium/server/internal/platform/database"
	"donarium/server/internal/properties"
	"donarium/server/internal/properties/application"
	"donarium/server/internal/properties/pgx"

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
		panic("TEST_DATABASE_URL: failed to create pool: " + err.Error())
	}
	if err := testPool.Ping(ctx); err != nil {
		panic("TEST_DATABASE_URL: failed to ping database (is PostgreSQL running? set TEST_DATABASE_URL): " + err.Error())
	}
	tables := []string{"property_stakeholders", "properties", "platform_grants", "memberships", "credentials", "organizations", "users", "schema_migrations"}
	for _, t := range tables {
		_, _ = testPool.Exec(ctx, "DROP TABLE IF EXISTS "+t+" CASCADE")
	}
	if err := database.RunMigrations(ctx, testPool); err != nil {
		panic("migrations failed: " + err.Error())
	}
	code := m.Run()
	testPool.Close()
	os.Exit(code)
}

func cleanAll(t *testing.T) {
	t.Helper()
	for _, tbl := range []string{"property_stakeholders", "properties", "platform_grants", "memberships", "credentials", "organizations", "users"} {
		if _, err := testPool.Exec(context.Background(), "DELETE FROM "+tbl); err != nil {
			t.Fatalf("delete %s: %v", tbl, err)
		}
	}
}

func mustCreateUser(t *testing.T) string {
	t.Helper()
	user, _ := identity.NewUser("svc-"+uuid.NewString()+"@example.com", "Svc")
	if err := identitypgx.NewUserRepo().Create(context.Background(), identitypgx.NewExecutorFromPool(testPool), user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return uuid.UUID(user.ID).String()
}

func newService() *application.Service {
	repo := pgx.NewRepository()
	tx := pgx.NewTransactionManager(testPool)
	return application.NewService(repo, tx)
}

func newServiceWithStakeholders() *application.Service {
	repo := pgx.NewRepository()
	stakeRepo := pgx.NewStakeholderRepository()
	tx := pgx.NewTransactionManager(testPool)
	return application.NewServiceWithStakeholders(repo, stakeRepo, tx)
}

func validCmd() application.RegisterPropertyCommand {
	return application.RegisterPropertyCommand{
		DisplayName:    "Casa Sol",
		Classification: "house",
		Address: properties.Address{
			Street:     "123 Calle",
			City:       "Madrid",
			State:      "Madrid",
			PostalCode: "28001",
			Country:    "ES",
		},
		RentalCadence: "monthly",
		StandardRent:  120000,
	}
}

func TestService_RegisterProperty_HappyPath(t *testing.T) {
	cleanAll(t)
	uid := mustCreateUser(t)
	svc := newService()

	res, err := svc.RegisterProperty(context.Background(), validCmd(), uid)
	if err != nil {
		t.Fatalf("RegisterProperty: %v", err)
	}
	if res.Property.DisplayName != "Casa Sol" {
		t.Errorf("displayName mismatch: %q", res.Property.DisplayName)
	}
	if res.Property.Classification != properties.ClassificationHouse {
		t.Errorf("classification mismatch: %q", res.Property.Classification)
	}
	// verify persisted via repo
	found, err := pgx.NewRepository().FindByID(context.Background(), pgx.NewExecutorFromPool(testPool), res.Property.ID)
	if err != nil {
		t.Fatalf("FindByID after create: %v", err)
	}
	if found.CreatedBy.String() != uid {
		t.Errorf("createdBy mismatch: %q vs %q", found.CreatedBy.String(), uid)
	}
}

func TestService_RegisterProperty_InvalidClassification(t *testing.T) {
	cleanAll(t)
	uid := mustCreateUser(t)
	svc := newService()
	cmd := validCmd()
	cmd.Classification = "castle"
	_, err := svc.RegisterProperty(context.Background(), cmd, uid)
	if err != properties.ErrInvalidClassification {
		t.Errorf("expected ErrInvalidClassification, got %v", err)
	}
}

func TestService_RegisterProperty_InvalidDisplayName(t *testing.T) {
	cleanAll(t)
	uid := mustCreateUser(t)
	svc := newService()
	cmd := validCmd()
	cmd.DisplayName = "A"
	_, err := svc.RegisterProperty(context.Background(), cmd, uid)
	if err != properties.ErrInvalidDisplayName {
		t.Errorf("expected ErrInvalidDisplayName, got %v", err)
	}
}

func TestService_RegisterProperty_InvalidAddress(t *testing.T) {
	cleanAll(t)
	uid := mustCreateUser(t)
	svc := newService()
	cmd := validCmd()
	cmd.Address.Street = ""
	_, err := svc.RegisterProperty(context.Background(), cmd, uid)
	if err != properties.ErrInvalidAddress {
		t.Errorf("expected ErrInvalidAddress, got %v", err)
	}
}

func TestService_RegisterProperty_InvalidCadence(t *testing.T) {
	cleanAll(t)
	uid := mustCreateUser(t)
	svc := newService()
	cmd := validCmd()
	cmd.RentalCadence = "biweekly"
	_, err := svc.RegisterProperty(context.Background(), cmd, uid)
	if err != properties.ErrInvalidRentalCadence {
		t.Errorf("expected ErrInvalidRentalCadence, got %v", err)
	}
}

func TestService_RegisterProperty_InvalidRent(t *testing.T) {
	cleanAll(t)
	uid := mustCreateUser(t)
	svc := newService()
	cmd := validCmd()
	cmd.StandardRent = 0
	_, err := svc.RegisterProperty(context.Background(), cmd, uid)
	if err != properties.ErrInvalidStandardRent {
		t.Errorf("expected ErrInvalidStandardRent, got %v", err)
	}
}

func TestService_ListAndGet_Authorization(t *testing.T) {
	cleanAll(t)
	uid1 := mustCreateUser(t)
	uid2 := mustCreateUser(t)
	svc := newService()

	res, err := svc.RegisterProperty(context.Background(), validCmd(), uid1)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// uid1 can list and get
	db := pgx.NewExecutorFromPool(testPool)
	list, err := svc.ListAccessible(context.Background(), db, uid1)
	if err != nil {
		t.Fatalf("ListAccessible uid1: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 for uid1, got %d", len(list))
	}
	prop, err := svc.GetByID(context.Background(), db, res.Property.ID.String(), uid1)
	if err != nil {
		t.Fatalf("GetByID authorized: %v", err)
	}
	if prop.ID != res.Property.ID {
		t.Error("id mismatch")
	}

	// uid2 cannot see uid1's property
	list2, err := svc.ListAccessible(context.Background(), db, uid2)
	if err != nil {
		t.Fatalf("ListAccessible uid2: %v", err)
	}
	if len(list2) != 0 {
		t.Errorf("expected 0 for uid2, got %d", len(list2))
	}
	_, err = svc.GetByID(context.Background(), db, res.Property.ID.String(), uid2)
	if err != properties.ErrPropertyNotFound {
		t.Errorf("expected ErrPropertyNotFound for unauthorized, got %v", err)
	}

	// invalid id format also maps to not found (do not disclose)
	_, err = svc.GetByID(context.Background(), db, "not-a-uuid", uid1)
	if err != properties.ErrPropertyNotFound {
		t.Errorf("expected ErrPropertyNotFound for invalid uuid, got %v", err)
	}
}

func TestService_ClassificationAndCadence_CaseInsensitive(t *testing.T) {
	cleanAll(t)
	uid := mustCreateUser(t)
	svc := newService()
	cmd := validCmd()
	cmd.Classification = "House"
	cmd.RentalCadence = "MONTHLY"
	res, err := svc.RegisterProperty(context.Background(), cmd, uid)
	if err != nil {
		t.Fatalf("case insensitive: %v", err)
	}
	if res.Property.Classification != properties.ClassificationHouse {
		t.Errorf("expected house, got %q", res.Property.Classification)
	}
	if res.Property.RentalCadence != properties.CadenceMonthly {
		t.Errorf("expected monthly, got %q", res.Property.RentalCadence)
	}

	// multi-unit variants
	for _, v := range []string{"Multi-unit", "multi_unit", "multiunit", "MULTI-UNIT"} {
		cmd2 := validCmd()
		cmd2.DisplayName = "M " + v
		cmd2.Classification = v
		if _, err := svc.RegisterProperty(context.Background(), cmd2, uid); err != nil {
			t.Errorf("multi_unit variant %q failed: %v", v, err)
		}
	}
}

func TestService_FourStitchOptions(t *testing.T) {
	cleanAll(t)
	uid := mustCreateUser(t)
	svc := newServiceWithStakeholders()
	db := pgx.NewExecutorFromPool(testPool)

	// Helper to fetch stakeholders for property
	getStakeholders := func(pid string) []string {
		// query via stakeholder repo
		pidParsed, _ := uuid.Parse(pid)
		stakes, err := pgx.NewStakeholderRepository().FindByProperty(context.Background(), db, properties.PropertyID(pidParsed))
		if err != nil {
			t.Fatalf("FindByProperty: %v", err)
		}
		var out []string
		for _, s := range stakes {
			out = append(out, string(s.Party.Type)+":"+string(s.Role))
		}
		return out
	}

	// Option 1: I am both Owner & Manager (self both)
	cmd1 := validCmd()
	cmd1.DisplayName = "Opt1 Both"
	cmd1.Stakeholders = []application.StakeholderInput{
		{Party: application.PartyInput{Type: "user", UserID: uid}, Role: "owner"},
		{Party: application.PartyInput{Type: "user", UserID: uid}, Role: "manager"},
	}
	res1, err := svc.RegisterProperty(context.Background(), cmd1, uid)
	if err != nil {
		t.Fatalf("option1: %v", err)
	}
	st1 := getStakeholders(res1.Property.ID.String())
	if len(st1) != 2 {
		t.Errorf("option1 expected 2 stakeholders, got %d: %v", len(st1), st1)
	}

	// Option 2: I am Manager on behalf of Owner (external owner, self manager)
	cmd2 := validCmd()
	cmd2.DisplayName = "Opt2 Manager for Owner"
	cmd2.Stakeholders = []application.StakeholderInput{
		{Party: application.PartyInput{Type: "user", UserID: uid}, Role: "manager"},
		{Party: application.PartyInput{Type: "external", ExternalName: "Owner Name", ExternalEmail: "owner@example.com"}, Role: "owner"},
	}
	res2, err := svc.RegisterProperty(context.Background(), cmd2, uid)
	if err != nil {
		t.Fatalf("option2: %v", err)
	}
	st2 := getStakeholders(res2.Property.ID.String())
	if len(st2) != 2 {
		t.Errorf("option2 expected 2, got %d: %v", len(st2), st2)
	}

	// Option 3: I am Owner delegating to another manager (self owner, external manager)
	cmd3 := validCmd()
	cmd3.DisplayName = "Opt3 Owner Delegates"
	cmd3.Stakeholders = []application.StakeholderInput{
		{Party: application.PartyInput{Type: "user", UserID: uid}, Role: "owner"},
		{Party: application.PartyInput{Type: "external", ExternalName: "Manager Name", ExternalEmail: "manager@example.com"}, Role: "manager"},
	}
	res3, err := svc.RegisterProperty(context.Background(), cmd3, uid)
	if err != nil {
		t.Fatalf("option3: %v", err)
	}
	st3 := getStakeholders(res3.Property.ID.String())
	if len(st3) != 2 {
		t.Errorf("option3 expected 2, got %d: %v", len(st3), st3)
	}

	// Option 4: Acting on behalf of an Organization (org owner + self manager)
	org, _ := identity.NewOrganization("Test Org", "test-org-opt4", identity.UserID(uuid.MustParse(uid)))
	if err := identitypgx.NewOrganizationRepo().Create(context.Background(), identitypgx.NewExecutorFromPool(testPool), org); err != nil {
		t.Fatalf("create org: %v", err)
	}
	membership, _ := identity.NewMembership(identity.UserID(uuid.MustParse(uid)), org.ID, identity.OrganizationRoleOwner)
	if err := identitypgx.NewMembershipRepo().Create(context.Background(), identitypgx.NewExecutorFromPool(testPool), membership); err != nil {
		t.Fatalf("create membership: %v", err)
	}
	cmd4 := validCmd()
	cmd4.DisplayName = "Opt4 Org Behalf"
	cmd4.Stakeholders = []application.StakeholderInput{
		{Party: application.PartyInput{Type: "organization", OrganizationID: uuid.UUID(org.ID).String()}, Role: "owner"},
		{Party: application.PartyInput{Type: "user", UserID: uid}, Role: "manager"},
	}
	res4, err := svc.RegisterProperty(context.Background(), cmd4, uid)
	if err != nil {
		t.Fatalf("option4: %v", err)
	}
	st4 := getStakeholders(res4.Property.ID.String())
	if len(st4) != 2 {
		t.Errorf("option4 expected 2, got %d: %v", len(st4), st4)
	}

	// Verify all four properties are accessible by actor
	list, err := svc.ListAccessible(context.Background(), db, uid)
	if err != nil {
		t.Fatalf("ListAccessible after 4 opts: %v", err)
	}
	if len(list) != 4 {
		t.Errorf("expected 4 accessible properties for actor, got %d", len(list))
	}
}

func TestService_AccessRule_UserDirectAndViaOrganizationAndExternalNoAccess(t *testing.T) {
	cleanAll(t)
	uidOwner := mustCreateUser(t)
	uidOrgMember := mustCreateUser(t)
	uidOutsider := mustCreateUser(t)
	uidExternal := mustCreateUser(t) // not relevant, external has no account

	svc := newServiceWithStakeholders()
	db := pgx.NewExecutorFromPool(testPool)

	// Create org and make uidOrgMember a member
	org, _ := identity.NewOrganization("Access Org", "access-org", identity.UserID(uuid.MustParse(uidOwner)))
	if err := identitypgx.NewOrganizationRepo().Create(context.Background(), identitypgx.NewExecutorFromPool(testPool), org); err != nil {
		t.Fatalf("create org: %v", err)
	}
	mem, _ := identity.NewMembership(identity.UserID(uuid.MustParse(uidOrgMember)), org.ID, identity.OrganizationRoleOwner)
	if err := identitypgx.NewMembershipRepo().Create(context.Background(), identitypgx.NewExecutorFromPool(testPool), mem); err != nil {
		t.Fatalf("create membership: %v", err)
	}

	// Property with UserRef direct stakeholder (uidOwner as owner)
	cmdDirect := validCmd()
	cmdDirect.DisplayName = "Direct User"
	cmdDirect.Stakeholders = []application.StakeholderInput{
		{Party: application.PartyInput{Type: "user", UserID: uidOwner}, Role: "owner"},
	}
	resDirect, err := svc.RegisterProperty(context.Background(), cmdDirect, uidOwner)
	if err != nil {
		t.Fatalf("create direct: %v", err)
	}

	// Property with OrganizationRef stakeholder
	cmdOrg := validCmd()
	cmdOrg.DisplayName = "Org Owned"
	cmdOrg.Stakeholders = []application.StakeholderInput{
		{Party: application.PartyInput{Type: "organization", OrganizationID: uuid.UUID(org.ID).String()}, Role: "owner"},
	}
	// actor is uidOwner who is creator of org but not member? Actually uidOwner created org but not member yet; for this test, actor needs to be member to tie. Let's make uidOwner also member? Simpler: actor for this property is uidOrgMember who is member, and org was created by uidOwner, but stakeholder is org. So actor uidOrgMember should have access.
	resOrg, err := svc.RegisterProperty(context.Background(), cmdOrg, uidOrgMember)
	if err != nil {
		t.Fatalf("create org stakeholder: %v", err)
	}

	// Property with ExternalParty only (no access for anyone)
	cmdExt := validCmd()
	cmdExt.DisplayName = "External Only"
	cmdExt.Stakeholders = []application.StakeholderInput{
		{Party: application.PartyInput{Type: "external", ExternalName: "Ext", ExternalEmail: "ext@example.com"}, Role: "owner"},
		// This should fail because no stakeholder ties actor (actor is uidExternal, but external email is different). So we need to make it fail validation.
	}
	_, err = svc.RegisterProperty(context.Background(), cmdExt, uidExternal)
	if err == nil {
		t.Fatal("expected error for external-only stakeholder (no tie to actor)")
	}
	if !isNoStakeholderError(err) {
		t.Errorf("expected ErrNoStakeholder, got %v", err)
	}

	// Now test access: uidOwner should see direct property
	listOwner, _ := svc.ListAccessible(context.Background(), db, uidOwner)
	if !containsProperty(listOwner, resDirect.Property.ID.String()) {
		t.Error("uidOwner should have access to direct property")
	}
	if containsProperty(listOwner, resOrg.Property.ID.String()) {
		t.Error("uidOwner should NOT have access to org property (not member)")
	}

	// uidOrgMember should see org property via membership
	listMember, _ := svc.ListAccessible(context.Background(), db, uidOrgMember)
	if !containsProperty(listMember, resOrg.Property.ID.String()) {
		t.Error("uidOrgMember should have access via organization")
	}
	if containsProperty(listMember, resDirect.Property.ID.String()) {
		t.Error("uidOrgMember should NOT have access to direct single-user property")
	}

	// outsider should see none
	listOutsider, _ := svc.ListAccessible(context.Background(), db, uidOutsider)
	if len(listOutsider) != 0 {
		t.Errorf("outsider expected 0, got %d", len(listOutsider))
	}

	// External party never confers access, so even if we create a property with external + self, outsider still not.
	// GetByID should return 404 for outsider
	_, err = svc.GetByID(context.Background(), db, resDirect.Property.ID.String(), uidOutsider)
	if err != properties.ErrPropertyNotFound {
		t.Errorf("expected ErrPropertyNotFound for outsider GetByID, got %v", err)
	}
	// GetByID for org member on org property should succeed
	_, err = svc.GetByID(context.Background(), db, resOrg.Property.ID.String(), uidOrgMember)
	if err != nil {
		t.Errorf("expected success for org member GetByID, got %v", err)
	}
}

func containsProperty(list []properties.Property, id string) bool {
	for _, p := range list {
		if p.ID.String() == id {
			return true
		}
	}
	return false
}

func isNoStakeholderError(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() != "" && (err.Error() == properties.ErrNoStakeholder.Error() || contains(err.Error(), properties.ErrNoStakeholder.Error()))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})()
}

func TestService_Stakeholder_Uniqueness_A12(t *testing.T) {
	cleanAll(t)
	uid := mustCreateUser(t)
	svc := newServiceWithStakeholders()

	// Same party same role duplicate should fail
	cmd := validCmd()
	cmd.DisplayName = "Dup Test"
	cmd.Stakeholders = []application.StakeholderInput{
		{Party: application.PartyInput{Type: "user", UserID: uid}, Role: "owner"},
		{Party: application.PartyInput{Type: "user", UserID: uid}, Role: "owner"},
	}
	_, err := svc.RegisterProperty(context.Background(), cmd, uid)
	if err == nil {
		t.Fatal("expected duplicate stakeholder error")
	}

	// Same party both roles should succeed (A-12)
	cmd2 := validCmd()
	cmd2.DisplayName = "Both Roles"
	cmd2.Stakeholders = []application.StakeholderInput{
		{Party: application.PartyInput{Type: "user", UserID: uid}, Role: "owner"},
		{Party: application.PartyInput{Type: "user", UserID: uid}, Role: "manager"},
	}
	if _, err := svc.RegisterProperty(context.Background(), cmd2, uid); err != nil {
		t.Fatalf("expected success for same party both roles, got %v", err)
	}
}

func TestService_Stakeholder_AtLeastOneTiesActor(t *testing.T) {
	cleanAll(t)
	uidActor := mustCreateUser(t)
	uidOther := mustCreateUser(t)
	svc := newServiceWithStakeholders()

	// Only other user as stakeholder, no tie to actor -> should fail
	cmd := validCmd()
	cmd.Stakeholders = []application.StakeholderInput{
		{Party: application.PartyInput{Type: "user", UserID: uidOther}, Role: "owner"},
	}
	_, err := svc.RegisterProperty(context.Background(), cmd, uidActor)
	if err == nil {
		t.Fatal("expected error for no tie to actor")
	}
	if !isNoStakeholderError(err) {
		t.Errorf("expected ErrNoStakeholder, got %v", err)
	}
}

