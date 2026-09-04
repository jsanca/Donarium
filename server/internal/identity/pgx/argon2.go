package pgx

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/application"
)

var (
	ErrArgon2InvalidFormat   = errors.New("argon2: invalid format")
	ErrArgon2Algorithm       = errors.New("argon2: invalid algorithm")
	ErrArgon2Version         = errors.New("argon2: invalid version")
	ErrArgon2MissingParam    = errors.New("argon2: missing required parameter")
	ErrArgon2DuplicateParam  = errors.New("argon2: duplicate parameter")
	ErrArgon2InvalidValue    = errors.New("argon2: invalid numeric value")
	ErrArgon2MemoryTooHigh   = errors.New("argon2: memory exceeds limit")
	ErrArgon2IterationsTooHigh = errors.New("argon2: iterations exceed limit")
	ErrArgon2ParallelismTooHigh = errors.New("argon2: parallelism exceeds limit")
	ErrArgon2EmptySalt       = errors.New("argon2: empty salt")
	ErrArgon2EmptyHash       = errors.New("argon2: empty hash")
	ErrArgon2Base64Salt      = errors.New("argon2: invalid base64 salt")
	ErrArgon2Base64Hash      = errors.New("argon2: invalid base64 hash")
)

const (
	argon2MaxMemory     uint32 = 256 * 1024
	argon2MaxIterations uint32 = 10
	argon2MaxThreads    uint8  = 8
	argon2MaxKeyLen     uint32 = 128
)

type Argon2Hasher struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
	saltLen uint32
}

func NewArgon2Hasher() *Argon2Hasher {
	return &Argon2Hasher{
		time:    3,
		memory:  64 * 1024,
		threads: 2,
		keyLen:  32,
		saltLen: 16,
	}
}

func (h *Argon2Hasher) Hash(password []byte) (identity.PasswordHash, error) {
	salt := make([]byte, h.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey(password, salt, h.time, h.memory, h.threads, h.keyLen)

	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		h.memory, h.time, h.threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return identity.PasswordHash(encoded), nil
}

func (h *Argon2Hasher) Verify(password []byte, encodedHash identity.PasswordHash) error {
	raw := string(encodedHash)
	parts := strings.Split(raw, "$")

	if len(parts) != 6 {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, ErrArgon2InvalidFormat)
	}

	if parts[1] != "argon2id" {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, ErrArgon2Algorithm)
	}

	if parts[2] != "v=19" {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, ErrArgon2Version)
	}

	params, err := parseArgon2Params(parts[3])
	if err != nil {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, err)
	}

	memory, ok := params["m"]
	if !ok {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, ErrArgon2MissingParam)
	}
	iterations, ok := params["t"]
	if !ok {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, ErrArgon2MissingParam)
	}
	parallelism, ok := params["p"]
	if !ok {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, ErrArgon2MissingParam)
	}

	if memory == 0 || iterations == 0 || parallelism == 0 {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, ErrArgon2InvalidValue)
	}

	if memory > argon2MaxMemory {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, ErrArgon2MemoryTooHigh)
	}
	if iterations > argon2MaxIterations {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, ErrArgon2IterationsTooHigh)
	}
	if parallelism > uint32(argon2MaxThreads) {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, ErrArgon2ParallelismTooHigh)
	}

	if parts[4] == "" {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, ErrArgon2EmptySalt)
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, ErrArgon2Base64Salt)
	}
	if len(salt) == 0 {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, ErrArgon2EmptySalt)
	}

	if parts[5] == "" {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, ErrArgon2EmptyHash)
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, ErrArgon2Base64Hash)
	}
	if len(expectedHash) == 0 || uint32(len(expectedHash)) > argon2MaxKeyLen {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, ErrArgon2InvalidValue)
	}

	computedHash := argon2.IDKey(password, salt, iterations, memory, uint8(parallelism), uint32(len(expectedHash)))

	if subtle.ConstantTimeCompare(computedHash, expectedHash) != 1 {
		return fmt.Errorf("%w: %w", identity.ErrInvalidCredentials, identity.ErrInvalidCredentials)
	}

	return nil
}

func parseArgon2Params(raw string) (map[string]uint32, error) {
	result := make(map[string]uint32)
	if raw == "" {
		return result, nil
	}
	for _, p := range strings.Split(raw, ",") {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			return nil, ErrArgon2InvalidFormat
		}
		key := kv[0]
		if key != "m" && key != "t" && key != "p" {
			return nil, ErrArgon2InvalidFormat
		}
		if _, exists := result[key]; exists {
			return nil, ErrArgon2DuplicateParam
		}
		v, err := parseArgon2Uint32(kv[1])
		if err != nil {
			return nil, err
		}
		result[key] = v
	}
	return result, nil
}

func parseArgon2Uint32(raw string) (uint32, error) {
	if raw == "" {
		return 0, ErrArgon2InvalidValue
	}
	v, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0, ErrArgon2InvalidValue
	}
	if v > math.MaxUint32 {
		return 0, ErrArgon2InvalidValue
	}
	return uint32(v), nil
}

var _ application.PasswordHasher = (*Argon2Hasher)(nil)
