package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/application/authentication"
	httphandler "donarium/server/internal/identity/http"
)

type fakeVerifier struct {
	result authentication.SessionClaims
	err    error
}

func (v *fakeVerifier) Verify(token string) (authentication.SessionClaims, error) {
	return v.result, v.err
}

type fakeResolver struct {
	result authentication.AuthenticatedPrincipal
	err    error
}

func (r *fakeResolver) Resolve(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (authentication.AuthenticatedPrincipal, error) {
	return r.result, r.err
}

func readCookie(rec *httptest.ResponseRecorder) httphandler.SessionCookieReadFn {
	return func(r *http.Request) (string, error) {
		c, err := r.Cookie("donarium_session")
		if err != nil {
			return "", err
		}
		return c.Value, nil
	}
}

func reqWithCookie(token string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "donarium_session", Value: token})
	}
	return req
}

func TestMiddleware_MissingCookie(t *testing.T) {
	mw := httphandler.NewAuthenticationMiddleware(
		&fakeVerifier{}, &fakeResolver{}, nil,
		func(r *http.Request) (string, error) { return "", errors.New("no cookie") },
	)

	called := false
	handler := mw.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, reqWithCookie(""))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	if called {
		t.Error("handler should not be called")
	}
}

func TestMiddleware_EmptyCookie(t *testing.T) {
	mw := httphandler.NewAuthenticationMiddleware(
		&fakeVerifier{}, &fakeResolver{}, nil,
		func(r *http.Request) (string, error) { return "", nil },
	)

	called := false
	handler := mw.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, reqWithCookie(""))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	if called {
		t.Error("handler should not be called")
	}
}

func TestMiddleware_InvalidToken(t *testing.T) {
	mw := httphandler.NewAuthenticationMiddleware(
		&fakeVerifier{err: identity.ErrInvalidSession},
		&fakeResolver{}, nil,
		func(r *http.Request) (string, error) { return "bad.token", nil },
	)

	called := false
	handler := mw.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, reqWithCookie("bad.token"))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	if called {
		t.Error("handler should not be called")
	}
}

func TestMiddleware_ExpiredToken(t *testing.T) {
	mw := httphandler.NewAuthenticationMiddleware(
		&fakeVerifier{err: identity.ErrExpiredSession},
		&fakeResolver{}, nil,
		func(r *http.Request) (string, error) { return "expired.token", nil },
	)

	called := false
	handler := mw.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, reqWithCookie("expired.token"))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	if called {
		t.Error("handler should not be called")
	}
}

func TestMiddleware_DeletedUser(t *testing.T) {
	mw := httphandler.NewAuthenticationMiddleware(
		&fakeVerifier{result: authentication.SessionClaims{Subject: "00000000-0000-0000-0000-000000000001"}},
		&fakeResolver{err: identity.ErrInvalidCredentials},
		nil,
		func(r *http.Request) (string, error) { return "valid.token", nil },
	)

	called := false
	handler := mw.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, reqWithCookie("valid.token"))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	if called {
		t.Error("handler should not be called")
	}
}

func TestMiddleware_RepositoryFailure(t *testing.T) {
	mw := httphandler.NewAuthenticationMiddleware(
		&fakeVerifier{result: authentication.SessionClaims{Subject: "00000000-0000-0000-0000-000000000001"}},
		&fakeResolver{err: errors.New("db connection refused")},
		nil,
		func(r *http.Request) (string, error) { return "valid.token", nil },
	)

	called := false
	handler := mw.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, reqWithCookie("valid.token"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	if called {
		t.Error("handler should not be called")
	}
}

func TestMiddleware_SuccessInvokesNext(t *testing.T) {
	principal := authentication.AuthenticatedPrincipal{
		UserID:    "user-uuid",
		Email:     "owner@test.com",
		PlatformRoles: []string{"super_admin"},
	}
	mw := httphandler.NewAuthenticationMiddleware(
		&fakeVerifier{result: authentication.SessionClaims{Subject: "00000000-0000-0000-0000-000000000001"}},
		&fakeResolver{result: principal},
		nil,
		func(r *http.Request) (string, error) { return "valid.token", nil },
	)

	called := false
	handler := mw.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		p, ok := httphandler.PrincipalFromContext(r.Context())
		if !ok {
			t.Error("principal not found in context")
		}
		if p.UserID != principal.UserID {
			t.Errorf("principal mismatch")
		}
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, reqWithCookie("valid.token"))

	if !called {
		t.Error("handler should have been called")
	}
}

