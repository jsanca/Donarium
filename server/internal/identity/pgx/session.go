package pgx

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/application/authentication"

	"github.com/google/uuid"
)

type sessionPayload struct {
	Sub string `json:"sub"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
}

type HMACSessionHandler struct {
	key   []byte
	ttl   time.Duration
	clock authentication.Clock
}

func NewHMACSessionIssuer(signingKey string, ttl time.Duration) *HMACSessionHandler {
	return &HMACSessionHandler{
		key:   []byte(signingKey),
		ttl:   ttl,
		clock: authentication.NewSystemClock(),
	}
}

func NewHMACSessionHandler(signingKey string, ttl time.Duration, clock authentication.Clock) *HMACSessionHandler {
	return &HMACSessionHandler{
		key:   []byte(signingKey),
		ttl:   ttl,
		clock: clock,
	}
}

func (s *HMACSessionHandler) Issue(sub string) (string, error) {
	now := s.clock.Now()
	payload := sessionPayload{
		Sub: sub,
		Iat: now.Unix(),
		Exp: now.Add(s.ttl).Unix(),
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal session payload: %w", err)
	}

	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadJSON)

	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(payloadEncoded))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return payloadEncoded + "." + signature, nil
}

func (s *HMACSessionHandler) Verify(token string) (authentication.SessionClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return authentication.SessionClaims{}, identity.ErrInvalidSession
	}

	payloadEncoded := parts[0]
	signatureEncoded := parts[1]

	if payloadEncoded == "" || signatureEncoded == "" {
		return authentication.SessionClaims{}, identity.ErrInvalidSession
	}

	sig, err := base64.RawURLEncoding.DecodeString(signatureEncoded)
	if err != nil {
		return authentication.SessionClaims{}, identity.ErrInvalidSession
	}

	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(payloadEncoded))
	expectedSig := mac.Sum(nil)

	if subtle.ConstantTimeCompare(sig, expectedSig) != 1 {
		return authentication.SessionClaims{}, identity.ErrInvalidSession
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadEncoded)
	if err != nil {
		return authentication.SessionClaims{}, identity.ErrInvalidSession
	}

	var payload sessionPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return authentication.SessionClaims{}, identity.ErrInvalidSession
	}

	if payload.Sub == "" {
		return authentication.SessionClaims{}, identity.ErrInvalidSession
	}
	if _, err := uuid.Parse(payload.Sub); err != nil {
		return authentication.SessionClaims{}, identity.ErrInvalidSession
	}

	iat := time.Unix(payload.Iat, 0).UTC()
	exp := time.Unix(payload.Exp, 0).UTC()

	if payload.Iat == 0 || payload.Exp == 0 {
		return authentication.SessionClaims{}, identity.ErrInvalidSession
	}
	if !exp.After(iat) {
		return authentication.SessionClaims{}, identity.ErrExpiredSession
	}

	now := s.clock.Now()
	if now.After(exp) {
		return authentication.SessionClaims{}, identity.ErrExpiredSession
	}

	allowedSkew := 1 * time.Minute
	if iat.After(now.Add(allowedSkew)) {
		return authentication.SessionClaims{}, identity.ErrInvalidSession
	}

	return authentication.SessionClaims{
		Subject:   payload.Sub,
		IssuedAt:  iat,
		ExpiresAt: exp,
	}, nil
}

var _ authentication.SessionIssuer = (*HMACSessionHandler)(nil)
var _ authentication.SessionVerifier = (*HMACSessionHandler)(nil)
