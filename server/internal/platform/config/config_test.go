package config_test

import (
	"os"
	"testing"

	"donarium/server/internal/platform/config"
)

func TestParseEnvironment_LocalDefault(t *testing.T) {
	env, err := config.ParseEnvironment("local")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env != config.EnvLocal {
		t.Errorf("expected local, got %q", env)
	}
}

func TestParseEnvironment_QA(t *testing.T) {
	env, err := config.ParseEnvironment("qa")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env != config.EnvQA {
		t.Errorf("expected qa, got %q", env)
	}
}

func TestParseEnvironment_UnknownFails(t *testing.T) {
	_, err := config.ParseEnvironment("mars")
	if err == nil {
		t.Fatal("expected error for unknown environment")
	}
}

func TestParseEnvironment_Staging(t *testing.T) {
	env, err := config.ParseEnvironment("staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env != config.EnvStaging {
		t.Errorf("expected staging, got %q", env)
	}
}

func TestParseEnvironment_Production(t *testing.T) {
	env, err := config.ParseEnvironment("production")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env != config.EnvProduction {
		t.Errorf("expected production, got %q", env)
	}
}

func TestProductionRejectsDevSigningKey(t *testing.T) {
	err := config.ValidateSessionSigningKey(config.EnvProduction, "donarium-dev-key-change-in-production")
	if err == nil {
		t.Fatal("expected error for dev key in production")
	}
}

func TestStagingRejectsDevSigningKey(t *testing.T) {
	err := config.ValidateSessionSigningKey(config.EnvStaging, "donarium-dev-key-change-in-production")
	if err == nil {
		t.Fatal("expected error for dev key in staging")
	}
}

func TestProductionRejectsShortKey(t *testing.T) {
	err := config.ValidateSessionSigningKey(config.EnvProduction, "short")
	if err == nil {
		t.Fatal("expected error for short key in production")
	}
}

func TestProductionAcceptsValidKey(t *testing.T) {
	err := config.ValidateSessionSigningKey(config.EnvProduction, "a-very-long-signing-key-that-is-at-least-32-chars")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLocalAllowsDevKey(t *testing.T) {
	err := config.ValidateSessionSigningKey(config.EnvLocal, "donarium-dev-key-change-in-production")
	if err != nil {
		t.Errorf("unexpected error for dev key in local: %v", err)
	}
}

func TestQAAllowsDevKey(t *testing.T) {
	err := config.ValidateSessionSigningKey(config.EnvQA, "donarium-dev-key-change-in-production")
	if err != nil {
		t.Errorf("unexpected error for dev key in qa: %v", err)
	}
}

func TestSessionTTL_InvalidDurationFails(t *testing.T) {
	_, err := config.ParseSessionTTL("not-a-duration")
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestSessionTTL_ZeroFails(t *testing.T) {
	_, err := config.ParseSessionTTL("0s")
	if err == nil {
		t.Fatal("expected error for zero TTL")
	}
}

func TestSessionTTL_NegativeFails(t *testing.T) {
	_, err := config.ParseSessionTTL("-1h")
	if err == nil {
		t.Fatal("expected error for negative TTL")
	}
}

func TestSessionTTL_Valid(t *testing.T) {
	d, err := config.ParseSessionTTL("1h30m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d == nil {
		t.Fatal("expected non-nil duration")
	}
}

func TestSessionTTL_EmptyReturnsNil(t *testing.T) {
	d, err := config.ParseSessionTTL("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != nil {
		t.Error("expected nil duration for empty string")
	}
}

func TestIsProduction_Staging(t *testing.T) {
	cfg := config.Config{Environment: config.EnvStaging}
	if !cfg.IsProduction() {
		t.Error("staging should be treated as production")
	}
}

func TestIsProduction_Production(t *testing.T) {
	cfg := config.Config{Environment: config.EnvProduction}
	if !cfg.IsProduction() {
		t.Error("production should be production")
	}
}

func TestIsProduction_Local(t *testing.T) {
	cfg := config.Config{Environment: config.EnvLocal}
	if cfg.IsProduction() {
		t.Error("local should not be production")
	}
}

func TestLoad_UnknownEnvFails(t *testing.T) {
	os.Setenv("DONARIUM_ENV", "mars")
	os.Setenv("POSTGRES_PASSWORD", "test")
	defer os.Unsetenv("DONARIUM_ENV")
	defer os.Unsetenv("POSTGRES_PASSWORD")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for unknown DONARIUM_ENV")
	}
}
