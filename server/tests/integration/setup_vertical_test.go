package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/application"
	httphandler "donarium/server/internal/identity/http"
	identitypgx "donarium/server/internal/identity/pgx"
	"donarium/server/internal/platform/database"

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

func cleanAll(t *testing.T) {
	t.Helper()
	tables := []string{"platform_grants", "memberships", "credentials", "organizations", "users"}
	for _, tbl := range tables {
		_, err := testPool.Exec(context.Background(), "DELETE FROM "+tbl)
		if err != nil {
			t.Fatalf("delete %s: %v", tbl, err)
		}
	}
}

func newTestHandler() *httphandler.SetupHandler {
	userRepo := identitypgx.NewUserRepo()
	credRepo := identitypgx.NewCredentialRepo()
	orgRepo := identitypgx.NewOrganizationRepo()
	memRepo := identitypgx.NewMembershipRepo()
	grantRepo := identitypgx.NewPlatformGrantRepo()
	hasher := identitypgx.NewArgon2Hasher()
	normalizer := identity.NewDefaultEmailNormalizer()

	canonical := application.NewCanonicalSetupService(
		userRepo, credRepo, orgRepo, memRepo, grantRepo,
		hasher, normalizer,
	)
	txManager := identitypgx.NewTransactionManager(testPool)
	txSetup := application.NewTransactionalSetupService(canonical, txManager)

	statusReader := &testStatusReader{
		pool:    testPool,
		orgRepo: orgRepo,
	}

	return httphandler.NewSetupHandler(txSetup, statusReader)
}

type testStatusReader struct {
	pool    *pgxpool.Pool
	orgRepo *identitypgx.OrganizationRepo
}

func (r *testStatusReader) IsInitialized(ctx context.Context) (bool, error) {
	return r.orgRepo.ExistsAny(ctx, identitypgx.NewExecutorFromPool(r.pool))
}

func buildSetupRequest() []byte {
	body, _ := json.Marshal(httphandler.SetupRequest{
		DisplayName:      "Owner",
		Email:            "owner@donarium.test",
		Password:         "ValidP@ss1",
		OrganizationName: "Donarium",
		OrganizationSlug: "donarium",
	})
	return body
}

func doSetup(t *testing.T, handler *httphandler.SetupHandler) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/setup", bytes.NewReader(buildSetupRequest()))
	req.Header.Set("Content-Type", "application/json")
	handler.Setup(rec, req)
	return rec
}

func doStatus(t *testing.T, handler *httphandler.SetupHandler) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/setup/status", nil)
	handler.Status(rec, req)
	return rec
}

func TestVerticalSlice_EmptyDatabase(t *testing.T) {
	cleanAll(t)
	handler := newTestHandler()

	rec := doStatus(t, handler)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var status httphandler.SetupStatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if status.Initialized {
		t.Error("expected initialized=false on empty database")
	}
}

