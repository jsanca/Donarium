package pgx

import (
	"context"
	"errors"
	"fmt"

	"donarium/server/internal/properties"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) Create(ctx context.Context, db properties.DBExecutor, p properties.Property) error {
	if p.ID.IsZero() {
		return fmt.Errorf("property id must not be zero")
	}
	_, err := db.Exec(ctx, `
		INSERT INTO properties (
			id, display_name, classification,
			address_street, address_city, address_state, address_postal_code, address_country,
			rental_cadence, standard_rent, created_by, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`,
		uuid.UUID(p.ID), p.DisplayName, string(p.Classification),
		p.Address.Street, p.Address.City, p.Address.State, p.Address.PostalCode, p.Address.Country,
		string(p.RentalCadence), p.StandardRent, p.CreatedBy, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *Repository) FindByID(ctx context.Context, db properties.DBExecutor, id properties.PropertyID) (properties.Property, error) {
	if id.IsZero() {
		return properties.Property{}, properties.ErrPropertyNotFound
	}
	var p properties.Property
	var pid uuid.UUID
	var createdBy uuid.UUID
	var classification, cadence string

	err := db.QueryRow(ctx, `
		SELECT id, display_name, classification,
		       address_street, address_city, address_state, address_postal_code, address_country,
		       rental_cadence, standard_rent, created_by, created_at, updated_at
		FROM properties WHERE id = $1
	`, uuid.UUID(id)).Scan(
		&pid, &p.DisplayName, &classification,
		&p.Address.Street, &p.Address.City, &p.Address.State, &p.Address.PostalCode, &p.Address.Country,
		&cadence, &p.StandardRent, &createdBy, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return properties.Property{}, properties.ErrPropertyNotFound
		}
		return properties.Property{}, err
	}
	p.ID = properties.PropertyID(pid)
	p.CreatedBy = createdBy
	p.Classification = properties.Classification(classification)
	p.RentalCadence = properties.RentalCadence(cadence)
	return p, nil
}

func (r *Repository) FindAccessibleByUser(ctx context.Context, db properties.DBExecutor, userID string) ([]properties.Property, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	rows, err := db.Query(ctx, `
		SELECT DISTINCT p.id, p.display_name, p.classification,
		       p.address_street, p.address_city, p.address_state, p.address_postal_code, p.address_country,
		       p.rental_cadence, p.standard_rent, p.created_by, p.created_at, p.updated_at
		FROM properties p
		JOIN property_stakeholders ps ON ps.property_id = p.id
		LEFT JOIN memberships m ON m.organization_id = ps.party_org_id AND m.user_id = $1
		WHERE (ps.party_type = 'user' AND ps.party_user_id = $1)
		   OR (ps.party_type = 'organization' AND m.user_id IS NOT NULL)
		ORDER BY p.created_at ASC, p.id ASC
	`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []properties.Property
	for rows.Next() {
		var p properties.Property
		var pid uuid.UUID
		var createdBy uuid.UUID
		var classification, cadence string
		if err := rows.Scan(
			&pid, &p.DisplayName, &classification,
			&p.Address.Street, &p.Address.City, &p.Address.State, &p.Address.PostalCode, &p.Address.Country,
			&cadence, &p.StandardRent, &createdBy, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		p.ID = properties.PropertyID(pid)
		p.CreatedBy = createdBy
		p.Classification = properties.Classification(classification)
		p.RentalCadence = properties.RentalCadence(cadence)
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []properties.Property{}
	}
	return out, nil
}

var _ properties.Repository = (*Repository)(nil)
