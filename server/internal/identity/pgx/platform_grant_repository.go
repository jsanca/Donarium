package pgx

import (
	"context"
	"errors"
	"time"

	"donarium/server/internal/identity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PlatformGrantRepo struct{}

func NewPlatformGrantRepo() *PlatformGrantRepo {
	return &PlatformGrantRepo{}
}

func (r *PlatformGrantRepo) Create(ctx context.Context, db identity.DBExecutor, grant identity.PlatformGrant) error {
	_, err := db.Exec(ctx,
		`INSERT INTO platform_grants (user_id, role, created_at)
		 VALUES ($1, $2, $3)`,
		uuid.UUID(grant.UserID), string(grant.Role), grant.CreatedAt,
	)
	if err != nil {
		return translateError(err, "create platform grant")
	}
	return nil
}

func (r *PlatformGrantRepo) FindByUser(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.PlatformGrant, error) {
	row := db.QueryRow(ctx,
		`SELECT user_id, role, created_at
		 FROM platform_grants WHERE user_id = $1`,
		uuid.UUID(userID),
	)

	var g identity.PlatformGrant
	var uid uuid.UUID
	var role string
	var createdAt time.Time

	if err := row.Scan(&uid, &role, &createdAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.PlatformGrant{}, identity.ErrMembershipNotFound
		}
		return identity.PlatformGrant{}, err
	}

	g.UserID = identity.UserID(uid)
	g.Role = identity.PlatformRole(role)
	g.CreatedAt = createdAt
	return g, nil
}
