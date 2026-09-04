package identity

import "testing"

func TestPlatformRole_Valid(t *testing.T) {
	tests := []struct {
		role  PlatformRole
		valid bool
	}{
		{PlatformRoleSuperAdmin, true},
		{PlatformRole("admin"), false},
		{PlatformRole(""), false},
		{PlatformRole("super_admin"), true},
	}

	for _, tt := range tests {
		if got := tt.role.Valid(); got != tt.valid {
			t.Errorf("PlatformRole(%q).Valid() = %v, want %v", tt.role, got, tt.valid)
		}
	}
}

func TestOrganizationRole_Valid(t *testing.T) {
	tests := []struct {
		role  OrganizationRole
		valid bool
	}{
		{OrganizationRoleOwner, true},
		{OrganizationRole("manager"), false},
		{OrganizationRole("tenant"), false},
		{OrganizationRole(""), false},
		{OrganizationRole("owner"), true},
	}

	for _, tt := range tests {
		if got := tt.role.Valid(); got != tt.valid {
			t.Errorf("OrganizationRole(%q).Valid() = %v, want %v", tt.role, got, tt.valid)
		}
	}
}
