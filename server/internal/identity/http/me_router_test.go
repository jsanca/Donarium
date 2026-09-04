package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/application/authentication"
	httphandler "donarium/server/internal/identity/http"

	"github.com/go-chi/chi/v5"
)

type routerVerifier struct {
	result authentication.SessionClaims
	err    error
}

func (v *routerVerifier) Verify(token string) (authentication.SessionClaims, error) {
	return v.result, v.err
}

type routerResolver struct {
	result authentication.AuthenticatedPrincipal
	err    error
}

func (r *routerResolver) Resolve(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (authentication.AuthenticatedPrincipal, error) {
	return r.result, r.err
}

type routerCookieReader struct{}

func (w *routerCookieReader) Read(r *http.Request) (string, error) {
	c, err := r.Cookie("donarium_session")
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

func buildMeRouter(verifier authentication.SessionVerifier, resolver authentication.PrincipalResolver) chi.Router {
	cookieReader := &routerCookieReader{}
	authMW := httphandler.NewAuthenticationMiddleware(verifier, resolver, nil, cookieReader.Read)

	r := chi.NewRouter()
	r.Route("/api/auth/me", func(protected chi.Router) {
		protected.Use(authMW.RequireAuthentication)
		protected.Use(httphandler.RequireAuthenticated())
		protected.HandleFunc("/", httphandler.NewMeHandler().Me)
	})
	return r
}

func reqMeWithCookie(token string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "donarium_session", Value: token})
	}
	return req
}

func TestMeRouter_Unauthenticated(t *testing.T) {
	verifier := &routerVerifier{err: identity.ErrInvalidSession}
	resolver := &routerResolver{}
	router := buildMeRouter(verifier, resolver)

	rec := httptest.NewRecorder()
	req := reqMeWithCookie("bad.token")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}

	var errResp httphandler.ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error != "authentication required" {
		t.Errorf("expected 'authentication required', got %q", errResp.Error)
	}
}

func TestMeRouter_MissingCookieReturns401(t *testing.T) {
	verifier := &routerVerifier{}
	resolver := &routerResolver{}
	router := buildMeRouter(verifier, resolver)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMeRouter_ExpiredSessionReturns401(t *testing.T) {
	verifier := &routerVerifier{err: identity.ErrExpiredSession}
	resolver := &routerResolver{}
	router := buildMeRouter(verifier, resolver)

	rec := httptest.NewRecorder()
	req := reqMeWithCookie("expired.token")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMeRouter_AuthenticatedReturns200(t *testing.T) {
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

	verifier := &routerVerifier{result: authentication.SessionClaims{Subject: "00000000-0000-0000-0000-000000000001"}}
	resolver := &routerResolver{result: principal}
	router := buildMeRouter(verifier, resolver)

	rec := httptest.NewRecorder()
	req := reqMeWithCookie("valid.token")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json content type, got %q", ct)
	}

	var resp httphandler.LoginResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Principal.Email != "owner@test.com" {
		t.Errorf("email mismatch: %q", resp.Principal.Email)
	}
	if resp.Principal.UserID != "user-uuid" {
		t.Errorf("userId mismatch: %q", resp.Principal.UserID)
	}
	if len(resp.Principal.OrganizationContexts) != 1 {
		t.Errorf("expected 1 organization context, got %d", len(resp.Principal.OrganizationContexts))
	}
}

func TestMeRouter_SessionTokenNotInResponse(t *testing.T) {
	principal := authentication.AuthenticatedPrincipal{
		UserID:        "user-uuid",
		Email:         "owner@test.com",
		PlatformRoles: []string{"super_admin"},
	}

	verifier := &routerVerifier{result: authentication.SessionClaims{Subject: "00000000-0000-0000-0000-000000000001"}}
	resolver := &routerResolver{result: principal}
	router := buildMeRouter(verifier, resolver)

	rec := httptest.NewRecorder()
	req := reqMeWithCookie("valid.token")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if strings.Contains(body, "sessionToken") || strings.Contains(body, `"sessionToken"`) {
		t.Error("session token should not be serialized in response")
	}
}

func TestMeRouter_WrongMethodReturns405(t *testing.T) {
	principal := authentication.AuthenticatedPrincipal{
		UserID: "u",
		Email:  "test@donarium.test",
	}

	verifier := &routerVerifier{result: authentication.SessionClaims{Subject: "00000000-0000-0000-0000-000000000001"}}
	resolver := &routerResolver{result: principal}
	router := buildMeRouter(verifier, resolver)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/me", nil)
	req.AddCookie(&http.Cookie{Name: "donarium_session", Value: "valid.token"})
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestMeRouter_ContextPassesThroughMiddleware(t *testing.T) {
	principal := authentication.AuthenticatedPrincipal{
		UserID:      "u",
		DisplayName: "Test",
		Email:       "test@donarium.test",
	}

	verifier := &routerVerifier{result: authentication.SessionClaims{Subject: "00000000-0000-0000-0000-000000000001"}}
	resolver := &routerResolver{result: principal}
	router := buildMeRouter(verifier, resolver)

	rec := httptest.NewRecorder()
	req := reqMeWithCookie("valid.token")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp httphandler.LoginResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Principal.DisplayName != "Test" {
		t.Errorf("expected DisplayName 'Test', got %q", resp.Principal.DisplayName)
	}
}

func TestMeRouter_PublicPathIsRegistered(t *testing.T) {
	verifier := &routerVerifier{}
	resolver := &routerResolver{}
	router := buildMeRouter(verifier, resolver)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/auth/me", nil)
	router.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Error("expected /api/auth/me route to be registered")
	}
}
