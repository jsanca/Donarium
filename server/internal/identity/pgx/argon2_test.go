package pgx_test

import (
	"errors"
	"testing"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/pgx"
)

var hasher = pgx.NewArgon2Hasher()

func hashPassword(t *testing.T, password string) identity.PasswordHash {
	t.Helper()
	h, err := hasher.Hash([]byte(password))
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	return h
}

func TestVerify_ValidHash(t *testing.T) {
	h := hashPassword(t, "ValidP@ss1")
	if err := hasher.Verify([]byte("ValidP@ss1"), h); err != nil {
		t.Errorf("expected success, got: %v", err)
	}
}

func TestVerify_WrongPassword(t *testing.T) {
	h := hashPassword(t, "ValidP@ss1")
	err := hasher.Verify([]byte("WrongPass1!"), h)
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestVerify_InvalidAlgorithm(t *testing.T) {
	bad := identity.PasswordHash("$argon2i$v=19$m=65536,t=3,p=2$c29tZXNhbHQ$c29tZWhhc2g")
	err := hasher.Verify([]byte("pass"), bad)
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestVerify_InvalidVersion(t *testing.T) {
	bad := identity.PasswordHash("$argon2id$v=20$m=65536,t=3,p=2$c29tZXNhbHQ$c29tZWhhc2g")
	err := hasher.Verify([]byte("pass"), bad)
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestVerify_TruncatedHash(t *testing.T) {
	bad := identity.PasswordHash("$argon2id$v=19$m=65536")
	err := hasher.Verify([]byte("pass"), bad)
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestVerify_ExtraSegments(t *testing.T) {
	bad := identity.PasswordHash("$argon2id$v=19$m=65536,t=3,p=2$c2FsdA$c2FsdA$extra")
	err := hasher.Verify([]byte("pass"), bad)
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestVerify_MissingParameter(t *testing.T) {
	bad := identity.PasswordHash("$argon2id$v=19$m=65536,t=3$c2FsdA$aGFzaA")
	err := hasher.Verify([]byte("pass"), bad)
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestVerify_DuplicateParameter(t *testing.T) {
	bad := identity.PasswordHash("$argon2id$v=19$m=65536,m=32768,t=3,p=2$c2FsdA$aGFzaA")
	err := hasher.Verify([]byte("pass"), bad)
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestVerify_UnknownParameter(t *testing.T) {
	bad := identity.PasswordHash("$argon2id$v=19$m=65536,t=3,p=2,x=1$c2FsdA$aGFzaA")
	err := hasher.Verify([]byte("pass"), bad)
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestVerify_InvalidNumber(t *testing.T) {
	bad := identity.PasswordHash("$argon2id$v=19$m=notanumber,t=3,p=2$c2FsdA$aGFzaA")
	err := hasher.Verify([]byte("pass"), bad)
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestVerify_ZeroMemory(t *testing.T) {
	bad := identity.PasswordHash("$argon2id$v=19$m=0,t=3,p=2$c2FsdA$aGFzaA")
	err := hasher.Verify([]byte("pass"), bad)
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestVerify_ExcessiveMemory(t *testing.T) {
	bad := identity.PasswordHash("$argon2id$v=19$m=999999,t=3,p=2$c2FsdA$aGFzaA")
	err := hasher.Verify([]byte("pass"), bad)
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestVerify_ExcessiveIterations(t *testing.T) {
	bad := identity.PasswordHash("$argon2id$v=19$m=65536,t=999,p=2$c2FsdA$aGFzaA")
	err := hasher.Verify([]byte("pass"), bad)
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestVerify_ExcessiveParallelism(t *testing.T) {
	bad := identity.PasswordHash("$argon2id$v=19$m=65536,t=3,p=999$c2FsdA$aGFzaA")
	err := hasher.Verify([]byte("pass"), bad)
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestVerify_InvalidBase64Salt(t *testing.T) {
	bad := identity.PasswordHash("$argon2id$v=19$m=65536,t=3,p=2$!!!notbase64!!!$aGFzaA")
	err := hasher.Verify([]byte("pass"), bad)
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestVerify_InvalidBase64Hash(t *testing.T) {
	bad := identity.PasswordHash("$argon2id$v=19$m=65536,t=3,p=2$c2FsdA$!!!notbase64!!!")
	err := hasher.Verify([]byte("pass"), bad)
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestVerify_EmptySalt(t *testing.T) {
	bad := identity.PasswordHash("$argon2id$v=19$m=65536,t=3,p=2$$c29tZWhhc2g")
	err := hasher.Verify([]byte("pass"), bad)
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestVerify_EmptyHash(t *testing.T) {
	bad := identity.PasswordHash("$argon2id$v=19$m=65536,t=3,p=2$c2FsdA$")
	err := hasher.Verify([]byte("pass"), bad)
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestVerify_NoPanicOnArbitraryInput(t *testing.T) {
	inputs := []identity.PasswordHash{
		"",
		"$",
		"$$$$$$",
		"$argon2id$v=19$m=1,t=1,p=1$c2FsdA$aGFzaA",
		"$argon2id$v=19$m=1,t=1,p=1$AA$AA",
		"not a hash at all",
	}
	for _, input := range inputs {
		_ = hasher.Verify([]byte("pass"), input)
	}
}

func TestVerify_HashRoundtrip(t *testing.T) {
	password := "MySecurePass123!"
	h := hashPassword(t, password)
	if err := hasher.Verify([]byte(password), h); err != nil {
		t.Fatalf("roundtrip failed: %v", err)
	}
}
