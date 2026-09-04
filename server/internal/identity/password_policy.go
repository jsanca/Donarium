package identity

import (
	"fmt"
	"unicode"
)

type PasswordPolicy struct {
	minLength      int
	requireUpper   bool
	requireLower   bool
	requireDigit   bool
	requireSpecial bool
}

func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		minLength:      8,
		requireUpper:   true,
		requireLower:   true,
		requireDigit:   true,
		requireSpecial: true,
	}
}

func (p PasswordPolicy) Validate(password string) error {
	if len(password) < p.minLength {
		return fmt.Errorf("%w: must be at least %d characters", ErrInvalidPassword, p.minLength)
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if p.requireUpper && !hasUpper {
		return fmt.Errorf("%w: must contain at least one uppercase letter", ErrInvalidPassword)
	}
	if p.requireLower && !hasLower {
		return fmt.Errorf("%w: must contain at least one lowercase letter", ErrInvalidPassword)
	}
	if p.requireDigit && !hasDigit {
		return fmt.Errorf("%w: must contain at least one digit", ErrInvalidPassword)
	}
	if p.requireSpecial && !hasSpecial {
		return fmt.Errorf("%w: must contain at least one special character", ErrInvalidPassword)
	}

	return nil
}
