package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/application/authentication"
	httphandler "donarium/server/internal/identity/http"
)

type fakeLoginPerformer struct {
	result authentication.AuthenticatedPrincipal
	err    error
}

func (f *fakeLoginPerformer) Execute(ctx context.Context, db identity.DBExecutor, cmd authentication.AuthenticateUserCommand) (authentication.AuthenticatedPrincipal, error) {
	return f.result, f.err
}

type fakeCookieWriter struct {
	token string
}

func (w *fakeCookieWriter) Write(_ http.ResponseWriter, token string) { w.token = token }
func (w *fakeCookieWriter) Clear(_ http.ResponseWriter)               { w.token = "" }

func newTestCookieWriter() *fakeCookieWriter {
	return &fakeCookieWriter{}
}

func buildLoginBody(email, password string) []byte {
	b, _ := json.Marshal(httphandler.LoginRequest{Email: email, Password: password})
	return b
}

func TestLogin_Success(t *testing.T) {
	performer := &fakeLoginPerformer{
		result: authentication.AuthenticatedPrincipal{
			SessionToken:  "token123",
			UserID:        "user-uuid",
			DisplayName:   "Owner",
			Email:         "owner@test.com",
			PlatformRoles: []string{"super_admin"},
			OrganizationContexts: []authentication.OrganizationContext{
				{OrganizationID: "org-uuid", OrganizationName: "Org", OrganizationSlug: "org", Role: "owner"},
			},
			DefaultContext: authentication.DefaultContext{Type: "organization", OrganizationID: "org-uuid", Role: "owner"},
		},
	}

	cw := newTestCookieWriter()
	handler := httphandler.NewLoginHandler(performer, nil, cw)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(buildLoginBody("a@b.com", "pass")))
	req.Header.Set("Content-Type", "application/json")
	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp httphandler.LoginResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Principal.Email != "owner@test.com" {
		t.Errorf("email mismatch: %q", resp.Principal.Email)
	}
	if len(resp.Principal.OrganizationContexts) != 1 {
		t.Errorf("expected 1 org context, got %d", len(resp.Principal.OrganizationContexts))
	}
	if cw.token != "token123" {
		t.Error("expected cookie writer to receive session token")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	performer := &fakeLoginPerformer{err: identity.ErrInvalidCredentials}
	handler := httphandler.NewLoginHandler(performer, nil, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(buildLoginBody("a@b.com", "wrong")))
	req.Header.Set("Content-Type", "application/json")
	handler.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}

	var errResp httphandler.ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error != "the email or password is incorrect" {
		t.Errorf("unexpected message: %q", errResp.Error)
	}
}

func TestLogin_InvalidEmail(t *testing.T) {
	performer := &fakeLoginPerformer{err: identity.ErrInvalidEmail}
	handler := httphandler.NewLoginHandler(performer, nil, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(buildLoginBody("bad", "pass")))
	req.Header.Set("Content-Type", "application/json")
	handler.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestLogin_MissingEmail(t *testing.T) {
	handler := httphandler.NewLoginHandler(&fakeLoginPerformer{}, nil, nil)

	rec := httptest.NewRecorder()
	body := buildLoginBody("", "pass")
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestLogin_MissingPassword(t *testing.T) {
	handler := httphandler.NewLoginHandler(&fakeLoginPerformer{}, nil, nil)

	rec := httptest.NewRecorder()
	body := buildLoginBody("a@b.com", "")
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestLogin_MalformedJSON(t *testing.T) {
	handler := httphandler.NewLoginHandler(&fakeLoginPerformer{}, nil, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	handler.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestLogin_WrongMethod(t *testing.T) {
	handler := httphandler.NewLoginHandler(&fakeLoginPerformer{}, nil, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	handler.Login(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	allow := rec.Header().Get("Allow")
	if allow != http.MethodPost {
		t.Errorf("expected Allow: POST, got %q", allow)
	}
}

func TestLogin_InternalError(t *testing.T) {
	performer := &fakeLoginPerformer{err: context.DeadlineExceeded}
	handler := httphandler.NewLoginHandler(performer, nil, nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(buildLoginBody("a@b.com", "pass")))
	req.Header.Set("Content-Type", "application/json")
	handler.Login(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestCookieSessionHandler_Write(t *testing.T) {
	cw := httphandler.NewCookieSessionHandler("donarium_session", "/", true, 1*time.Hour)
	rec := httptest.NewRecorder()
	cw.Write(rec, "test-token")

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	c := cookies[0]
	if c.Name != "donarium_session" {
		t.Errorf("name: %q", c.Name)
	}
	if c.Value != "test-token" {
		t.Errorf("value: %q", c.Value)
	}
	if c.Path != "/" {
		t.Errorf("path: %q", c.Path)
	}
	if !c.HttpOnly {
		t.Error("expected HttpOnly")
	}
	if !c.Secure {
		t.Error("expected Secure=true")
	}
	if c.SameSite != http.SameSiteLaxMode {
		t.Errorf("sameSite: %v", c.SameSite)
	}
	if c.MaxAge != 3600 {
		t.Errorf("maxAge: %d", c.MaxAge)
	}
	if c.Expires.IsZero() {
		t.Error("expected Expires to be set")
	}
}

func TestCookieSessionHandler_WriteInsecure(t *testing.T) {
	cw := httphandler.NewCookieSessionHandler("donarium_session", "/", false, 1*time.Hour)
	rec := httptest.NewRecorder()
	cw.Write(rec, "test-token")

	c := rec.Result().Cookies()[0]
	if c.Secure {
		t.Error("expected Secure=false")
	}
	if !c.HttpOnly {
		t.Error("expected HttpOnly")
	}
	if c.SameSite != http.SameSiteLaxMode {
		t.Errorf("sameSite: %v", c.SameSite)
	}
}

func TestCookieSessionHandler_Clear(t *testing.T) {
	cw := httphandler.NewCookieSessionHandler("donarium_session", "/", true, 1*time.Hour)
	rec := httptest.NewRecorder()
	cw.Clear(rec)

	c := rec.Result().Cookies()[0]
	if c.Value != "" {
		t.Errorf("expected empty value, got %q", c.Value)
	}
	if c.MaxAge != -1 {
		t.Errorf("expected MaxAge=-1, got %d", c.MaxAge)
	}
}
