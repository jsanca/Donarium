package pgx

import (
	"context"
	"errors"
	"strings"
	"time"

	"donarium/server/internal/identity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepo struct{}

func NewUserRepo() *UserRepo {
	return &UserRepo{}
}

func (r *UserRepo) Create(ctx context.Context, db identity.DBExecutor, user identity.User) error {
	_, err := db.Exec(ctx,
		`INSERT INTO users (id, email, display_name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		uuid.UUID(user.ID), user.Email, user.DisplayName, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return translateError(err, "create user")
	}
	return nil
}

func (r *UserRepo) FindByID(ctx context.Context, db identity.DBExecutor, id identity.UserID) (identity.User, error) {
	row := db.QueryRow(ctx,
		`SELECT id, email, display_name, created_at, updated_at FROM users WHERE id = $1`,
		uuid.UUID(id),
	)
	return scanUser(row)
}

func (r *UserRepo) FindByEmail(ctx context.Context, db identity.DBExecutor, email string) (identity.User, error) {
	row := db.QueryRow(ctx,
		`SELECT id, email, display_name, created_at, updated_at FROM users WHERE email = $1`,
		email,
	)
	return scanUser(row)
}

func (r *UserRepo) ExistsByEmail(ctx context.Context, db identity.DBExecutor, email string) (bool, error) {
	var count int
	row := db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE email = $1`, email)
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func scanUser(row identity.RowScanner) (identity.User, error) {
	var u identity.User
	var id uuid.UUID
	var createdAt, updatedAt time.Time

	if err := row.Scan(&id, &u.Email, &u.DisplayName, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.User{}, identity.ErrUserNotFound
		}
		return identity.User{}, err
	}

	u.ID = identity.UserID(id)
	u.CreatedAt = createdAt
	u.UpdatedAt = updatedAt
	return u, nil
}

func translateError(err error, operation string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			switch {
			case strings.Contains(pgErr.ConstraintName, "users_email"):
				return identity.ErrDuplicateEmail
			case strings.Contains(pgErr.ConstraintName, "email"):
				return identity.ErrDuplicateEmail
			case strings.Contains(pgErr.ConstraintName, "slug"):
				return identity.ErrDuplicateSlug
			case strings.Contains(pgErr.ConstraintName, "membership"):
				return identity.ErrMembershipNotFound
			case strings.Contains(pgErr.ConstraintName, "platform_grant"):
				return identity.ErrMembershipNotFound
			}
		}
	}
	return err
}