func TestMiddleware_SuccessInvokesNextExactlyOnce(t *testing.T) {
	mw := httphandler.NewAuthenticationMiddleware(
		&fakeVerifier{result: authentication.SessionClaims{Subject: "00000000-0000-0000-0000-000000000001"}},
		&fakeResolver{result: authentication.AuthenticatedPrincipal{UserID: "u"}},
		nil,
		func(r *http.Request) (string, error) { return "valid.token", nil },
	)

	count := 0
	handler := mw.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
	}))

	handler.ServeHTTP(httptest.NewRecorder(), reqWithCookie("valid.token"))
	if count != 1 {
		t.Errorf("expected 1 invocation, got %d", count)
	}
}

func TestMeHandler_Success(t *testing.T) {
	principal := authentication.AuthenticatedPrincipal{
		UserID:        "user-uuid",
		DisplayName:   "Owner",
		Email:         "owner@test.com",
		PlatformRoles: []string{"super_admin"},
		OrganizationContexts: []authentication.OrganizationContext{
			{OrganizationID: "org-uuid", OrganizationName: "Org", OrganizationSlug: "org", Role: "owner"},
		},
		DefaultContext: authentication.DefaultContext{Type: "organization", OrganizationID: "org-uuid", Role: "owner"},
	}

	ctx := httphandler.WithPrincipal(context.Background(), principal)
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	handler := httphandler.NewMeHandler()
	handler.Me(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp httphandler.LoginResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Principal.Email != "owner@test.com" {
		t.Errorf("email mismatch: %q", resp.Principal.Email)
	}

	body := rec.Body.String()
	if strings.Contains(body, "sessionToken") || strings.Contains(body, `"sessionToken"`) {
		t.Error("session token should not be serialized")
	}
}

func TestMeHandler_NoPrincipal(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()

	handler := httphandler.NewMeHandler()
	handler.Me(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMeHandler_WrongMethod(t *testing.T) {
	principal := authentication.AuthenticatedPrincipal{UserID: "u"}
	ctx := httphandler.WithPrincipal(context.Background(), principal)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/me", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	handler := httphandler.NewMeHandler()
	handler.Me(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestMeHandler_ContentType(t *testing.T) {
	principal := authentication.AuthenticatedPrincipal{UserID: "u", PlatformRoles: []string{}}
	ctx := httphandler.WithPrincipal(context.Background(), principal)
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	handler := httphandler.NewMeHandler()
	handler.Me(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestRequireAuthenticated_PrincipalPresent(t *testing.T) {
	principal := authentication.AuthenticatedPrincipal{UserID: "u"}
	ctx := httphandler.WithPrincipal(context.Background(), principal)
	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	called := false
	handler := httphandler.RequireAuthenticated()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRequireAuthenticated_PrincipalMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	called := false
	handler := httphandler.RequireAuthenticated()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	if called {
		t.Error("handler should not be called")
	}
}

func TestRequireAuthenticated_NextCalledExactlyOnce(t *testing.T) {
	principal := authentication.AuthenticatedPrincipal{UserID: "u"}
	ctx := httphandler.WithPrincipal(context.Background(), principal)
	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	count := 0
	handler := httphandler.RequireAuthenticated()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
	}))
	handler.ServeHTTP(rec, req)

	if count != 1 {
		t.Errorf("expected 1 call, got %d", count)
	}
}
