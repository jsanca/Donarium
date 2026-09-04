package identity

import "time"

type PlatformGrant struct {
	UserID    UserID
	Role      PlatformRole
	CreatedAt time.Time
}

func NewPlatformGrant(userID UserID, role PlatformRole) (PlatformGrant, error) {
	if userID.IsZero() {
		return PlatformGrant{}, ErrEmptyUserID
	}
	if !role.Valid() {
		return PlatformGrant{}, ErrInvalidPlatformRole
	}

	return PlatformGrant{
		UserID:    userID,
		Role:      role,
		CreatedAt: time.Now().UTC(),
	}, nil
}