func TestVerticalSlice_InitialSetup(t *testing.T) {
	cleanAll(t)
	handler := newTestHandler()

	rec := doSetup(t, handler)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp httphandler.SetupResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.UserID == "" {
		t.Error("expected non-empty userId")
	}
	if resp.OrganizationID == "" {
		t.Error("expected non-empty organizationId")
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestVerticalSlice_EntitiesPersisted(t *testing.T) {
	cleanAll(t)
	handler := newTestHandler()

	rec := doSetup(t, handler)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup failed: %d - %s", rec.Code, rec.Body.String())
	}

	var resp httphandler.SetupResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode setup response: %v", err)
	}

	exec := identitypgx.NewExecutorFromPool(testPool)

	orgRepo := identitypgx.NewOrganizationRepo()
	orgsExist, err := orgRepo.ExistsAny(context.Background(), exec)
	if err != nil {
		t.Fatalf("ExistsAny: %v", err)
	}
	if !orgsExist {
		t.Fatal("expected at least one organization to exist")
	}

	userRepo := identitypgx.NewUserRepo()
	foundUser, err := userRepo.FindByEmail(context.Background(), exec, "owner@donarium.test")
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if foundUser.DisplayName != "Owner" {
		t.Errorf("expected DisplayName 'Owner', got %q", foundUser.DisplayName)
	}

	credRepo := identitypgx.NewCredentialRepo()
	_, err = credRepo.FindByUserID(context.Background(), exec, foundUser.ID)
	if err != nil {
		t.Fatalf("FindByUserID: %v", err)
	}

	org, err := orgRepo.FindBySlug(context.Background(), exec, "donarium")
	if err != nil {
		t.Fatalf("FindBySlug: %v", err)
	}
	if org.Name != "Donarium" {
		t.Errorf("expected org name 'Donarium', got %q", org.Name)
	}

	memRepo := identitypgx.NewMembershipRepo()
	membership, err := memRepo.FindByUserAndOrg(context.Background(), exec, foundUser.ID, org.ID)
	if err != nil {
		t.Fatalf("FindByUserAndOrg: %v", err)
	}
	if membership.Role != identity.OrganizationRoleOwner {
		t.Errorf("expected owner role, got %q", membership.Role)
	}

	grantRepo := identitypgx.NewPlatformGrantRepo()
	grant, err := grantRepo.FindByUser(context.Background(), exec, foundUser.ID)
	if err != nil {
		t.Fatalf("FindByUser: %v", err)
	}
	if grant.Role != identity.PlatformRoleSuperAdmin {
		t.Errorf("expected super_admin role, got %q", grant.Role)
	}
}

func TestVerticalSlice_StatusIsInitializedAfterSetup(t *testing.T) {
	cleanAll(t)
	handler := newTestHandler()

	rec := doSetup(t, handler)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup failed: %d", rec.Code)
	}

	statusRec := doStatus(t, handler)

	if statusRec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", statusRec.Code)
	}
	var status httphandler.SetupStatusResponse
	if err := json.NewDecoder(statusRec.Body).Decode(&status); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if !status.Initialized {
		t.Error("expected initialized=true after setup")
	}
}

func TestVerticalSlice_DuplicateSetupReturns409(t *testing.T) {
	cleanAll(t)
	handler := newTestHandler()

	rec1 := doSetup(t, handler)
	if rec1.Code != http.StatusCreated {
		t.Fatalf("first setup failed: %d - %s", rec1.Code, rec1.Body.String())
	}

	rec2 := doSetup(t, handler)

	if rec2.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", rec2.Code, rec2.Body.String())
	}

	var errResp httphandler.ErrorResponse
	if err := json.NewDecoder(rec2.Body).Decode(&errResp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if errResp.Error != "system is already initialized" {
		t.Errorf("expected 'system is already initialized', got %q", errResp.Error)
	}
}

func TestVerticalSlice_NoPartialWritesOnError(t *testing.T) {
	cleanAll(t)
	handler := newTestHandler()

	rec := doSetup(t, handler)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first setup failed: %d", rec.Code)
	}

	exec := identitypgx.NewExecutorFromPool(testPool)
	orgRepo := identitypgx.NewOrganizationRepo()

	orgsAfter, _ := orgRepo.ExistsAny(context.Background(), exec)
	if !orgsAfter {
		t.Fatal("expected organization to exist")
	}

	var count int
	row := identitypgx.NewExecutorFromPool(testPool).QueryRow(context.Background(), "SELECT COUNT(*) FROM users")
	if err := row.Scan(&count); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 user, got %d", count)
	}

	var credCount int
	credRow := identitypgx.NewExecutorFromPool(testPool).QueryRow(context.Background(), "SELECT COUNT(*) FROM credentials")
	if err := credRow.Scan(&credCount); err != nil {
		t.Fatalf("count creds: %v", err)
	}
	if credCount != 1 {
		t.Errorf("expected 1 credential, got %d", credCount)
	}

	var orgCount int
	orgRow := identitypgx.NewExecutorFromPool(testPool).QueryRow(context.Background(), "SELECT COUNT(*) FROM organizations")
	if err := orgRow.Scan(&orgCount); err != nil {
		t.Fatalf("count orgs: %v", err)
	}
	if orgCount != 1 {
		t.Errorf("expected 1 organization, got %d", orgCount)
	}

	var memCount int
	memRow := identitypgx.NewExecutorFromPool(testPool).QueryRow(context.Background(), "SELECT COUNT(*) FROM memberships")
	if err := memRow.Scan(&memCount); err != nil {
		t.Fatalf("count memberships: %v", err)
	}
	if memCount != 1 {
		t.Errorf("expected 1 membership, got %d", memCount)
	}

	var grantCount int
	grantRow := identitypgx.NewExecutorFromPool(testPool).QueryRow(context.Background(), "SELECT COUNT(*) FROM platform_grants")
	if err := grantRow.Scan(&grantCount); err != nil {
		t.Fatalf("count platform_grants: %v", err)
	}
	if grantCount != 1 {
		t.Errorf("expected 1 platform grant, got %d", grantCount)
	}
}

