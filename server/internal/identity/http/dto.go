package http

import "donarium/server/internal/identity/application/authentication"

type SetupRequest struct {
	DisplayName      string `json:"displayName"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	OrganizationName string `json:"organizationName"`
	OrganizationSlug string `json:"organizationSlug"`
}

type SetupResponse struct {
	UserID         string `json:"userId"`
	OrganizationID string `json:"organizationId"`
}

type SetupStatusResponse struct {
	Initialized bool `json:"initialized"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginPrincipal struct {
	UserID               string                                `json:"userId"`
	DisplayName          string                                `json:"displayName"`
	Email                string                                `json:"email"`
	PlatformRoles        []string                              `json:"platformRoles"`
	OrganizationContexts []authentication.OrganizationContext   `json:"organizationContexts"`
	DefaultContext       authentication.DefaultContext         `json:"defaultContext"`
}

type LoginResponse struct {
	Principal loginPrincipal `json:"principal"`
}
