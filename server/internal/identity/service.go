package identity

import "context"

type SetupRequest struct {
	OwnerEmail       string
	OwnerPassword    string
	OrganizationName string
	DisplayName      string
}

type SetupResult struct {
	UserID         UserID
	OrganizationID OrganizationID
}

type SetupService interface {
	Setup(ctx context.Context, req SetupRequest) (*SetupResult, error)
}
