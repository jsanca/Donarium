package identity

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

type OrganizationID uuid.UUID

func NewOrganizationID() OrganizationID {
	return OrganizationID(uuid.New())
}

func (id OrganizationID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

type Organization struct {
	ID        OrganizationID
	Name      string
	Slug      string
	CreatedAt time.Time
	CreatedBy UserID
}

func NewOrganization(name, slug string, createdBy UserID) (Organization, error) {
	if name == "" {
		return Organization{}, ErrEmptyName
	}
	if slug == "" {
		return Organization{}, ErrEmptySlug
	}
	if !slugPattern.MatchString(slug) {
		return Organization{}, ErrInvalidSlug
	}
	if createdBy.IsZero() {
		return Organization{}, ErrEmptyUserID
	}

	return Organization{
		ID:        NewOrganizationID(),
		Name:      name,
		Slug:      slug,
		CreatedAt: time.Now().UTC(),
		CreatedBy: createdBy,
	}, nil
}
