package pgx

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"donarium/server/internal/properties"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type StakeholderRepository struct{}

func NewStakeholderRepository() *StakeholderRepository {
	return &StakeholderRepository{}
}

func (r *StakeholderRepository) Create(ctx context.Context, db properties.DBExecutor, s properties.PropertyStakeholder) error {
	if s.PropertyID.IsZero() {
		return fmt.Errorf("property id must not be zero")
	}
	if err := s.Party.Validate(); err != nil {
		return err
	}

	var userID, orgID *uuid.UUID
	var extName, extEmail *string
	switch s.Party.Type {
	case properties.PartyTypeUser:
		uid := *s.Party.UserID
		userID = &uid
	case properties.PartyTypeOrganization:
		oid := *s.Party.OrganizationID
		orgID = &oid
	case properties.PartyTypeExternal:
		n := strings.TrimSpace(s.Party.ExternalName)
		e := strings.ToLower(strings.TrimSpace(s.Party.ExternalEmail))
		extName = &n
		extEmail = &e
	}

	_, err := db.Exec(ctx, `
		INSERT INTO property_stakeholders (
			property_id, party_type, party_user_id, party_org_id, party_external_name, party_external_email, role, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, uuid.UUID(s.PropertyID), string(s.Party.Type), userID, orgID, extName, extEmail, string(s.Role), s.CreatedAt)
	if err != nil {
		// unique violation -> duplicate stakeholder
		if strings.Contains(err.Error(), "uq_property_stakeholder") || strings.Contains(err.Error(), "duplicate key") {
			return properties.ErrDuplicateStakeholder
		}
		return err
	}
	return nil
}

func (r *StakeholderRepository) FindByProperty(ctx context.Context, db properties.DBExecutor, propertyID properties.PropertyID) ([]properties.PropertyStakeholder, error) {
	if propertyID.IsZero() {
		return nil, fmt.Errorf("property id must not be zero")
	}
	rows, err := db.Query(ctx, `
		SELECT property_id, party_type, party_user_id, party_org_id, party_external_name, party_external_email, role, created_at
		FROM property_stakeholders
		WHERE property_id = $1
		ORDER BY created_at ASC, role ASC
	`, uuid.UUID(propertyID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []properties.PropertyStakeholder
	for rows.Next() {
		var pid uuid.UUID
		var partyType, role string
		var puID, poID *uuid.UUID
		var extName, extEmail *string
		var created time.Time
		if err := rows.Scan(&pid, &partyType, &puID, &poID, &extName, &extEmail, &role, &created); err != nil {
			return nil, err
		}
		party := properties.Party{Type: properties.PartyType(partyType)}
		switch party.Type {
		case properties.PartyTypeUser:
			if puID != nil {
				uid := *puID
				party.UserID = &uid
			}
		case properties.PartyTypeOrganization:
			if poID != nil {
				oid := *poID
				party.OrganizationID = &oid
			}
		case properties.PartyTypeExternal:
			if extName != nil {
				party.ExternalName = *extName
			}
			if extEmail != nil {
				party.ExternalEmail = *extEmail
			}
		}
		stake := properties.PropertyStakeholder{
			PropertyID: properties.PropertyID(pid),
			Party:      party,
			Role:       properties.StakeholderRole(role),
			CreatedAt:  created,
		}
		out = append(out, stake)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []properties.PropertyStakeholder{}
	}
	return out, nil
}

func (r *StakeholderRepository) HasAccess(ctx context.Context, db properties.DBExecutor, propertyID properties.PropertyID, userID string) (bool, error) {
	if propertyID.IsZero() {
		return false, fmt.Errorf("property id must not be zero")
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user id: %w", err)
	}
	var exists int
	err = db.QueryRow(ctx, `
		SELECT 1
		FROM property_stakeholders ps
		LEFT JOIN memberships m ON m.organization_id = ps.party_org_id AND m.user_id = $2
		WHERE ps.property_id = $1
		  AND ((ps.party_type = 'user' AND ps.party_user_id = $2)
		    OR (ps.party_type = 'organization' AND m.user_id IS NOT NULL))
		LIMIT 1
	`, uuid.UUID(propertyID), uid).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return exists == 1, nil
}

func (r *StakeholderRepository) FindAccessiblePropertyIDs(ctx context.Context, db properties.DBExecutor, userID string) ([]properties.PropertyID, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	rows, err := db.Query(ctx, `
		SELECT DISTINCT ps.property_id
		FROM property_stakeholders ps
		LEFT JOIN memberships m ON m.organization_id = ps.party_org_id AND m.user_id = $1
		WHERE (ps.party_type = 'user' AND ps.party_user_id = $1)
		   OR (ps.party_type = 'organization' AND m.user_id IS NOT NULL)
		ORDER BY ps.property_id ASC
	`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []properties.PropertyID
	for rows.Next() {
		var pid uuid.UUID
		if err := rows.Scan(&pid); err != nil {
			return nil, err
		}
		ids = append(ids, properties.PropertyID(pid))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if ids == nil {
		ids = []properties.PropertyID{}
	}
	return ids, nil
}

// timeScan helps scanning TIMESTAMPTZ into time.Time regardless of driver
type timeScan struct {
	Time time.Time
}

func (t *timeScan) Scan(src interface{}) error {
	switch v := src.(type) {
	case time.Time:
		t.Time = v
		return nil
	case string:
		tt, err := time.Parse(time.RFC3339Nano, v)
		if err != nil {
			tt, err = time.Parse("2006-01-02 15:04:05.999999999-07", v)
			if err != nil {
				return err
			}
		}
		t.Time = tt
		return nil
	default:
		return fmt.Errorf("unsupported time scan type %T", src)
	}
}

// Ensure pgx.ErrNoRows handling is available
var _ = pgx.ErrNoRows

var _ properties.StakeholderRepository = (*StakeholderRepository)(nil)
