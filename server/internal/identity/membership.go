package identity

import (
	"time"
)

type Membership struct {
	UserID         UserID
	OrganizationID OrganizationID
	Role           OrganizationRole
	CreatedAt      time.Time
}

func NewMembership(userID UserID, organizationID OrganizationID, role OrganizationRole) (Membership, error) {
	if userID.IsZero() {
		return Membership{}, ErrEmptyUserID
	}
	if organizationID.IsZero() {
		return Membership{}, ErrEmptyOrganizationID
	}
	if !role.Valid() {
		return Membership{}, ErrInvalidOrganizationRole
	}

	return Membership{
		UserID:         userID,
		OrganizationID: organizationID,
		Role:           role,
		CreatedAt:      time.Now().UTC(),
	}, nil
}
