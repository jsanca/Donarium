package identity

import (
	"net/mail"
	"strings"
)

type DefaultEmailNormalizer struct{}

func NewDefaultEmailNormalizer() *DefaultEmailNormalizer {
	return &DefaultEmailNormalizer{}
}

func (n *DefaultEmailNormalizer) Normalize(email string) (string, error) {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		trimmed := strings.TrimSpace(email)
		addr, err = mail.ParseAddress("<" + trimmed + ">")
		if err != nil {
			return "", ErrInvalidEmail
		}
	}

	normalized := strings.TrimSpace(strings.ToLower(addr.Address))
	if normalized == "" || !strings.Contains(normalized, "@") {
		return "", ErrInvalidEmail
	}

	return normalized, nil
}
