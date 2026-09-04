package pgx

import (
	"context"
	"errors"
	"time"

	"donarium/server/internal/identity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OrganizationRepo struct{}

func NewOrganizationRepo() *OrganizationRepo {
	return &OrganizationRepo{}
}

func (r *OrganizationRepo) Create(ctx context.Context, db identity.DBExecutor, org identity.Organization) error {
	_, err := db.Exec(ctx,
		`INSERT INTO organizations (id, name, slug, created_by, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		uuid.UUID(org.ID), org.Name, org.Slug,
		uuid.UUID(org.CreatedBy), org.CreatedAt,
	)
	if err != nil {
		return translateError(err, "create organization")
	}
	return nil
}

func (r *OrganizationRepo) FindByID(ctx context.Context, db identity.DBExecutor, id identity.OrganizationID) (identity.Organization, error) {
	row := db.QueryRow(ctx,
		`SELECT id, name, slug, created_by, created_at
		 FROM organizations WHERE id = $1`,
		uuid.UUID(id),
	)
	return scanOrganization(row)
}

func (r *OrganizationRepo) FindBySlug(ctx context.Context, db identity.DBExecutor, slug string) (identity.Organization, error) {
	row := db.QueryRow(ctx,
		`SELECT id, name, slug, created_by, created_at
		 FROM organizations WHERE slug = $1`,
		slug,
	)
	return scanOrganization(row)
}

func (r *OrganizationRepo) ExistsAny(ctx context.Context, db identity.DBExecutor) (bool, error) {
	var count int
	row := db.QueryRow(ctx, `SELECT COUNT(*) FROM organizations`)
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func scanOrganization(row identity.RowScanner) (identity.Organization, error) {
	var o identity.Organization
	var id, createdBy uuid.UUID
	var createdAt time.Time

	if err := row.Scan(&id, &o.Name, &o.Slug, &createdBy, &createdAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.Organization{}, identity.ErrOrganizationNotFound
		}
		return identity.Organization{}, err
	}

	o.ID = identity.OrganizationID(id)
	o.CreatedBy = identity.UserID(createdBy)
	o.CreatedAt = createdAt
	return o, nil
}
