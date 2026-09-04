package properties

import "errors"

var (
	ErrPropertyNotFound       = errors.New("property not found")
	ErrInvalidClassification  = errors.New("classification is not valid")
	ErrInvalidRentalCadence   = errors.New("rental cadence is not valid")
	ErrInvalidDisplayName     = errors.New("display name is not valid")
	ErrInvalidAddress         = errors.New("address is not valid")
	ErrInvalidStandardRent    = errors.New("standard rent is not valid")
	ErrUnauthorized           = errors.New("not authorized to access property")
	ErrInvalidParty           = errors.New("party is not valid")
	ErrInvalidStakeholderRole = errors.New("stakeholder role is not valid")
	ErrInvalidStakeholder     = errors.New("stakeholder is not valid")
	ErrDuplicateStakeholder   = errors.New("stakeholder already exists")
	ErrNoStakeholder          = errors.New("at least one stakeholder is required")
)
