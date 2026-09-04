package authorization

import (
	"encoding/json"
	"net/http"

	"donarium/server/internal/identity"
	identityhttp "donarium/server/internal/identity/http"
	"donarium/server/internal/identity/application/authentication"
)

func HasPlatformRole(principal authentication.AuthenticatedPrincipal, role identity.PlatformRole) bool {
	for _, r := range principal.PlatformRoles {
		if identity.PlatformRole(r) == role {
			return true
		}
	}
	return false
}

func HasOrganizationRole(principal authentication.AuthenticatedPrincipal, role identity.OrganizationRole) bool {
	for _, c := range principal.OrganizationContexts {
		if identity.OrganizationRole(c.Role) == role {
			return true
		}
	}
	return false
}

func RequirePlatformRole(role identity.PlatformRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := identityhttp.PrincipalFromContext(r.Context())
			if !ok {
				writeAuthzError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			if !HasPlatformRole(principal, role) {
				writeAuthzError(w, http.StatusForbidden, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireOrganizationRole(role identity.OrganizationRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := identityhttp.PrincipalFromContext(r.Context())
			if !ok {
				writeAuthzError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			if !HasOrganizationRole(principal, role) {
				writeAuthzError(w, http.StatusForbidden, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type authzErrorResponse struct {
	Error string `json:"error"`
}

func writeAuthzError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(authzErrorResponse{Error: message})
}
