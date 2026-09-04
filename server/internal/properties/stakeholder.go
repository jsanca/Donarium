package properties

import (
	"strings"
	"time"
)

type StakeholderRole string

const (
	StakeholderRoleOwner   StakeholderRole = "owner"
	StakeholderRoleManager StakeholderRole = "manager"
)

func ParseStakeholderRole(raw string) (StakeholderRole, error) {
	key := strings.ToLower(strings.TrimSpace(raw))
	switch key {
	case "owner":
		return StakeholderRoleOwner, nil
	case "manager":
		return StakeholderRoleManager, nil
	default:
		return "", ErrInvalidStakeholderRole
	}
}

type PropertyStakeholder struct {
	PropertyID PropertyID      `json:"propertyId"`
	Party      Party           `json:"party"`
	Role       StakeholderRole `json:"role"`
	CreatedAt  time.Time       `json:"createdAt"`
}

func NewPropertyStakeholder(propertyID PropertyID, party Party, role StakeholderRole) (PropertyStakeholder, error) {
	if propertyID.IsZero() {
		return PropertyStakeholder{}, ErrInvalidStakeholder
	}
	if err := party.Validate(); err != nil {
		return PropertyStakeholder{}, err
	}
	if role != StakeholderRoleOwner && role != StakeholderRoleManager {
		return PropertyStakeholder{}, ErrInvalidStakeholderRole
	}
	return PropertyStakeholder{
		PropertyID: propertyID,
		Party:      party,
		Role:       role,
		CreatedAt:  time.Now().UTC(),
	}, nil
}
