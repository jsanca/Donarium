package pgx

import (
	"context"
	"errors"
	"time"

	"donarium/server/internal/identity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type MembershipRepo struct{}

func NewMembershipRepo() *MembershipRepo {
	return &MembershipRepo{}
}

func (r *MembershipRepo) Create(ctx context.Context, db identity.DBExecutor, m identity.Membership) error {
	_, err := db.Exec(ctx,
		`INSERT INTO memberships (user_id, organization_id, role, created_at)
		 VALUES ($1, $2, $3, $4)`,
		uuid.UUID(m.UserID), uuid.UUID(m.OrganizationID),
		string(m.Role), m.CreatedAt,
	)
	if err != nil {
		return translateError(err, "create membership")
	}
	return nil
}

func (r *MembershipRepo) FindByUserAndOrg(ctx context.Context, db identity.DBExecutor, userID identity.UserID, orgID identity.OrganizationID) (identity.Membership, error) {
	row := db.QueryRow(ctx,
		`SELECT user_id, organization_id, role, created_at
		 FROM memberships WHERE user_id = $1 AND organization_id = $2`,
		uuid.UUID(userID), uuid.UUID(orgID),
	)

	var m identity.Membership
	var uid, oid uuid.UUID
	var role string
	var createdAt time.Time

	if err := row.Scan(&uid, &oid, &role, &createdAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.Membership{}, identity.ErrMembershipNotFound
		}
		return identity.Membership{}, err
	}

	m.UserID = identity.UserID(uid)
	m.OrganizationID = identity.OrganizationID(oid)
	m.Role = identity.OrganizationRole(role)
	m.CreatedAt = createdAt
	return m, nil
}

func (r *MembershipRepo) FindByUser(ctx context.Context, db identity.DBExecutor, userID identity.UserID) ([]identity.Membership, error) {
	rows, err := db.Query(ctx,
		`SELECT user_id, organization_id, role, created_at
		 FROM memberships WHERE user_id = $1
		 ORDER BY created_at ASC, organization_id ASC`,
		uuid.UUID(userID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []identity.Membership
	for rows.Next() {
		var m identity.Membership
		var uid, oid uuid.UUID
		var role string
		var createdAt time.Time
		if err := rows.Scan(&uid, &oid, &role, &createdAt); err != nil {
			return nil, err
		}
		m.UserID = identity.UserID(uid)
		m.OrganizationID = identity.OrganizationID(oid)
		m.Role = identity.OrganizationRole(role)
		m.CreatedAt = createdAt
		result = append(result, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
