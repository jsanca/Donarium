package authorization_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/application/authentication"
	"donarium/server/internal/identity/authorization"
	httphandler "donarium/server/internal/identity/http"
)

func TestHasPlatformRole_Positive(t *testing.T) {
	p := authentication.AuthenticatedPrincipal{PlatformRoles: []string{"super_admin"}}
	if !authorization.HasPlatformRole(p, identity.PlatformRoleSuperAdmin) {
		t.Error("expected HasPlatformRole=true")
	}
}

func TestHasPlatformRole_Negative(t *testing.T) {
	p := authentication.AuthenticatedPrincipal{PlatformRoles: []string{"super_admin"}}
	if authorization.HasPlatformRole(p, "other") {
		t.Error("expected HasPlatformRole=false")
	}
}

func TestHasPlatformRole_EmptyRoles(t *testing.T) {
	p := authentication.AuthenticatedPrincipal{PlatformRoles: []string{}}
	if authorization.HasPlatformRole(p, identity.PlatformRoleSuperAdmin) {
		t.Error("expected false for empty roles")
	}
}

func TestHasOrganizationRole_Positive(t *testing.T) {
	p := authentication.AuthenticatedPrincipal{
		OrganizationContexts: []authentication.OrganizationContext{{Role: "owner"}},
	}
	if !authorization.HasOrganizationRole(p, identity.OrganizationRoleOwner) {
		t.Error("expected true")
	}
}

func TestHasOrganizationRole_Negative(t *testing.T) {
	p := authentication.AuthenticatedPrincipal{
		OrganizationContexts: []authentication.OrganizationContext{{Role: "owner"}},
	}
	if authorization.HasOrganizationRole(p, "tenant") {
		t.Error("expected false")
	}
}

func TestHasOrganizationRole_MultipleContexts(t *testing.T) {
	p := authentication.AuthenticatedPrincipal{
		OrganizationContexts: []authentication.OrganizationContext{
			{Role: "owner"},
			{Role: "tenant"},
		},
	}
	if !authorization.HasOrganizationRole(p, identity.OrganizationRoleOwner) {
		t.Error("expected true for owner in multiple contexts")
	}
	if !authorization.HasOrganizationRole(p, "tenant") {
		t.Error("expected true for tenant in multiple contexts")
	}
}

func TestRequirePlatformRole_Matching(t *testing.T) {
	principal := authentication.AuthenticatedPrincipal{
		UserID:        "u",
		PlatformRoles: []string{"super_admin"},
	}
	ctx := httphandler.WithPrincipal(context.Background(), principal)
	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	called := false
	handler := authorization.RequirePlatformRole(identity.PlatformRoleSuperAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called")
	}
}

func TestRequirePlatformRole_MissingRole(t *testing.T) {
	principal := authentication.AuthenticatedPrincipal{
		UserID:        "u",
		PlatformRoles: []string{"super_admin"},
	}
	ctx := httphandler.WithPrincipal(context.Background(), principal)
	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	called := false
	handler := authorization.RequirePlatformRole("other_role")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
	if called {
		t.Error("handler should not be called")
	}
}

func TestRequirePlatformRole_NoPrincipal(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	called := false
	handler := authorization.RequirePlatformRole(identity.PlatformRoleSuperAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestRequireOrganizationRole_Matching(t *testing.T) {
	principal := authentication.AuthenticatedPrincipal{
		UserID: "u",
		OrganizationContexts: []authentication.OrganizationContext{{Role: "owner"}},
	}
	ctx := httphandler.WithPrincipal(context.Background(), principal)
	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	called := false
	handler := authorization.RequireOrganizationRole(identity.OrganizationRoleOwner)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called")
	}
}

func TestRequireOrganizationRole_MissingRole(t *testing.T) {
	principal := authentication.AuthenticatedPrincipal{
		UserID: "u",
		OrganizationContexts: []authentication.OrganizationContext{{Role: "owner"}},
	}
	ctx := httphandler.WithPrincipal(context.Background(), principal)
	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	called := false
	handler := authorization.RequireOrganizationRole("tenant")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
	if called {
		t.Error("handler should not be called")
	}
}

func TestRequireOrganizationRole_NoPrincipal(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	called := false
	handler := authorization.RequireOrganizationRole(identity.OrganizationRoleOwner)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestMiddlewareComposition(t *testing.T) {
	principal := authentication.AuthenticatedPrincipal{
		UserID:               "u",
		PlatformRoles:        []string{"super_admin"},
		OrganizationContexts: []authentication.OrganizationContext{{Role: "owner"}},
	}
	ctx := httphandler.WithPrincipal(context.Background(), principal)
	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	called := false
	handler := httphandler.RequireAuthenticated()(
		authorization.RequirePlatformRole(identity.PlatformRoleSuperAdmin)(
			authorization.RequireOrganizationRole(identity.OrganizationRoleOwner)(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					called = true
				}),
			),
		),
	)
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected all middleware to pass")
	}
}

func TestMiddlewareComposition_SecondFails(t *testing.T) {
	principal := authentication.AuthenticatedPrincipal{
		UserID:        "u",
		PlatformRoles: []string{},
	}
	ctx := httphandler.WithPrincipal(context.Background(), principal)
	req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	called := false
	handler := httphandler.RequireAuthenticated()(
		authorization.RequirePlatformRole(identity.PlatformRoleSuperAdmin)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
			}),
		),
	)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
	if called {
		t.Error("handler should not be called")
	}
}
