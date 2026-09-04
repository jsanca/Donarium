package http

import "donarium/server/internal/properties"

type AddressRequest struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
}

type PartyRequest struct {
	Type           string `json:"type"`
	UserID         string `json:"userId,omitempty"`
	OrganizationID string `json:"organizationId,omitempty"`
	ExternalName   string `json:"externalName,omitempty"`
	ExternalEmail  string `json:"externalEmail,omitempty"`
}

type StakeholderRequest struct {
	Party PartyRequest `json:"party"`
	Role  string       `json:"role"`
}

type RegisterPropertyRequest struct {
	DisplayName    string               `json:"displayName"`
	Classification string               `json:"classification"`
	Address        AddressRequest       `json:"address"`
	RentalCadence  string               `json:"rentalCadence"`
	StandardRent   int64                `json:"standardRent"`
	Stakeholders   []StakeholderRequest `json:"stakeholders,omitempty"`
}

type PropertyResponse struct {
	ID             string            `json:"id"`
	DisplayName    string            `json:"displayName"`
	Classification string            `json:"classification"`
	Address        properties.Address `json:"address"`
	RentalCadence  string            `json:"rentalCadence"`
	StandardRent   int64             `json:"standardRent"`
	CreatedBy      string            `json:"createdBy"`
	CreatedAt      string            `json:"createdAt"`
	UpdatedAt      string            `json:"updatedAt"`
}

type ListPropertiesResponse struct {
	Properties []PropertyResponse `json:"properties"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func toPropertyResponse(p properties.Property) PropertyResponse {
	return PropertyResponse{
		ID:             p.ID.String(),
		DisplayName:    p.DisplayName,
		Classification: string(p.Classification),
		Address:        p.Address,
		RentalCadence:  string(p.RentalCadence),
		StandardRent:   p.StandardRent,
		CreatedBy:      p.CreatedBy.String(),
		CreatedAt:      p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
