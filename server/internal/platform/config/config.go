package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Environment string

const (
	EnvLocal      Environment = "local"
	EnvQA         Environment = "qa"
	EnvStaging    Environment = "staging"
	EnvProduction Environment = "production"
)

const devSigningKey = "donarium-dev-key-change-in-production"
const minSigningKeyLen = 32

var validEnvs = map[string]Environment{
	"local":      EnvLocal,
	"qa":         EnvQA,
	"staging":    EnvStaging,
	"production": EnvProduction,
}

type Config struct {
	Environment       Environment
	HTTPPort          string
	PostgresHost      string
	PostgresPort      string
	PostgresDB        string
	PostgresUser      string
	PostgresPassword  string
	PostgresSSLMode   string
	DatabaseConnectTimeout time.Duration
	ShutdownTimeout        time.Duration
	SessionSigningKey      string
	SessionTTL             time.Duration
}

func (c Config) IsProduction() bool {
	return c.Environment == EnvProduction || c.Environment == EnvStaging
}

func Load() (*Config, error) {
	env, err := ParseEnvironment(getEnv("DONARIUM_ENV", "local"))
	if err != nil {
		return nil, err
	}

	signingKey := os.Getenv("SESSION_SIGNING_KEY")
	if signingKey == "" && (env == EnvLocal || env == EnvQA) {
		signingKey = devSigningKey
	}

	ttl, err := ParseSessionTTL(os.Getenv("SESSION_TTL"))
	if err != nil {
		return nil, err
	}

	if ttl == nil {
		defaultTTL := 24 * time.Hour
		ttl = &defaultTTL
	}

	if err := ValidateSessionSigningKey(env, signingKey); err != nil {
		return nil, err
	}

	cfg := &Config{
		Environment:            env,
		HTTPPort:               getEnv("HTTP_PORT", "8080"),
		PostgresHost:           getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:           getEnv("POSTGRES_PORT", "5432"),
		PostgresDB:             getEnv("POSTGRES_DB", "donarium"),
		PostgresUser:           getEnv("POSTGRES_USER", "donarium"),
		PostgresPassword:       os.Getenv("POSTGRES_PASSWORD"),
		PostgresSSLMode:        getEnv("POSTGRES_SSLMODE", "disable"),
		DatabaseConnectTimeout: getDurationEnv("DATABASE_CONNECT_TIMEOUT", 10*time.Second),
		ShutdownTimeout:        getDurationEnv("SHUTDOWN_TIMEOUT", 10*time.Second),
		SessionSigningKey:      signingKey,
		SessionTTL:             *ttl,
	}

	if cfg.PostgresPassword == "" {
		return nil, fmt.Errorf("POSTGRES_PASSWORD is required")
	}

	if _, err := strconv.Atoi(cfg.HTTPPort); err != nil {
		return nil, fmt.Errorf("HTTP_PORT must be a valid port number: %w", err)
	}

	return cfg, nil
}

func ParseEnvironment(raw string) (Environment, error) {
	env, ok := validEnvs[strings.ToLower(strings.TrimSpace(raw))]
	if !ok {
		return "", fmt.Errorf("DONARIUM_ENV: unknown environment %q (valid: local, qa, staging, production)", raw)
	}
	return env, nil
}

func ParseSessionTTL(raw string) (*time.Duration, error) {
	if raw == "" {
		return nil, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return nil, fmt.Errorf("SESSION_TTL: invalid duration %q: %w", raw, err)
	}
	if d <= 0 {
		return nil, fmt.Errorf("SESSION_TTL: must be positive, got %v", d)
	}
	return &d, nil
}

func ValidateSessionSigningKey(env Environment, key string) error {
	if (env == EnvStaging || env == EnvProduction) && key == devSigningKey {
		return fmt.Errorf("SESSION_SIGNING_KEY: development key is not allowed in %s environment", env)
	}
	if (env == EnvStaging || env == EnvProduction) && len(key) < minSigningKeyLen {
		return fmt.Errorf("SESSION_SIGNING_KEY: must be at least %d characters in %s environment", minSigningKeyLen, env)
	}
	return nil
}

func (c Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDB,
		c.PostgresSSLMode,
	)
}

func (c Config) DatabaseURLRedacted() string {
	return fmt.Sprintf(
		"postgres://%s:***@%s:%s/%s?sslmode=%s",
		c.PostgresUser,
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDB,
		c.PostgresSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
