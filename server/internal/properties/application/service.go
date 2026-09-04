package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"donarium/server/internal/properties"

	"github.com/google/uuid"
)

type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context, db properties.DBExecutor) error) error
}

type StakeholderInput struct {
	Party PartyInput `json:"party"`
	Role  string     `json:"role"`
}

type PartyInput struct {
	Type           string `json:"type"`
	UserID         string `json:"userId,omitempty"`
	OrganizationID string `json:"organizationId,omitempty"`
	ExternalName   string `json:"externalName,omitempty"`
	ExternalEmail  string `json:"externalEmail,omitempty"`
}

type RegisterPropertyCommand struct {
	DisplayName    string
	Classification string
	Address        properties.Address
	RentalCadence  string
	StandardRent   int64
	Stakeholders   []StakeholderInput
}

type RegisterPropertyResult struct {
	Property     properties.Property
	Stakeholders []properties.PropertyStakeholder
}

type Service struct {
	repo            properties.Repository
	stakeholderRepo properties.StakeholderRepository
	tx              TransactionManager
}

func NewService(repo properties.Repository, tx TransactionManager) *Service {
	return &Service{repo: repo, tx: tx}
}

func NewServiceWithStakeholders(repo properties.Repository, stakeholderRepo properties.StakeholderRepository, tx TransactionManager) *Service {
	return &Service{repo: repo, stakeholderRepo: stakeholderRepo, tx: tx}
}

