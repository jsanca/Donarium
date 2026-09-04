package pgx

import (
	"context"
	"errors"
	"time"

	"donarium/server/internal/identity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CredentialRepo struct{}

func NewCredentialRepo() *CredentialRepo {
	return &CredentialRepo{}
}

func (r *CredentialRepo) Create(ctx context.Context, db identity.DBExecutor, cred identity.Credential) error {
	_, err := db.Exec(ctx,
		`INSERT INTO credentials (id, user_id, password_hash, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		uuid.UUID(cred.ID), uuid.UUID(cred.UserID),
		string(cred.PasswordHash), cred.CreatedAt, cred.UpdatedAt,
	)
	if err != nil {
		return translateError(err, "create credential")
	}
	return nil
}

func (r *CredentialRepo) FindByUserID(ctx context.Context, db identity.DBExecutor, userID identity.UserID) (identity.Credential, error) {
	row := db.QueryRow(ctx,
		`SELECT id, user_id, password_hash, created_at, updated_at
		 FROM credentials WHERE user_id = $1`,
		uuid.UUID(userID),
	)

	var c identity.Credential
	var id, uid uuid.UUID
	var hash string
	var createdAt, updatedAt time.Time

	if err := row.Scan(&id, &uid, &hash, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.Credential{}, identity.ErrCredentialNotFound
		}
		return identity.Credential{}, err
	}

	c.ID = identity.CredentialID(id)
	c.UserID = identity.UserID(uid)
	c.PasswordHash = identity.PasswordHash(hash)
	c.CreatedAt = createdAt
	c.UpdatedAt = updatedAt
	return c, nil
}
