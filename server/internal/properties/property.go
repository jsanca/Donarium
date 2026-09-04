package properties

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type PropertyID uuid.UUID

func NewPropertyID() PropertyID {
	return PropertyID(uuid.New())
}

func (id PropertyID) String() string {
	return uuid.UUID(id).String()
}

func (id PropertyID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

type Classification string

const (
	ClassificationHouse       Classification = "house"
	ClassificationApartment   Classification = "apartment"
	ClassificationCondominium Classification = "condominium"
	ClassificationMultiUnit   Classification = "multi_unit"
	ClassificationCommercial  Classification = "commercial"
	ClassificationOther       Classification = "other"
)

var validClassifications = map[string]Classification{
	"house":       ClassificationHouse,
	"apartment":   ClassificationApartment,
	"condominium": ClassificationCondominium,
	"multi_unit":  ClassificationMultiUnit,
	"multi-unit":  ClassificationMultiUnit,
	"multiunit":   ClassificationMultiUnit,
	"commercial":  ClassificationCommercial,
	"other":       ClassificationOther,
}

func ParseClassification(raw string) (Classification, error) {
	key := strings.ToLower(strings.TrimSpace(raw))
	key = strings.ReplaceAll(key, "-", "_")
	key = strings.ReplaceAll(key, " ", "_")
	if c, ok := validClassifications[key]; ok {
		return c, nil
	}
	return "", ErrInvalidClassification
}

type RentalCadence string

const (
	CadenceMonthly RentalCadence = "monthly"
	CadenceWeekly  RentalCadence = "weekly"
	CadenceDaily   RentalCadence = "daily"
	CadenceAnnual  RentalCadence = "annual"
)

var validCadences = map[string]RentalCadence{
	"monthly": CadenceMonthly,
	"weekly":  CadenceWeekly,
	"daily":   CadenceDaily,
	"annual":  CadenceAnnual,
}

func ParseRentalCadence(raw string) (RentalCadence, error) {
	key := strings.ToLower(strings.TrimSpace(raw))
	if c, ok := validCadences[key]; ok {
		return c, nil
	}
	return "", ErrInvalidRentalCadence
}

type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
}

func (a Address) Validate() error {
	if strings.TrimSpace(a.Street) == "" {
		return ErrInvalidAddress
	}
	if strings.TrimSpace(a.City) == "" {
		return ErrInvalidAddress
	}
	if strings.TrimSpace(a.PostalCode) == "" {
		return ErrInvalidAddress
	}
	if strings.TrimSpace(a.Country) == "" {
		return ErrInvalidAddress
	}
	return nil
}

type Property struct {
	ID             PropertyID    `json:"id"`
	DisplayName    string        `json:"displayName"`
	Classification Classification `json:"classification"`
	Address        Address       `json:"address"`
	RentalCadence  RentalCadence `json:"rentalCadence"`
	StandardRent   int64         `json:"standardRent"`
	CreatedBy      uuid.UUID     `json:"createdBy"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

func ValidateDisplayName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ErrInvalidDisplayName
	}
	if len(trimmed) < 2 || len(trimmed) > 100 {
		return ErrInvalidDisplayName
	}
	return nil
}

func ValidateStandardRent(rent int64) error {
	if rent <= 0 {
		return ErrInvalidStandardRent
	}
	if rent > 1_000_000_000_00 { // 10M in cents, generous upper bound
		return ErrInvalidStandardRent
	}
	return nil
}
