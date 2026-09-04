package authentication

import "time"

type SessionIssuer interface {
	Issue(sub string) (string, error)
}

type SessionVerifier interface {
	Verify(token string) (SessionClaims, error)
}

type SessionClaims struct {
	Subject   string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

func NewSystemClock() Clock { return realClock{} }
