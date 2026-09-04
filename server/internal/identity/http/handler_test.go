package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/application"
	httphandler "donarium/server/internal/identity/http"
)

type fakePerformer struct {
	result application.InitialOwnerSetupResult
	err    error
}

func (f *fakePerformer) Execute(ctx context.Context, cmd application.InitialOwnerSetupCommand) (application.InitialOwnerSetupResult, error) {
	return f.result, f.err
}

type fakeStatus struct {
	initialized bool
	err         error
}

func (f *fakeStatus) IsInitialized(ctx context.Context) (bool, error) {
	return f.initialized, f.err
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func buildBody() []byte {
	return mustJSON(httphandler.SetupRequest{
		DisplayName:      "Owner",
		Email:            "owner@example.com",
		Password:         "ValidP@ss1",
		OrganizationName: "My Org",
		OrganizationSlug: "my-org",
	})
}

func TestSetup_Success(t *testing.T) {
	performer := &fakePerformer{
		result: application.InitialOwnerSetupResult{
			UserID:         identity.NewUserID(),
			OrganizationID: identity.NewOrganizationID(),
		},
	}
	handler := httphandler.NewSetupHandler(performer, &fakeStatus{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/setup", bytes.NewReader(buildBody()))
	req.Header.Set("Content-Type", "application/json")
	handler.Setup(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp httphandler.SetupResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.UserID == "" || resp.OrganizationID == "" {
		t.Error("expected non-empty IDs")
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestSetup_AlreadyInitialized(t *testing.T) {
	performer := &fakePerformer{
		err: identity.ErrAlreadyInitialized,
	}
	handler := httphandler.NewSetupHandler(performer, &fakeStatus{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/setup", bytes.NewReader(buildBody()))
	req.Header.Set("Content-Type", "application/json")
	handler.Setup(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSetup_DuplicateEmail(t *testing.T) {
	performer := &fakePerformer{
		err: identity.ErrDuplicateEmail,
	}
	handler := httphandler.NewSetupHandler(performer, &fakeStatus{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/setup", bytes.NewReader(buildBody()))
	req.Header.Set("Content-Type", "application/json")
	handler.Setup(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSetup_DuplicateSlug(t *testing.T) {
	performer := &fakePerformer{
		err: identity.ErrDuplicateSlug,
	}
	handler := httphandler.NewSetupHandler(performer, &fakeStatus{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/setup", bytes.NewReader(buildBody()))
	req.Header.Set("Content-Type", "application/json")
	handler.Setup(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSetup_InvalidEmail(t *testing.T) {
	performer := &fakePerformer{
		err: identity.ErrInvalidEmail,
	}
	handler := httphandler.NewSetupHandler(performer, &fakeStatus{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/setup", bytes.NewReader(buildBody()))
	req.Header.Set("Content-Type", "application/json")
	handler.Setup(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSetup_InvalidPassword(t *testing.T) {
	performer := &fakePerformer{
		err: identity.ErrInvalidPassword,
	}
	handler := httphandler.NewSetupHandler(performer, &fakeStatus{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/setup", bytes.NewReader(buildBody()))
	req.Header.Set("Content-Type", "application/json")
	handler.Setup(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSetup_MalformedJSON(t *testing.T) {
	handler := httphandler.NewSetupHandler(&fakePerformer{}, &fakeStatus{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/setup", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	handler.Setup(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSetup_MissingDisplayName(t *testing.T) {
	handler := httphandler.NewSetupHandler(&fakePerformer{}, &fakeStatus{})

	body := mustJSON(httphandler.SetupRequest{
		Email:            "a@b.com",
		Password:         "ValidP@ss1",
		OrganizationName: "Org",
		OrganizationSlug: "org",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.Setup(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStatus_NotInitialized(t *testing.T) {
	reader := &fakeStatus{initialized: false}
	handler := httphandler.NewSetupHandler(&fakePerformer{}, reader)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/setup/status", nil)
	handler.Status(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var resp httphandler.SetupStatusResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Initialized {
		t.Error("expected initialized=false")
	}
}

func TestStatus_Initialized(t *testing.T) {
	reader := &fakeStatus{initialized: true}
	handler := httphandler.NewSetupHandler(&fakePerformer{}, reader)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/setup/status", nil)
	handler.Status(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var resp2 httphandler.SetupStatusResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp2)
	if !resp2.Initialized {
		t.Error("expected initialized=true")
	}
}

func TestSetup_InternalError(t *testing.T) {
	performer := &fakePerformer{
		err: context.DeadlineExceeded,
	}
	handler := httphandler.NewSetupHandler(performer, &fakeStatus{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/setup", bytes.NewReader(buildBody()))
	req.Header.Set("Content-Type", "application/json")
	handler.Setup(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestSetup_WrongMethod(t *testing.T) {
	handler := httphandler.NewSetupHandler(&fakePerformer{}, &fakeStatus{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/setup", nil)
	handler.Setup(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	allow := rec.Header().Get("Allow")
	if allow != http.MethodPost {
		t.Errorf("expected Allow: POST, got %q", allow)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type: application/json, got %q", ct)
	}
	var errResp httphandler.ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error != "method not allowed" {
		t.Errorf("expected 'method not allowed', got %q", errResp.Error)
	}
}

func TestStatus_WrongMethod(t *testing.T) {
	handler := httphandler.NewSetupHandler(&fakePerformer{}, &fakeStatus{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/setup/status", nil)
	handler.Status(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	allow := rec.Header().Get("Allow")
	if allow != http.MethodGet {
		t.Errorf("expected Allow: GET, got %q", allow)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type: application/json, got %q", ct)
	}
	var errResp httphandler.ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error != "method not allowed" {
		t.Errorf("expected 'method not allowed', got %q", errResp.Error)
	}
}
