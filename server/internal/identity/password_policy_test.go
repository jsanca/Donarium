package identity

import (
	"errors"
	"testing"
)

func TestPasswordPolicy_DefaultValidPassword(t *testing.T) {
	policy := DefaultPasswordPolicy()
	if err := policy.Validate("ValidP@ss1"); err != nil {
		t.Errorf("expected valid password to pass, got: %v", err)
	}
}

func TestPasswordPolicy_TooShort(t *testing.T) {
	policy := DefaultPasswordPolicy()
	err := policy.Validate("Short1!")
	if err == nil {
		t.Fatal("expected error for short password")
	}
	if !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got: %v", err)
	}
}

func TestPasswordPolicy_MissingUpper(t *testing.T) {
	policy := DefaultPasswordPolicy()
	err := policy.Validate("alllowercase1!")
	if err == nil {
		t.Fatal("expected error for missing uppercase")
	}
	if !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got: %v", err)
	}
}

func TestPasswordPolicy_MissingLower(t *testing.T) {
	policy := DefaultPasswordPolicy()
	err := policy.Validate("ALLUPPERCASE1!")
	if err == nil {
		t.Fatal("expected error for missing lowercase")
	}
	if !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got: %v", err)
	}
}

func TestPasswordPolicy_MissingDigit(t *testing.T) {
	policy := DefaultPasswordPolicy()
	err := policy.Validate("NoDigitsHere!")
	if err == nil {
		t.Fatal("expected error for missing digit")
	}
	if !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got: %v", err)
	}
}

func TestPasswordPolicy_MissingSpecial(t *testing.T) {
	policy := DefaultPasswordPolicy()
	err := policy.Validate("NoSpecialChar1")
	if err == nil {
		t.Fatal("expected error for missing special character")
	}
	if !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got: %v", err)
	}
}

func TestPasswordPolicy_ImmutableFields(t *testing.T) {
	policy := DefaultPasswordPolicy()
	if err := policy.Validate("ValidP@ss1"); err != nil {
		t.Errorf("default policy should accept valid password: %v", err)
	}
}

func TestNewUser_ValidatesEmptyEmail(t *testing.T) {
	_, err := NewUser("", "Display Name")
	if err == nil {
		t.Fatal("expected error for empty email")
	}
	if !errors.Is(err, ErrEmptyEmail) {
		t.Errorf("expected ErrEmptyEmail, got %v", err)
	}
}

func TestNewUser_ValidatesEmptyDisplayName(t *testing.T) {
	_, err := NewUser("user@example.com", "")
	if err == nil {
		t.Fatal("expected error for empty display name")
	}
	if !errors.Is(err, ErrEmptyDisplayName) {
		t.Errorf("expected ErrEmptyDisplayName, got %v", err)
	}
}

func TestNewUser_CreatesValidUser(t *testing.T) {
	user, err := NewUser("user@example.com", "Test User")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "user@example.com" {
		t.Errorf("expected email to be 'user@example.com', got %q", user.Email)
	}
	if user.ID.IsZero() {
		t.Error("expected non-zero user ID")
	}
	if user.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestNewOrganization_ValidatesEmptyName(t *testing.T) {
	_, err := NewOrganization("", "my-org", NewUserID())
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !errors.Is(err, ErrEmptyName) {
		t.Errorf("expected ErrEmptyName, got %v", err)
	}
}

func TestNewOrganization_ValidatesEmptySlug(t *testing.T) {
	_, err := NewOrganization("My Org", "", NewUserID())
	if err == nil {
		t.Fatal("expected error for empty slug")
	}
	if !errors.Is(err, ErrEmptySlug) {
		t.Errorf("expected ErrEmptySlug, got %v", err)
	}
}

func TestNewOrganization_ValidatesInvalidSlug(t *testing.T) {
	tests := []string{"My Org", "UPPERCASE", "with spaces", "trailing-", "-leading"}
	for _, slug := range tests {
		_, err := NewOrganization("My Org", slug, NewUserID())
		if err == nil {
			t.Errorf("expected error for invalid slug %q", slug)
		}
		if !errors.Is(err, ErrInvalidSlug) {
			t.Errorf("expected ErrInvalidSlug for %q, got %v", slug, err)
		}
	}
}

func TestNewOrganization_ValidatesEmptyCreatedBy(t *testing.T) {
	_, err := NewOrganization("My Org", "my-org", UserID{})
	if err == nil {
		t.Fatal("expected error for empty createdBy")
	}
	if !errors.Is(err, ErrEmptyUserID) {
		t.Errorf("expected ErrEmptyUserID, got %v", err)
	}
}

func TestNewOrganization_ValidSlugPasses(t *testing.T) {
	tests := []string{"my-org", "org", "a-b-c", "org-123", "slug123"}
	for _, slug := range tests {
		org, err := NewOrganization("My Org", slug, NewUserID())
		if err != nil {
			t.Errorf("expected valid slug %q to pass, got: %v", slug, err)
		}
		if org.Slug != slug {
			t.Errorf("expected slug %q, got %q", slug, org.Slug)
		}
	}
}

func TestNewMembership_ValidatesEmptyUserID(t *testing.T) {
	_, err := NewMembership(UserID{}, NewOrganizationID(), OrganizationRoleOwner)
	if err == nil {
		t.Fatal("expected error for empty user ID")
	}
	if !errors.Is(err, ErrEmptyUserID) {
		t.Errorf("expected ErrEmptyUserID, got %v", err)
	}
}

func TestNewMembership_ValidatesEmptyOrgID(t *testing.T) {
	_, err := NewMembership(NewUserID(), OrganizationID{}, OrganizationRoleOwner)
	if err == nil {
		t.Fatal("expected error for empty org ID")
	}
	if !errors.Is(err, ErrEmptyOrganizationID) {
		t.Errorf("expected ErrEmptyOrganizationID, got %v", err)
	}
}

func TestNewMembership_ValidatesInvalidRole(t *testing.T) {
	_, err := NewMembership(NewUserID(), NewOrganizationID(), "invalid")
	if err == nil {
		t.Fatal("expected error for invalid role")
	}
	if !errors.Is(err, ErrInvalidOrganizationRole) {
		t.Errorf("expected ErrInvalidOrganizationRole, got %v", err)
	}
}

func TestNewMembership_CreatesValidMembership(t *testing.T) {
	userID := NewUserID()
	orgID := NewOrganizationID()
	m, err := NewMembership(userID, orgID, OrganizationRoleOwner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.UserID != userID {
		t.Error("UserID mismatch")
	}
	if m.OrganizationID != orgID {
		t.Error("OrganizationID mismatch")
	}
	if m.Role != OrganizationRoleOwner {
		t.Error("Role mismatch")
	}
}

func TestNewCredential_ValidatesEmptyUserID(t *testing.T) {
	_, err := NewCredential(UserID{}, PasswordHash("hash"))
	if err == nil {
		t.Fatal("expected error for empty user ID")
	}
	if !errors.Is(err, ErrEmptyUserID) {
		t.Errorf("expected ErrEmptyUserID, got %v", err)
	}
}

func TestNewCredential_ValidatesEmptyPasswordHash(t *testing.T) {
	_, err := NewCredential(NewUserID(), PasswordHash(""))
	if err == nil {
		t.Fatal("expected error for empty password hash")
	}
	if !errors.Is(err, ErrEmptyPasswordHash) {
		t.Errorf("expected ErrEmptyPasswordHash, got %v", err)
	}
}