func TestVerticalSlice_RestartPreservesState(t *testing.T) {
	cleanAll(t)
	handler := newTestHandler()

	rec := doSetup(t, handler)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup failed: %d", rec.Code)
	}

	handler2 := newTestHandler()
	statusRec := doStatus(t, handler2)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("status after restart: %d", statusRec.Code)
	}
	var status2 httphandler.SetupStatusResponse
	if err := json.NewDecoder(statusRec.Body).Decode(&status2); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if !status2.Initialized {
		t.Error("expected initialized=true after handler restart")
	}

	dupRec := doSetup(t, handler2)
	if dupRec.Code != http.StatusConflict {
		t.Errorf("expected 409 on duplicate after restart, got %d: %s", dupRec.Code, dupRec.Body.String())
	}
}

func TestVerticalSlice_SetuWithNewPool(t *testing.T) {
	cleanAll(t)
	handler := newTestHandler()

	rec := doSetup(t, handler)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup failed: %d", rec.Code)
	}

	exec := identitypgx.NewExecutorFromPool(testPool)
	orgRepo := identitypgx.NewOrganizationRepo()
	ok, err := orgRepo.ExistsAny(context.Background(), exec)
	if err != nil {
		t.Fatalf("ExistsAny: %v", err)
	}
	if !ok {
		t.Fatal("expected initialized=true through new executor")
	}
}

func TestVerticalSlice_SecondSetupDoesNotModifyData(t *testing.T) {
	cleanAll(t)
	handler := newTestHandler()

	rec1 := doSetup(t, handler)
	if rec1.Code != http.StatusCreated {
		t.Fatalf("first setup failed: %d", rec1.Code)
	}
	_ = json.NewDecoder(rec1.Body).Decode(&httphandler.SetupResponse{})

	rec2 := doSetup(t, handler)
	if rec2.Code != http.StatusConflict {
		t.Fatalf("expected 409: got %d", rec2.Code)
	}

	var count int
	row := identitypgx.NewExecutorFromPool(testPool).QueryRow(context.Background(), "SELECT COUNT(*) FROM users")
	if err := row.Scan(&count); err != nil {
		t.Fatalf("scan users count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 user after duplicate attempt, got %d", count)
	}

	row2 := identitypgx.NewExecutorFromPool(testPool).QueryRow(context.Background(), "SELECT COUNT(*) FROM organizations")
	if err := row2.Scan(&count); err != nil {
		t.Fatalf("scan orgs count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 org after duplicate attempt, got %d", count)
	}
}

