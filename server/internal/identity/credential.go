package identity

import (
	"time"

	"github.com/google/uuid"
)

type PasswordHash string

type CredentialID uuid.UUID

type Credential struct {
	ID           CredentialID
	UserID       UserID
	PasswordHash PasswordHash
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewCredentialID() CredentialID {
	return CredentialID(uuid.New())
}

func (id CredentialID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func NewCredential(userID UserID, passwordHash PasswordHash) (Credential, error) {
	if userID.IsZero() {
		return Credential{}, ErrEmptyUserID
	}
	if passwordHash == "" {
		return Credential{}, ErrEmptyPasswordHash
	}

	now := time.Now().UTC()
	return Credential{
		ID:           NewCredentialID(),
		UserID:       userID,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}
