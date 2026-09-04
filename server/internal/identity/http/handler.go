package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/application"

	"github.com/google/uuid"
)

type SetupPerformer interface {
	Execute(ctx context.Context, cmd application.InitialOwnerSetupCommand) (application.InitialOwnerSetupResult, error)
}

type StatusReader interface {
	IsInitialized(ctx context.Context) (bool, error)
}

type SetupHandler struct {
	setup  SetupPerformer
	status StatusReader
}

func NewSetupHandler(setup SetupPerformer, status StatusReader) *SetupHandler {
	return &SetupHandler{setup: setup, status: status}
}

func (h *SetupHandler) Setup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "displayName is required")
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
	if req.OrganizationName == "" {
		writeError(w, http.StatusBadRequest, "organizationName is required")
		return
	}
	if req.OrganizationSlug == "" {
		writeError(w, http.StatusBadRequest, "organizationSlug is required")
		return
	}

	cmd := application.InitialOwnerSetupCommand{
		DisplayName:      req.DisplayName,
		Email:            req.Email,
		Password:         req.Password,
		OrganizationName: req.OrganizationName,
		OrganizationSlug: req.OrganizationSlug,
	}

	result, err := h.setup.Execute(r.Context(), cmd)
	if err != nil {
		slog.Error("setup failed", "error", err)
		mapError(w, err)
		return
	}

	resp := SetupResponse{
		UserID:         uuid.UUID(result.UserID).String(),
		OrganizationID: uuid.UUID(result.OrganizationID).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *SetupHandler) Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	initialized, err := h.status.IsInitialized(r.Context())
	if err != nil {
		slog.Error("status check failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(SetupStatusResponse{Initialized: initialized})
}

func mapError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, identity.ErrAlreadyInitialized):
		writeError(w, http.StatusConflict, "system is already initialized")
	case errors.Is(err, identity.ErrDuplicateEmail):
		writeError(w, http.StatusConflict, "email already exists")
	case errors.Is(err, identity.ErrDuplicateSlug):
		writeError(w, http.StatusConflict, "organization slug already exists")
	case errors.Is(err, identity.ErrInvalidEmail):
		writeError(w, http.StatusBadRequest, "email is not valid")
	case errors.Is(err, identity.ErrInvalidPassword):
		writeError(w, http.StatusBadRequest, "password does not meet requirements")
	case errors.Is(err, identity.ErrInvalidDisplayName):
		writeError(w, http.StatusBadRequest, "display name is not valid")
	case errors.Is(err, identity.ErrInvalidOrganizationName):
		writeError(w, http.StatusBadRequest, "organization name is not valid")
	case errors.Is(err, identity.ErrInvalidSlug):
		writeError(w, http.StatusBadRequest, "slug is not valid")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