func TestVerticalSlice_ValidationErrors(t *testing.T) {
	cleanAll(t)
	handler := newTestHandler()

	tests := []struct {
		name     string
		body     httphandler.SetupRequest
		wantCode int
	}{
		{
			name:     "missing displayName",
			body:     httphandler.SetupRequest{Email: "a@b.com", Password: "ValidP@ss1", OrganizationName: "Org", OrganizationSlug: "org"},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "missing email",
			body:     httphandler.SetupRequest{DisplayName: "Owner", Password: "ValidP@ss1", OrganizationName: "Org", OrganizationSlug: "org"},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "missing password",
			body:     httphandler.SetupRequest{DisplayName: "Owner", Email: "a@b.com", OrganizationName: "Org", OrganizationSlug: "org"},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "missing organizationName",
			body:     httphandler.SetupRequest{DisplayName: "Owner", Email: "a@b.com", Password: "ValidP@ss1", OrganizationSlug: "org"},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "missing organizationSlug",
			body:     httphandler.SetupRequest{DisplayName: "Owner", Email: "a@b.com", Password: "ValidP@ss1", OrganizationName: "Org"},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/setup", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			handler.Setup(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("expected %d, got %d: %s", tt.wantCode, rec.Code, rec.Body.String())
			}

			var errResp httphandler.ErrorResponse
			_ = json.NewDecoder(rec.Body).Decode(&errResp)
			if errResp.Error == "" {
				t.Error("expected error message in response")
			}
		})
	}
}

func TestVerticalSlice_HTTPErrorMapping(t *testing.T) {
	cleanAll(t)
	handler := newTestHandler()

	doSetup(t, handler)

	rec := doSetup(t, handler)
	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rec.Code)
	}

	var errResp httphandler.ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error != "system is already initialized" {
		t.Errorf("unexpected error message: %q", errResp.Error)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json content type on error, got %q", ct)
	}
}

