package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	identityhttp "donarium/server/internal/identity/http"
	"donarium/server/internal/properties"
	"donarium/server/internal/properties/application"
	propertiespgx "donarium/server/internal/properties/pgx"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *application.Service
	pool    *pgxpool.Pool
}

func NewHandler(service *application.Service, pool *pgxpool.Pool) *Handler {
	return &Handler{service: service, pool: pool}
}

func (h *Handler) Collection(w nethttp.ResponseWriter, r *nethttp.Request) {
	switch r.Method {
	case nethttp.MethodGet:
		h.ListProperties(w, r)
	case nethttp.MethodPost:
		h.RegisterProperty(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, nethttp.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) RegisterProperty(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodPost {
		w.Header().Set("Allow", nethttp.MethodPost)
		writeError(w, nethttp.StatusMethodNotAllowed, "method not allowed")
		return
	}

	principal, ok := identityhttp.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, nethttp.StatusUnauthorized, "authentication required")
		return
	}

	var req RegisterPropertyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, nethttp.StatusBadRequest, "invalid request body")
		return
	}

	stakeholders := make([]application.StakeholderInput, 0, len(req.Stakeholders))
	for _, s := range req.Stakeholders {
		stakeholders = append(stakeholders, application.StakeholderInput{
			Party: application.PartyInput{
				Type:           s.Party.Type,
				UserID:         s.Party.UserID,
				OrganizationID: s.Party.OrganizationID,
				ExternalName:   s.Party.ExternalName,
				ExternalEmail:  s.Party.ExternalEmail,
			},
			Role: s.Role,
		})
	}

	cmd := application.RegisterPropertyCommand{
		DisplayName:    req.DisplayName,
		Classification: req.Classification,
		Address: properties.Address{
			Street:     req.Address.Street,
			City:       req.Address.City,
			State:      req.Address.State,
			PostalCode: req.Address.PostalCode,
			Country:    req.Address.Country,
		},
		RentalCadence: req.RentalCadence,
		StandardRent:  req.StandardRent,
		Stakeholders:  stakeholders,
	}

	result, err := h.service.RegisterProperty(r.Context(), cmd, principal.UserID)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusCreated)
	_ = json.NewEncoder(w).Encode(toPropertyResponse(result.Property))
}

func (h *Handler) ListProperties(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodGet {
		w.Header().Set("Allow", nethttp.MethodGet)
		writeError(w, nethttp.StatusMethodNotAllowed, "method not allowed")
		return
	}

	principal, ok := identityhttp.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, nethttp.StatusUnauthorized, "authentication required")
		return
	}

	db := propertiespgx.NewExecutorFromPool(h.pool)
	props, err := h.service.ListAccessible(r.Context(), db, principal.UserID)
	if err != nil {
		writeError(w, nethttp.StatusInternalServerError, "internal server error")
		return
	}

	resp := ListPropertiesResponse{
		Properties: make([]PropertyResponse, 0, len(props)),
	}
	for _, p := range props {
		resp.Properties = append(resp.Properties, toPropertyResponse(p))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetProperty(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodGet {
		w.Header().Set("Allow", nethttp.MethodGet)
		writeError(w, nethttp.StatusMethodNotAllowed, "method not allowed")
		return
	}

	principal, ok := identityhttp.PrincipalFromContext(r.Context())
	if !ok {
		writeError(w, nethttp.StatusUnauthorized, "authentication required")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, nethttp.StatusBadRequest, "property id is required")
		return
	}

	db := propertiespgx.NewExecutorFromPool(h.pool)
	prop, err := h.service.GetByID(r.Context(), db, id, principal.UserID)
	if err != nil {
		if errors.Is(err, properties.ErrPropertyNotFound) {
			writeError(w, nethttp.StatusNotFound, "property not found")
			return
		}
		writeError(w, nethttp.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toPropertyResponse(prop))
}

func mapDomainError(w nethttp.ResponseWriter, err error) {
	switch {
	case errors.Is(err, properties.ErrInvalidDisplayName):
		writeError(w, nethttp.StatusBadRequest, "displayName is not valid")
	case errors.Is(err, properties.ErrInvalidClassification):
		writeError(w, nethttp.StatusBadRequest, "classification is not valid")
	case errors.Is(err, properties.ErrInvalidAddress):
		writeError(w, nethttp.StatusBadRequest, "address is not valid")
	case errors.Is(err, properties.ErrInvalidRentalCadence):
		writeError(w, nethttp.StatusBadRequest, "rental cadence is not valid")
	case errors.Is(err, properties.ErrInvalidStandardRent):
		writeError(w, nethttp.StatusBadRequest, "standard rent is not valid")
	case errors.Is(err, properties.ErrInvalidParty):
		writeError(w, nethttp.StatusBadRequest, "party is not valid")
	case errors.Is(err, properties.ErrInvalidStakeholderRole):
		writeError(w, nethttp.StatusBadRequest, "stakeholder role is not valid")
	case errors.Is(err, properties.ErrInvalidStakeholder):
		writeError(w, nethttp.StatusBadRequest, "stakeholder is not valid")
	case errors.Is(err, properties.ErrDuplicateStakeholder):
		writeError(w, nethttp.StatusConflict, "stakeholder already exists")
	case errors.Is(err, properties.ErrNoStakeholder):
		writeError(w, nethttp.StatusBadRequest, "at least one stakeholder is required")
	default:
		writeError(w, nethttp.StatusInternalServerError, "internal server error")
	}
}

func writeError(w nethttp.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
