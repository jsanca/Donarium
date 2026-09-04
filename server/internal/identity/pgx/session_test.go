package pgx_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/pgx"

	"github.com/google/uuid"
)

type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time { return c.now }

func newHandlerWithClock(t *testing.T, now time.Time) *pgx.HMACSessionHandler {
	t.Helper()
	return pgx.NewHMACSessionHandler("test-signing-key-32-chars!", 24*time.Hour, &fakeClock{now: now})
}

func issueToken(t *testing.T, h *pgx.HMACSessionHandler, sub string) string {
	t.Helper()
	token, err := h.Issue(sub)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	return token
}

func TestVerify_ValidToken(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	h := newHandlerWithClock(t, now)
	sub := uuid.New().String()
	token := issueToken(t, h, sub)

	claims, err := h.Verify(token)
	if err != nil {
		t.Fatalf("expected valid token: %v", err)
	}
	if claims.Subject != sub {
		t.Errorf("subject mismatch: %q != %q", claims.Subject, sub)
	}
	if !claims.IssuedAt.Equal(now) {
		t.Errorf("iat mismatch: %v != %v", claims.IssuedAt, now)
	}
}

func TestVerify_InvalidSignature(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	h1 := newHandlerWithClock(t, now)
	h2 := pgx.NewHMACSessionHandler("different-key-32-chars!!!", 24*time.Hour, &fakeClock{now: now})

	token := issueToken(t, h1, "00000000-0000-0000-0000-000000000001")

	_, err := h2.Verify(token)
	if !errors.Is(err, identity.ErrInvalidSession) {
		t.Errorf("expected ErrInvalidSession, got %v", err)
	}
}

func TestVerify_ModifiedPayload(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	h := newHandlerWithClock(t, now)
	sub := uuid.New().String()
	token := issueToken(t, h, sub)

	modified := "A" + token[1:]

	_, err := h.Verify(modified)
	if !errors.Is(err, identity.ErrInvalidSession) {
		t.Errorf("expected ErrInvalidSession, got %v", err)
	}
}

func TestVerify_ModifiedSignature(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	h := newHandlerWithClock(t, now)
	sub := uuid.New().String()
	token := issueToken(t, h, sub)

	token = token[:len(token)-1] + "X"

	_, err := h.Verify(token)
	if !errors.Is(err, identity.ErrInvalidSession) {
		t.Errorf("expected ErrInvalidSession, got %v", err)
	}
}

func TestVerify_MissingSegment(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	h := newHandlerWithClock(t, now)

	_, err := h.Verify("onlypayload")
	if !errors.Is(err, identity.ErrInvalidSession) {
		t.Errorf("expected ErrInvalidSession, got %v", err)
	}
}

func TestVerify_AdditionalSegments(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	h := newHandlerWithClock(t, now)

	_, err := h.Verify("a.b.c")
	if !errors.Is(err, identity.ErrInvalidSession) {
		t.Errorf("expected ErrInvalidSession, got %v", err)
	}
}

func TestVerify_InvalidBase64Payload(t *testing.T) {
	h := newHandlerWithClock(t, time.Now())

	_, err := h.Verify("!!!.c2ln")
	if !errors.Is(err, identity.ErrInvalidSession) {
		t.Errorf("expected ErrInvalidSession, got %v", err)
	}
}

func TestVerify_InvalidBase64Signature(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	h := newHandlerWithClock(t, now)
	token := issueToken(t, h, uuid.New().String())
	parts := strings.Split(token, ".")

	bad := parts[0] + ".!!!"

	_, err := h.Verify(bad)
	if !errors.Is(err, identity.ErrInvalidSession) {
		t.Errorf("expected ErrInvalidSession, got %v", err)
	}
}