func TestVerticalSlice_WeakPassword(t *testing.T) {
	cleanAll(t)
	handler := newTestHandler()

	body, _ := json.Marshal(httphandler.SetupRequest{
		DisplayName:      "Owner",
		Email:            "weak@donarium.test",
		Password:         "short",
		OrganizationName: "Org",
		OrganizationSlug: "org",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.Setup(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for weak password, got %d: %s", rec.Code, rec.Body.String())
	}

	var errResp httphandler.ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error != "password does not meet requirements" {
		t.Errorf("unexpected error message: %q", errResp.Error)
	}
}

func TestVerticalSlice_InvalidEmail(t *testing.T) {
	cleanAll(t)
	handler := newTestHandler()

	body, _ := json.Marshal(httphandler.SetupRequest{
		DisplayName:      "Owner",
		Email:            "not-an-email",
		Password:         "ValidP@ss1",
		OrganizationName: "Org",
		OrganizationSlug: "org",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.Setup(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid email, got %d: %s", rec.Code, rec.Body.String())
	}

	var errResp httphandler.ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error != "email is not valid" {
		t.Errorf("unexpected error message: %q", errResp.Error)
	}
}

func TestVerticalSlice_InvalidSlug(t *testing.T) {
	cleanAll(t)
	handler := newTestHandler()

	body, _ := json.Marshal(httphandler.SetupRequest{
		DisplayName:      "Owner",
		Email:            "slug@donarium.test",
		Password:         "ValidP@ss1",
		OrganizationName: "Org",
		OrganizationSlug: "UPPER CASE",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.Setup(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid slug, got %d: %s", rec.Code, rec.Body.String())
	}

	var errResp httphandler.ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error != "slug is not valid" {
		t.Errorf("unexpected error message: %q", errResp.Error)
	}
}

func TestVerticalSlice_MalformedJSON(t *testing.T) {
	cleanAll(t)
	handler := newTestHandler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/setup", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	handler.Setup(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for malformed JSON, got %d", rec.Code)
	}
}

func TestVerticalSlice_TransactionRollbackPreservesConsistency(t *testing.T) {
	cleanAll(t)
	handler := newTestHandler()

	rec := doSetup(t, handler)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first setup failed: %d", rec.Code)
	}

	exec := identitypgx.NewExecutorFromPool(testPool)

	var userCount, credCount, orgCount, memCount, grantCount int
	if err := exec.QueryRow(context.Background(), "SELECT COUNT(*) FROM users").Scan(&userCount); err != nil {
		userCount = -1
	}
	if err := exec.QueryRow(context.Background(), "SELECT COUNT(*) FROM credentials").Scan(&credCount); err != nil {
		credCount = -1
	}
	if err := exec.QueryRow(context.Background(), "SELECT COUNT(*) FROM organizations").Scan(&orgCount); err != nil {
		orgCount = -1
	}
	if err := exec.QueryRow(context.Background(), "SELECT COUNT(*) FROM memberships").Scan(&memCount); err != nil {
		memCount = -1
	}
	if err := exec.QueryRow(context.Background(), "SELECT COUNT(*) FROM platform_grants").Scan(&grantCount); err != nil {
		grantCount = -1
	}

	if userCount != 1 || credCount != 1 || orgCount != 1 || memCount != 1 || grantCount != 1 {
		t.Errorf("expected 1 of each, got user=%d cred=%d org=%d mem=%d grant=%d",
			userCount, credCount, orgCount, memCount, grantCount)
	}

	dupRec := doSetup(t, handler)
	if dupRec.Code != http.StatusConflict {
		t.Fatalf("expected 409 on duplicate, got %d", dupRec.Code)
	}

	if err := exec.QueryRow(context.Background(), "SELECT COUNT(*) FROM users").Scan(&userCount); err != nil {
		userCount = -1
	}
	if err := exec.QueryRow(context.Background(), "SELECT COUNT(*) FROM credentials").Scan(&credCount); err != nil {
		credCount = -1
	}
	if err := exec.QueryRow(context.Background(), "SELECT COUNT(*) FROM organizations").Scan(&orgCount); err != nil {
		orgCount = -1
	}
	if err := exec.QueryRow(context.Background(), "SELECT COUNT(*) FROM memberships").Scan(&memCount); err != nil {
		memCount = -1
	}
	if err := exec.QueryRow(context.Background(), "SELECT COUNT(*) FROM platform_grants").Scan(&grantCount); err != nil {
		grantCount = -1
	}

	if userCount != 1 || credCount != 1 || orgCount != 1 || memCount != 1 || grantCount != 1 {
		t.Errorf("expected still 1 of each after failed duplicate, got user=%d cred=%d org=%d mem=%d grant=%d",
			userCount, credCount, orgCount, memCount, grantCount)
	}
}

func TestVerticalSlice_StatusInternalError(t *testing.T) {
	handler := httphandler.NewSetupHandler(nil, &errorStatusReader{err: errors.New("db down")})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/setup/status", nil)
	handler.Status(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestVerticalSlice_RealRollbackOnExecFailure(t *testing.T) {
	cleanAll(t)

	tx, err := testPool.Begin(context.Background())
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	realExec := identitypgx.NewExecutorFromPool(tx)
	failing := &failingExecutor{real: realExec, failOnExec: 2}

	canonical := application.NewCanonicalSetupService(
		identitypgx.NewUserRepo(),
		identitypgx.NewCredentialRepo(),
		identitypgx.NewOrganizationRepo(),
		identitypgx.NewMembershipRepo(),
		identitypgx.NewPlatformGrantRepo(),
		identitypgx.NewArgon2Hasher(),
		identity.NewDefaultEmailNormalizer(),
	)

	cmd := application.InitialOwnerSetupCommand{
		DisplayName:      "Owner",
		Email:            "rollback@donarium.test",
		Password:         "ValidP@ss1",
		OrganizationName: "Donarium",
		OrganizationSlug: "donarium",
	}

	_, execErr := canonical.Execute(context.Background(), failing, cmd)
	if execErr == nil {
		tx.Rollback(context.Background())
		t.Fatal("expected induced exec failure, got nil")
	}

	_ = tx.Rollback(context.Background())

	poolExec := identitypgx.NewExecutorFromPool(testPool)

	var count int
	if err := poolExec.QueryRow(context.Background(), "SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 users after rollback, got %d", count)
	}

	if err := poolExec.QueryRow(context.Background(), "SELECT COUNT(*) FROM credentials").Scan(&count); err != nil {
		t.Fatalf("count credentials: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 credentials after rollback, got %d", count)
	}

	if err := poolExec.QueryRow(context.Background(), "SELECT COUNT(*) FROM organizations").Scan(&count); err != nil {
		t.Fatalf("count organizations: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 organizations after rollback, got %d", count)
	}

	if err := poolExec.QueryRow(context.Background(), "SELECT COUNT(*) FROM memberships").Scan(&count); err != nil {
		t.Fatalf("count memberships: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 memberships after rollback, got %d", count)
	}

	if err := poolExec.QueryRow(context.Background(), "SELECT COUNT(*) FROM platform_grants").Scan(&count); err != nil {
		t.Fatalf("count platform_grants: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 platform_grants after rollback, got %d", count)
	}
}

type failingExecutor struct {
	real       identity.DBExecutor
	failOnExec int
	execCount  int
	mu         sync.Mutex
}

func (f *failingExecutor) Exec(ctx context.Context, sql string, arguments ...any) (int64, error) {
	f.mu.Lock()
	f.execCount++
	count := f.execCount
	f.mu.Unlock()
	if count >= f.failOnExec {
		return 0, errors.New("induced exec failure for rollback test")
	}
	return f.real.Exec(ctx, sql, arguments...)
}

func (f *failingExecutor) Query(ctx context.Context, sql string, args ...any) (identity.Rows, error) {
	return f.real.Query(ctx, sql, args...)
}

func (f *failingExecutor) QueryRow(ctx context.Context, sql string, args ...any) identity.RowScanner {
	return f.real.QueryRow(ctx, sql, args...)
}

type errorStatusReader struct {
	err error
}

func (r *errorStatusReader) IsInitialized(ctx context.Context) (bool, error) {
	return false, r.err
}

type failingOrgRepo struct {
	identity.OrganizationRepository
}

func (r *failingOrgRepo) Create(ctx context.Context, db identity.DBExecutor, org identity.Organization) error {
	if err := r.OrganizationRepository.Create(ctx, db, org); err != nil {
		return err
	}
	return errors.New("induced failure after org creation for rollback test")
}

func TestTC_001_09_RepositoryDecoratorRollback(t *testing.T) {
	cleanAll(t)

	canonical := application.NewCanonicalSetupService(
		identitypgx.NewUserRepo(),
		identitypgx.NewCredentialRepo(),
		&failingOrgRepo{OrganizationRepository: identitypgx.NewOrganizationRepo()},
		identitypgx.NewMembershipRepo(),
		identitypgx.NewPlatformGrantRepo(),
		identitypgx.NewArgon2Hasher(),
		identity.NewDefaultEmailNormalizer(),
	)

	txManager := identitypgx.NewTransactionManager(testPool)
	txSetup := application.NewTransactionalSetupService(canonical, txManager)

	cmd := application.InitialOwnerSetupCommand{
		DisplayName:      "Owner",
		Email:            "rollback-repo@donarium.test",
		Password:         "ValidP@ss1",
		OrganizationName: "Donarium",
		OrganizationSlug: "donarium",
	}

	_, err := txSetup.Execute(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected induced failure, got nil")
	}

	poolExec := identitypgx.NewExecutorFromPool(testPool)

	var count int
	if err := poolExec.QueryRow(context.Background(), "SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 users after rollback, got %d", count)
	}

	if err := poolExec.QueryRow(context.Background(), "SELECT COUNT(*) FROM credentials").Scan(&count); err != nil {
		t.Fatalf("count credentials: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 credentials after rollback, got %d", count)
	}

	if err := poolExec.QueryRow(context.Background(), "SELECT COUNT(*) FROM organizations").Scan(&count); err != nil {
		t.Fatalf("count organizations: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 organizations after rollback, got %d", count)
	}

	if err := poolExec.QueryRow(context.Background(), "SELECT COUNT(*) FROM memberships").Scan(&count); err != nil {
		t.Fatalf("count memberships: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 memberships after rollback, got %d", count)
	}

	if err := poolExec.QueryRow(context.Background(), "SELECT COUNT(*) FROM platform_grants").Scan(&count); err != nil {
		t.Fatalf("count platform_grants: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 platform_grants after rollback, got %d", count)
	}
}
