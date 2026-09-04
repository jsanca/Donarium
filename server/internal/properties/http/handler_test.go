package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"donarium/server/internal/identity"
	identityhttp "donarium/server/internal/identity/http"
	identitypgx "donarium/server/internal/identity/pgx"
	"donarium/server/internal/identity/application/authentication"
	"donarium/server/internal/platform/database"
	"donarium/server/internal/properties/application"
	propertieshttp "donarium/server/internal/properties/http"
	propertiespgx "donarium/server/internal/properties/pgx"

	"github.com/go-chi/chi/v5"
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
	for _, t := range []string{"property_stakeholders", "properties", "platform_grants", "memberships", "credentials", "organizations", "users", "schema_migrations"} {
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
	user, _ := identity.NewUser("http-"+uuid.NewString()+"@example.com", "HttpTest")
	if err := identitypgx.NewUserRepo().Create(context.Background(), identitypgx.NewExecutorFromPool(testPool), user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return uuid.UUID(user.ID).String()
}

func newHandler() *propertieshttp.Handler {
	repo := propertiespgx.NewRepository()
	stakeRepo := propertiespgx.NewStakeholderRepository()
	tx := propertiespgx.NewTransactionManager(testPool)
	svc := application.NewServiceWithStakeholders(repo, stakeRepo, tx)
	return propertieshttp.NewHandler(svc, testPool)
}

func withPrincipal(req *http.Request, userID string) *http.Request {
	principal := authentication.AuthenticatedPrincipal{
		UserID:      userID,
		DisplayName: "Test",
		Email:       "test@example.com",
	}
	ctx := identityhttp.WithPrincipal(req.Context(), principal)
	return req.WithContext(ctx)
}

func withChiParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	return req.WithContext(ctx)
}