func TestVerify_MalformedJSON(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	h := newHandlerWithClock(t, now)

	badJSON := base64.RawURLEncoding.EncodeToString([]byte("{not json"))
	sig := hmac.New(sha256.New, []byte("test-signing-key-32-chars!"))
	sig.Write([]byte(badJSON))
	sigEncoded := base64.RawURLEncoding.EncodeToString(sig.Sum(nil))

	_, err := h.Verify(badJSON + "." + sigEncoded)
	if !errors.Is(err, identity.ErrInvalidSession) {
		t.Errorf("expected ErrInvalidSession, got %v", err)
	}
}

func TestVerify_EmptySubject(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	h := newHandlerWithClock(t, now)

	payload := `{"sub":"","iat":1735689600,"exp":1735776000}`
	payloadEnc := base64.RawURLEncoding.EncodeToString([]byte(payload))
	mac := hmac.New(sha256.New, []byte("test-signing-key-32-chars!"))
	mac.Write([]byte(payloadEnc))
	sigEnc := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	_, err := h.Verify(payloadEnc + "." + sigEnc)
	if !errors.Is(err, identity.ErrInvalidSession) {
		t.Errorf("expected ErrInvalidSession, got %v", err)
	}
}

func TestVerify_NonUUIDSubject(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	h := newHandlerWithClock(t, now)

	payload := `{"sub":"not-a-uuid","iat":1735689600,"exp":1735776000}`
	payloadEnc := base64.RawURLEncoding.EncodeToString([]byte(payload))
	mac := hmac.New(sha256.New, []byte("test-signing-key-32-chars!"))
	mac.Write([]byte(payloadEnc))
	sigEnc := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	_, err := h.Verify(payloadEnc + "." + sigEnc)
	if !errors.Is(err, identity.ErrInvalidSession) {
		t.Errorf("expected ErrInvalidSession, got %v", err)
	}
}

func TestVerify_ExpiredToken(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	h := newHandlerWithClock(t, now)
	sub := uuid.New().String()
	token := issueToken(t, h, sub)

	later := now.Add(25 * time.Hour)
	verifyH := newHandlerWithClock(t, later)

	_, err := verifyH.Verify(token)
	if !errors.Is(err, identity.ErrExpiredSession) {
		t.Errorf("expected ErrExpiredSession, got %v", err)
	}
}

func TestVerify_ExpBeforeIat(t *testing.T) {
	h := newHandlerWithClock(t, time.Now())

	payload := `{"sub":"` + uuid.New().String() + `","iat":1735776000,"exp":1735689600}`
	payloadEnc := base64.RawURLEncoding.EncodeToString([]byte(payload))
	mac := hmac.New(sha256.New, []byte("test-signing-key-32-chars!"))
	mac.Write([]byte(payloadEnc))
	sigEnc := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	_, err := h.Verify(payloadEnc + "." + sigEnc)
	if !errors.Is(err, identity.ErrExpiredSession) {
		t.Errorf("expected ErrExpiredSession, got %v", err)
	}
}

func TestVerify_FutureIatBeyondClockSkew(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	futureIat := now.Add(2 * time.Minute)

	payload := `{"sub":"` + uuid.New().String() + `","iat":` + fmt.Sprintf("%d", futureIat.Unix()) + `,"exp":` + fmt.Sprintf("%d", futureIat.Add(24*time.Hour).Unix()) + `}`
	payloadEnc := base64.RawURLEncoding.EncodeToString([]byte(payload))
	mac := hmac.New(sha256.New, []byte("test-signing-key-32-chars!"))
	mac.Write([]byte(payloadEnc))
	sigEnc := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	h := newHandlerWithClock(t, now)
	_, err := h.Verify(payloadEnc + "." + sigEnc)
	if !errors.Is(err, identity.ErrInvalidSession) {
		t.Errorf("expected ErrInvalidSession, got %v", err)
	}
}

func TestVerify_NoPanicOnMalformedInput(t *testing.T) {
	h := newHandlerWithClock(t, time.Now())
	inputs := []string{"", ".", "..", "...", "a", "a.", ".b", "a.b.c.d"}
	for _, input := range inputs {
		_, _ = h.Verify(input)
	}
}
