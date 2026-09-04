package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/application/authentication"
	"donarium/server/internal/identity/pgx"
)

type LoginPerformer interface {
	Execute(ctx context.Context, db identity.DBExecutor, cmd authentication.AuthenticateUserCommand) (authentication.AuthenticatedPrincipal, error)
}

type LoginHandler struct {
	login LoginPerformer
	pool  *pgxpool.Pool
	cookieWriter SessionCookieWriter
}

func NewLoginHandler(login LoginPerformer, pool *pgxpool.Pool, cookieWriter SessionCookieWriter) *LoginHandler {
	return &LoginHandler{login: login, pool: pool, cookieWriter: cookieWriter}
}

func (h *LoginHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}
	if req.Password == "" {
		writeError(w, http.StatusBadRequest, "password is required")
		return
	}

	cmd := authentication.AuthenticateUserCommand{
		Email:    req.Email,
		Password: req.Password,
	}

	principal, err := h.login.Execute(r.Context(), pgx.NewExecutorFromPool(h.pool), cmd)
	if err != nil {
		slog.Error("login failed", "error", err)
		mapLoginError(w, err)
		return
	}

	h.cookieWriter.Write(w, principal.SessionToken)

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

func mapLoginError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, identity.ErrInvalidEmail):
		writeError(w, http.StatusBadRequest, "email is not valid")
	case errors.Is(err, identity.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "the email or password is incorrect")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