func (s *Service) RegisterProperty(ctx context.Context, cmd RegisterPropertyCommand, actorUserID string) (RegisterPropertyResult, error) {
	if strings.TrimSpace(actorUserID) == "" {
		return RegisterPropertyResult{}, fmt.Errorf("actor user id is required")
	}
	actorUUID, err := uuid.Parse(actorUserID)
	if err != nil {
		return RegisterPropertyResult{}, fmt.Errorf("actor user id is not valid: %w", err)
	}

	if err := properties.ValidateDisplayName(cmd.DisplayName); err != nil {
		return RegisterPropertyResult{}, err
	}

	classification, err := properties.ParseClassification(cmd.Classification)
	if err != nil {
		return RegisterPropertyResult{}, err
	}

	if err := cmd.Address.Validate(); err != nil {
		return RegisterPropertyResult{}, err
	}

	cadence, err := properties.ParseRentalCadence(cmd.RentalCadence)
	if err != nil {
		return RegisterPropertyResult{}, err
	}

	if err := properties.ValidateStandardRent(cmd.StandardRent); err != nil {
		return RegisterPropertyResult{}, err
	}

	// Parse and validate stakeholders
	stakeholders, err := parseStakeholders(cmd.Stakeholders)
	if err != nil {
		return RegisterPropertyResult{}, err
	}

	// If none provided, default to actor as both OWNER and MANAGER (Stitch option 1: "I am both").
	// This keeps EP-001.02 callers (which send no stakeholders) working and satisfies post-condition
	// that at least one stakeholder ties the actor.
	if len(stakeholders) == 0 {
		uid := actorUUID
		stakeholders = []parsedStakeholder{
			{Party: properties.Party{Type: properties.PartyTypeUser, UserID: &uid}, Role: properties.StakeholderRoleOwner},
			{Party: properties.Party{Type: properties.PartyTypeUser, UserID: &uid}, Role: properties.StakeholderRoleManager},
		}
	}

	// Deduplicate by (partyType, reference, role) to enforce A-12 uniqueness before DB.
	seen := make(map[string]bool)
	deduped := make([]parsedStakeholder, 0, len(stakeholders))
	for _, st := range stakeholders {
		key := fmt.Sprintf("%s|%s|%s", st.Party.Type, st.Party.ReferenceKey(), st.Role)
		if seen[key] {
			return RegisterPropertyResult{}, fmt.Errorf("%w: duplicate stakeholder %s %s", properties.ErrDuplicateStakeholder, st.Party.Type, st.Role)
		}
		seen[key] = true
		deduped = append(deduped, st)
	}
	stakeholders = deduped

	now := time.Now().UTC()
	prop := properties.Property{
		ID:             properties.NewPropertyID(),
		DisplayName:    strings.TrimSpace(cmd.DisplayName),
		Classification: classification,
		Address: properties.Address{
			Street:     strings.TrimSpace(cmd.Address.Street),
			City:       strings.TrimSpace(cmd.Address.City),
			State:      strings.TrimSpace(cmd.Address.State),
			PostalCode: strings.TrimSpace(cmd.Address.PostalCode),
			Country:    strings.TrimSpace(cmd.Address.Country),
		},
		RentalCadence: cadence,
		StandardRent:  cmd.StandardRent,
		CreatedBy:     actorUUID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Build stakeholder entities with property ID and timestamp
	entities := make([]properties.PropertyStakeholder, 0, len(stakeholders))
	for _, st := range stakeholders {
		entities = append(entities, properties.PropertyStakeholder{
			PropertyID: prop.ID,
			Party:      st.Party,
			Role:       st.Role,
			CreatedAt:  now,
		})
	}

	// Validate at least one stakeholder ties the actor directly or via membership.
	// This validation requires DB for organization membership check, so defer to transaction.
	var validationErr error
	if s.stakeholderRepo == nil {
		// No stakeholder repo injected (legacy wiring) — skip membership validation; only check direct tie.
		if !tiesActorDirectly(entities, actorUUID.String()) {
			validationErr = fmt.Errorf("%w: at least one stakeholder must tie the acting user", properties.ErrNoStakeholder)
		}
	}

	if validationErr != nil {
		return RegisterPropertyResult{}, validationErr
	}

	if err := s.tx.WithinTransaction(ctx, func(ctx context.Context, db properties.DBExecutor) error {
		// Validate existence of referenced users/orgs and membership tying
		for _, st := range stakeholders {
			switch st.Party.Type {
			case properties.PartyTypeUser:
				if err := checkUserExists(ctx, db, *st.Party.UserID); err != nil {
					return err
				}
			case properties.PartyTypeOrganization:
				if err := checkOrganizationExists(ctx, db, *st.Party.OrganizationID); err != nil {
					return err
				}
			case properties.PartyTypeExternal:
				// No existence check; external is contact-only
			}
		}
		// Ensure at least one stakeholder ties actor (with membership check)
		if s.stakeholderRepo != nil {
			// We need to check tying after we know stakeholders are valid.
			// For org stakeholders, verify actor is member of that org.
			ties, err := stakeholdersTieActor(ctx, db, entities, actorUUID.String())
			if err != nil {
				return err
			}
			if !ties {
				return fmt.Errorf("%w: at least one stakeholder must tie the acting user", properties.ErrNoStakeholder)
			}
		}

		if err := s.repo.Create(ctx, db, prop); err != nil {
			return err
		}
		for _, ent := range entities {
			// Use stakeholder repo if available, else try via properties repo's DB directly
			if s.stakeholderRepo != nil {
				if err := s.stakeholderRepo.Create(ctx, db, ent); err != nil {
					return err
				}
			} else {
				// Fallback direct insert (for legacy wiring)
				if err := insertStakeholderDirect(ctx, db, ent); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		if errors.Is(err, properties.ErrDuplicateStakeholder) {
			return RegisterPropertyResult{}, err
		}
		if errors.Is(err, properties.ErrNoStakeholder) || errors.Is(err, properties.ErrInvalidParty) || errors.Is(err, properties.ErrInvalidStakeholderRole) {
			return RegisterPropertyResult{}, err
		}
		// Map not-found of user/org to 400 invalid stakeholder? Keep as invalid party
		if strings.Contains(err.Error(), "user not found") || strings.Contains(err.Error(), "organization not found") {
			return RegisterPropertyResult{}, fmt.Errorf("%w: %v", properties.ErrInvalidParty, err)
		}
		return RegisterPropertyResult{}, err
	}

	return RegisterPropertyResult{Property: prop, Stakeholders: entities}, nil
}

func (s *Service) ListAccessible(ctx context.Context, db properties.DBExecutor, actorUserID string) ([]properties.Property, error) {
	if strings.TrimSpace(actorUserID) == "" {
		return nil, fmt.Errorf("actor user id is required")
	}
	return s.repo.FindAccessibleByUser(ctx, db, actorUserID)
}

func (s *Service) GetByID(ctx context.Context, db properties.DBExecutor, id string, actorUserID string) (properties.Property, error) {
	parsed, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return properties.Property{}, properties.ErrPropertyNotFound
	}
	prop, err := s.repo.FindByID(ctx, db, properties.PropertyID(parsed))
	if err != nil {
		return properties.Property{}, err
	}
	// Authorization via stakeholders, not CreatedBy
	if s.stakeholderRepo != nil {
		has, err := s.stakeholderRepo.HasAccess(ctx, db, prop.ID, actorUserID)
		if err != nil {
			return properties.Property{}, err
		}
		if !has {
			return properties.Property{}, properties.ErrPropertyNotFound
		}
	} else {
		// Legacy fallback: created_by check
		if prop.CreatedBy.String() != actorUserID {
			return properties.Property{}, properties.ErrPropertyNotFound
		}
	}
	return prop, nil
}

type parsedStakeholder struct {
	Party properties.Party
	Role  properties.StakeholderRole
}

func parseStakeholders(inputs []StakeholderInput) ([]parsedStakeholder, error) {
	var out []parsedStakeholder
	for i, in := range inputs {
		if strings.TrimSpace(in.Role) == "" {
			return nil, fmt.Errorf("%w: stakeholder %d role is required", properties.ErrInvalidStakeholderRole, i)
		}
		role, err := properties.ParseStakeholderRole(in.Role)
		if err != nil {
			return nil, err
		}
		partyType := properties.PartyType(strings.ToLower(strings.TrimSpace(in.Party.Type)))
		if partyType == "" {
			// Infer from presence of fields if type omitted (backwards compat)
			if in.Party.UserID != "" {
				partyType = properties.PartyTypeUser
			} else if in.Party.OrganizationID != "" {
				partyType = properties.PartyTypeOrganization
			} else if in.Party.ExternalEmail != "" || in.Party.ExternalName != "" {
				partyType = properties.PartyTypeExternal
			} else {
				return nil, fmt.Errorf("%w: stakeholder %d party type is required", properties.ErrInvalidParty, i)
			}
		}
		var party properties.Party
		switch partyType {
		case properties.PartyTypeUser:
			uidStr := strings.TrimSpace(in.Party.UserID)
			if uidStr == "" {
				return nil, fmt.Errorf("%w: stakeholder %d userId is required", properties.ErrInvalidParty, i)
			}
			uid, err := uuid.Parse(uidStr)
			if err != nil {
				return nil, fmt.Errorf("%w: stakeholder %d userId is not valid: %v", properties.ErrInvalidParty, i, err)
			}
			party = properties.Party{Type: properties.PartyTypeUser, UserID: &uid}
		case properties.PartyTypeOrganization:
			oidStr := strings.TrimSpace(in.Party.OrganizationID)
			if oidStr == "" {
				return nil, fmt.Errorf("%w: stakeholder %d organizationId is required", properties.ErrInvalidParty, i)
			}
			oid, err := uuid.Parse(oidStr)
			if err != nil {
				return nil, fmt.Errorf("%w: stakeholder %d organizationId is not valid: %v", properties.ErrInvalidParty, i, err)
			}
			party = properties.Party{Type: properties.PartyTypeOrganization, OrganizationID: &oid}
		case properties.PartyTypeExternal:
			name := strings.TrimSpace(in.Party.ExternalName)
			email := strings.TrimSpace(in.Party.ExternalEmail)
			if name == "" || email == "" {
				return nil, fmt.Errorf("%w: stakeholder %d external name and email are required", properties.ErrInvalidParty, i)
			}
			party = properties.Party{Type: properties.PartyTypeExternal, ExternalName: name, ExternalEmail: email}
		default:
			return nil, fmt.Errorf("%w: stakeholder %d party type is not valid: %s", properties.ErrInvalidParty, i, partyType)
		}
		if err := party.Validate(); err != nil {
			return nil, fmt.Errorf("%w: stakeholder %d %v", properties.ErrInvalidParty, i, err)
		}
		out = append(out, parsedStakeholder{Party: party, Role: role})
	}
	return out, nil
}

func tiesActorDirectly(entities []properties.PropertyStakeholder, actorID string) bool {
	for _, e := range entities {
		if e.Party.Type == properties.PartyTypeUser && e.Party.UserID != nil && e.Party.UserID.String() == actorID {
			return true
		}
	}
	return false
}

func stakeholdersTieActor(ctx context.Context, db properties.DBExecutor, entities []properties.PropertyStakeholder, actorID string) (bool, error) {
	actorUUID, _ := uuid.Parse(actorID)
	for _, e := range entities {
		switch e.Party.Type {
		case properties.PartyTypeUser:
			if e.Party.UserID != nil && e.Party.UserID.String() == actorID {
				return true, nil
			}
		case properties.PartyTypeOrganization:
			if e.Party.OrganizationID != nil {
				// Check membership
				var exists int
				err := db.QueryRow(ctx, `SELECT 1 FROM memberships WHERE user_id = $1 AND organization_id = $2 LIMIT 1`, actorUUID, *e.Party.OrganizationID).Scan(&exists)
				if err == nil && exists == 1 {
					return true, nil
				}
			}
		case properties.PartyTypeExternal:
			// never ties
		}
	}
	return false, nil
}

func checkUserExists(ctx context.Context, db properties.DBExecutor, userID uuid.UUID) error {
	var exists int
	err := db.QueryRow(ctx, `SELECT 1 FROM users WHERE id = $1 LIMIT 1`, userID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	return nil
}

func checkOrganizationExists(ctx context.Context, db properties.DBExecutor, orgID uuid.UUID) error {
	var exists int
	err := db.QueryRow(ctx, `SELECT 1 FROM organizations WHERE id = $1 LIMIT 1`, orgID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}
	return nil
}

func insertStakeholderDirect(ctx context.Context, db properties.DBExecutor, s properties.PropertyStakeholder) error {
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
		n := s.Party.ExternalName
		e := s.Party.ExternalEmail
		extName = &n
		extEmail = &e
	}
	_, err := db.Exec(ctx, `
		INSERT INTO property_stakeholders (
			property_id, party_type, party_user_id, party_org_id, party_external_name, party_external_email, role, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, uuid.UUID(s.PropertyID), string(s.Party.Type), userID, orgID, extName, extEmail, string(s.Role), s.CreatedAt)
	return err
}
