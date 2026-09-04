package http

import (
	"encoding/json"
	"net/http"
)

type MeHandler struct{}

func NewMeHandler() *MeHandler {
	return &MeHandler{}
}

func (h *MeHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	principal, ok := PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(LoginResponse{
		Principal: loginPrincipal{
			UserID:               principal.UserID,
			DisplayName:          principal.DisplayName,
			Email:                principal.Email,
			PlatformRoles:        principal.PlatformRoles,
			OrganizationContexts: principal.OrganizationContexts,
			DefaultContext:       principal.DefaultContext,
		},
	})
}
