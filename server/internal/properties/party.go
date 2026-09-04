package properties

import (
	"net/mail"
	"strings"

	"github.com/google/uuid"
)

type PartyType string

const (
	PartyTypeUser         PartyType = "user"
	PartyTypeOrganization PartyType = "organization"
	PartyTypeExternal     PartyType = "external"
)

type Party struct {
	Type             PartyType `json:"type"`
	UserID           *uuid.UUID `json:"userId,omitempty"`
	OrganizationID   *uuid.UUID `json:"organizationId,omitempty"`
	ExternalName     string    `json:"externalName,omitempty"`
	ExternalEmail    string    `json:"externalEmail,omitempty"`
}

func NewUserParty(userID uuid.UUID) Party {
	uid := userID
	return Party{Type: PartyTypeUser, UserID: &uid}
}

func NewOrganizationParty(orgID uuid.UUID) Party {
	oid := orgID
	return Party{Type: PartyTypeOrganization, OrganizationID: &oid}
}

func NewExternalParty(name, email string) Party {
	return Party{Type: PartyTypeExternal, ExternalName: strings.TrimSpace(name), ExternalEmail: strings.TrimSpace(email)}
}

func (p Party) Validate() error {
	switch p.Type {
	case PartyTypeUser:
		if p.UserID == nil || *p.UserID == uuid.Nil {
			return ErrInvalidParty
		}
		if p.OrganizationID != nil || p.ExternalName != "" || p.ExternalEmail != "" {
			return ErrInvalidParty
		}
	case PartyTypeOrganization:
		if p.OrganizationID == nil || *p.OrganizationID == uuid.Nil {
			return ErrInvalidParty
		}
		if p.UserID != nil || p.ExternalName != "" || p.ExternalEmail != "" {
			return ErrInvalidParty
		}
	case PartyTypeExternal:
		if strings.TrimSpace(p.ExternalName) == "" {
			return ErrInvalidParty
		}
		if len(p.ExternalName) < 2 || len(p.ExternalName) > 100 {
			return ErrInvalidParty
		}
		email := strings.TrimSpace(strings.ToLower(p.ExternalEmail))
		if email == "" {
			return ErrInvalidParty
		}
		if _, err := mail.ParseAddress(email); err != nil {
			return ErrInvalidParty
		}
		if !strings.Contains(email, "@") {
			return ErrInvalidParty
		}
		if p.UserID != nil || p.OrganizationID != nil {
			return ErrInvalidParty
		}
	default:
		return ErrInvalidParty
	}
	return nil
}

func (p Party) NormalizedExternalEmail() string {
	if p.Type != PartyTypeExternal {
		return ""
	}
	return strings.TrimSpace(strings.ToLower(p.ExternalEmail))
}

func (p Party) ReferenceKey() string {
	switch p.Type {
	case PartyTypeUser:
		return p.UserID.String()
	case PartyTypeOrganization:
		return p.OrganizationID.String()
	case PartyTypeExternal:
		return strings.ToLower(strings.TrimSpace(p.ExternalEmail))
	default:
		return ""
	}
}
