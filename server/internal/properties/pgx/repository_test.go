package pgx_test

import (
	"context"
	"os"
	"testing"
	"time"

	"donarium/server/internal/identity"
	identitypgx "donarium/server/internal/identity/pgx"
	"donarium/server/internal/platform/database"
	"donarium/server/internal/properties"
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
	initializeSchema(ctx)
	code := m.Run()
	testPool.Close()
	os.Exit(code)
}

func initializeSchema(ctx context.Context) {
	tables := []string{"property_stakeholders", "properties", "platform_grants", "memberships", "credentials", "organizations", "users", "schema_migrations"}
	for _, t := range tables {
		_, _ = testPool.Exec(ctx, "DROP TABLE IF EXISTS "+t+" CASCADE")
	}
	if err := database.RunMigrations(ctx, testPool); err != nil {
		panic("migrations failed: " + err.Error())
	}
}

func db() properties.DBExecutor {
	return pgx.NewExecutorFromPool(testPool)
}

func cleanAll(t *testing.T) {
	t.Helper()
	tables := []string{"property_stakeholders", "properties", "platform_grants", "memberships", "credentials", "organizations", "users"}
	for _, tbl := range tables {
		_, err := testPool.Exec(context.Background(), "DELETE FROM "+tbl)
		if err != nil {
			t.Fatalf("delete %s: %v", tbl, err)
		}
	}
}

func mustCreateUser(t *testing.T) identity.UserID {
	t.Helper()
	user, _ := identity.NewUser("prop-test@example.com", "PropTester")
	// ensure unique email per call
	user.Email = "prop-" + uuid.NewString() + "@example.com"
	if err := identitypgx.NewUserRepo().Create(context.Background(), identitypgx.NewExecutorFromPool(testPool), user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user.ID
}

func TestPropertyRepo_CreateAndFindByID(t *testing.T) {
	cleanAll(t)
	uid := mustCreateUser(t)
	now := time.Now().UTC()
	prop := properties.Property{
		ID:             properties.NewPropertyID(),
		DisplayName:    "Casa Sol",
		Classification: properties.ClassificationHouse,
		Address: properties.Address{
			Street:     "123 Calle Principal",
			City:       "Madrid",
			State:      "Madrid",
			PostalCode: "28001",
			Country:    "ES",
		},
		RentalCadence: properties.CadenceMonthly,
		StandardRent:  120000,
		CreatedBy:     uuid.UUID(uid),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := pgx.NewRepository().Create(context.Background(), db(), prop); err != nil {
		t.Fatalf("Create: %v", err)
	}
	found, err := pgx.NewRepository().FindByID(context.Background(), db(), prop.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.DisplayName != "Casa Sol" {
		t.Errorf("displayName mismatch: %q", found.DisplayName)
	}
	if found.Classification != properties.ClassificationHouse {
		t.Errorf("classification mismatch: %q", found.Classification)
	}
	if found.StandardRent != 120000 {
		t.Errorf("rent mismatch: %d", found.StandardRent)
	}
	if found.CreatedBy != uuid.UUID(uid) {
		t.Errorf("createdBy mismatch")
	}
}

func TestPropertyRepo_FindAccessibleByUser(t *testing.T) {
	cleanAll(t)
	uid1 := mustCreateUser(t)
	uid2 := mustCreateUser(t)

	// uid1 creates two properties, uid2 creates one (with stakeholders for new access rule)
	stakeRepo := pgx.NewStakeholderRepository()
	for i := 0; i < 2; i++ {
		p := properties.Property{
			ID:             properties.NewPropertyID(),
			DisplayName:    "Prop A" + string(rune('0'+i)),
			Classification: properties.ClassificationApartment,
			Address: properties.Address{
				Street:     "Street",
				City:       "City",
				PostalCode: "12345",
				Country:    "ES",
			},
			RentalCadence: properties.CadenceMonthly,
			StandardRent:  100000,
			CreatedBy:     uuid.UUID(uid1),
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}
		if err := pgx.NewRepository().Create(context.Background(), db(), p); err != nil {
			t.Fatalf("create uid1 prop: %v", err)
		}
		uid := uuid.UUID(uid1)
		stake := properties.PropertyStakeholder{
			PropertyID: p.ID,
			Party:      properties.Party{Type: properties.PartyTypeUser, UserID: &uid},
			Role:       properties.StakeholderRoleOwner,
			CreatedAt:  time.Now().UTC(),
		}
		if err := stakeRepo.Create(context.Background(), db(), stake); err != nil {
			t.Fatalf("create stakeholder uid1: %v", err)
		}
	}

	p := properties.Property{
		ID:             properties.NewPropertyID(),
		DisplayName:    "Prop B",
		Classification: properties.ClassificationCommercial,
		Address: properties.Address{
			Street:     "Street B",
			City:       "City",
			PostalCode: "99999",
			Country:    "ES",
		},
		RentalCadence: properties.CadenceAnnual,
		StandardRent:  500000,
		CreatedBy:     uuid.UUID(uid2),
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	if err := pgx.NewRepository().Create(context.Background(), db(), p); err != nil {
		t.Fatalf("create uid2 prop: %v", err)
	}
	uid2Raw := uuid.UUID(uid2)
	stake2 := properties.PropertyStakeholder{
		PropertyID: p.ID,
		Party:      properties.Party{Type: properties.PartyTypeUser, UserID: &uid2Raw},
		Role:       properties.StakeholderRoleOwner,
		CreatedAt:  time.Now().UTC(),
	}
	if err := stakeRepo.Create(context.Background(), db(), stake2); err != nil {
		t.Fatalf("create stakeholder uid2: %v", err)
	}

	list1, err := pgx.NewRepository().FindAccessibleByUser(context.Background(), db(), uuid.UUID(uid1).String())
	if err != nil {
		t.Fatalf("FindAccessibleByUser uid1: %v", err)
	}
	if len(list1) != 2 {
		t.Errorf("expected 2 for uid1, got %d", len(list1))
	}

	list2, err := pgx.NewRepository().FindAccessibleByUser(context.Background(), db(), uuid.UUID(uid2).String())
	if err != nil {
		t.Fatalf("FindAccessibleByUser uid2: %v", err)
	}
	if len(list2) != 1 {
		t.Errorf("expected 1 for uid2, got %d", len(list2))
	}

	// unknown user gets empty
	unknown := uuid.NewString()
	list3, err := pgx.NewRepository().FindAccessibleByUser(context.Background(), db(), unknown)
	if err != nil {
		t.Fatalf("FindAccessibleByUser unknown: %v", err)
	}
	if len(list3) != 0 {
		t.Errorf("expected 0 for unknown, got %d", len(list3))
	}
}

func TestPropertyRepo_FindByID_NotFound(t *testing.T) {
	cleanAll(t)
	_, err := pgx.NewRepository().FindByID(context.Background(), db(), properties.NewPropertyID())
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if err != properties.ErrPropertyNotFound {
		// pgx may wrap? check via errors.Is in handler but repo returns directly
		t.Errorf("expected ErrPropertyNotFound, got %v", err)
	}
}
