package application

import "donarium/server/internal/identity"

type PasswordHasher interface {
	Hash(password []byte) (identity.PasswordHash, error)
	Verify(password []byte, encodedHash identity.PasswordHash) error
}
