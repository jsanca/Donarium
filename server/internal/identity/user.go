package identity

import (
	"time"

	"github.com/google/uuid"
)

type UserID uuid.UUID

func NewUserID() UserID {
	return UserID(uuid.New())
}

func (id UserID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

type User struct {
	ID          UserID
	Email       string
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewUser(email, displayName string) (User, error) {
	if email == "" {
		return User{}, ErrEmptyEmail
	}
	if displayName == "" {
		return User{}, ErrEmptyDisplayName
	}

	now := time.Now().UTC()
	return User{
		ID:          NewUserID(),
		Email:       email,
		DisplayName: displayName,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