func TestHandler_Unauthenticated_ListReturns401(t *testing.T) {
	cleanAll(t)
	h := newHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/properties", nil)
	h.ListProperties(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["error"] != "authentication required" {
		t.Errorf("unexpected error: %v", body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestHandler_Unauthenticated_CreateReturns401(t *testing.T) {
	cleanAll(t)
	h := newHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/properties", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	h.RegisterProperty(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_Create_HappyPath(t *testing.T) {
	cleanAll(t)
	uid := mustCreateUser(t)
	h := newHandler()

	payload := map[string]interface{}{
		"displayName":    "Casa Sol",
		"classification": "house",
		"address": map[string]string{
			"street":     "123 Calle",
			"city":       "Madrid",
			"postalCode": "28001",
			"country":    "ES",
		},
		"rentalCadence": "monthly",
		"standardRent":  120000,
	}
	body, _ := json.Marshal(payload)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/properties", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withPrincipal(req, uid)

	h.RegisterProperty(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["id"] == nil || resp["id"] == "" {
		t.Error("expected id in response")
	}
	if resp["displayName"] != "Casa Sol" {
		t.Errorf("displayName mismatch: %v", resp["displayName"])
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestHandler_Create_InvalidClassification(t *testing.T) {
	cleanAll(t)
	uid := mustCreateUser(t)
	h := newHandler()

	payload := map[string]interface{}{
		"displayName":    "Bad",
		"classification": "castle",
		"address": map[string]string{
			"street":     "S",
			"city":       "C",
			"postalCode": "12345",
			"country":    "ES",
		},
		"rentalCadence": "monthly",
		"standardRent":  100000,
	}
	body, _ := json.Marshal(payload)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/properties", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withPrincipal(req, uid)

	h.RegisterProperty(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	var errBody map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&errBody)
	if errBody["error"] != "classification is not valid" {
		t.Errorf("expected classification error, got %v", errBody)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestHandler_Create_MissingDisplayName(t *testing.T) {
	cleanAll(t)
	uid := mustCreateUser(t)
	h := newHandler()

	payload := map[string]interface{}{
		"displayName":    "",
		"classification": "house",
		"address": map[string]string{
			"street":     "S",
			"city":       "C",
			"postalCode": "12345",
			"country":    "ES",
		},
		"rentalCadence": "monthly",
		"standardRent":  100000,
	}
	body, _ := json.Marshal(payload)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/properties", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withPrincipal(req, uid)

	h.RegisterProperty(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_405_Collection(t *testing.T) {
	cleanAll(t)
	uid := mustCreateUser(t)
	h := newHandler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/properties", nil)
	req = withPrincipal(req, uid)
	h.Collection(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); allow != "GET, POST" {
		t.Errorf("expected Allow GET, POST, got %q", allow)
	}
	var errBody map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&errBody)
	if errBody["error"] != "method not allowed" {
		t.Errorf("expected method not allowed, got %v", errBody)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestHandler_405_GetProperty(t *testing.T) {
	cleanAll(t)
	uid := mustCreateUser(t)
	h := newHandler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/properties/some-id", nil)
	req = withPrincipal(req, uid)
	req = withChiParam(req, "id", "some-id")
	h.GetProperty(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); allow != http.MethodGet {
		t.Errorf("expected Allow GET, got %q", allow)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestHandler_List_EmptyAndFiltered(t *testing.T) {
	cleanAll(t)
	uid1 := mustCreateUser(t)
	uid2 := mustCreateUser(t)
	h := newHandler()

	// uid1 creates one
	payload := map[string]interface{}{
		"displayName":    "Owned By 1",
		"classification": "apartment",
		"address": map[string]string{
			"street":     "Street",
			"city":       "City",
			"postalCode": "12345",
			"country":    "ES",
		},
		"rentalCadence": "monthly",
		"standardRent":  100000,
	}
	body, _ := json.Marshal(payload)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/properties", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withPrincipal(req, uid1)
	h.RegisterProperty(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create uid1: %d %s", rec.Code, rec.Body.String())
	}

	// uid1 lists -> 1
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/properties", nil)
	req = withPrincipal(req, uid1)
	h.ListProperties(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list uid1: %d %s", rec.Code, rec.Body.String())
	}
	var listResp map[string][]map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode list uid1: %v", err)
	}
	if len(listResp["properties"]) != 1 {
		t.Errorf("expected 1 property for uid1, got %d", len(listResp["properties"]))
	}

	// uid2 lists -> 0 (does not see uid1's property)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/properties", nil)
	req = withPrincipal(req, uid2)
	h.ListProperties(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list uid2: %d %s", rec.Code, rec.Body.String())
	}
	var listResp2 map[string][]map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&listResp2); err != nil {
		t.Fatalf("decode list uid2: %v", err)
	}
	if listResp2["properties"] != nil && len(listResp2["properties"]) != 0 {
		t.Errorf("expected 0 for uid2, got %d", len(listResp2["properties"]))
	}
}

func TestHandler_GetProperty_AuthorizedAndUnauthorized(t *testing.T) {
	cleanAll(t)
	uid1 := mustCreateUser(t)
	uid2 := mustCreateUser(t)
	h := newHandler()

	payload := map[string]interface{}{
		"displayName":    "Private Casa",
		"classification": "house",
		"address": map[string]string{
			"street":     "Street",
			"city":       "City",
			"postalCode": "12345",
			"country":    "ES",
		},
		"rentalCadence": "monthly",
		"standardRent":  100000,
	}
	body, _ := json.Marshal(payload)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/properties", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withPrincipal(req, uid1)
	h.RegisterProperty(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}
	var created map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&created)
	id := created["id"].(string)

	// authorized get
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/properties/"+id, nil)
	req = withPrincipal(req, uid1)
	req = withChiParam(req, "id", id)
	h.GetProperty(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("authorized get: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var fetched map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&fetched)
	if fetched["id"] != id {
		t.Errorf("id mismatch: %v vs %q", fetched["id"], id)
	}

	// unauthorized get -> 404 (do not disclose)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/properties/"+id, nil)
	req = withPrincipal(req, uid2)
	req = withChiParam(req, "id", id)
	h.GetProperty(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unauthorized get: expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	var errBody map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&errBody)
	if errBody["error"] != "property not found" {
		t.Errorf("expected 'property not found', got %v", errBody)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}

	// unauthenticated -> 401
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/properties/"+id, nil)
	req = withChiParam(req, "id", id)
	h.GetProperty(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated: expected 401, got %d", rec.Code)
	}

	// invalid uuid also returns 404 (do not disclose)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/properties/not-a-uuid", nil)
	req = withPrincipal(req, uid1)
	req = withChiParam(req, "id", "not-a-uuid")
	h.GetProperty(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for invalid uuid, got %d", rec.Code)
	}
	_ = json.NewDecoder(rec.Body).Decode(&errBody)
	if errBody["error"] != "property not found" {
		t.Errorf("expected 'property not found', got %v", errBody)
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	cleanAll(t)
	uid := mustCreateUser(t)
	h := newHandler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/properties", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	req = withPrincipal(req, uid)
	h.RegisterProperty(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid json, got %d", rec.Code)
	}
	var errBody map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&errBody)
	if errBody["error"] != "invalid request body" {
		t.Errorf("unexpected error: %v", errBody)
	}
}
