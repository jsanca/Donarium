## Recovery Checkpoint

### 1. Original Objective

DON-008.1: Implement UC-002 authentication — email + password → credential
verification → session token issuance. Expose `POST /api/auth/login`.

### 2. Architecture

```
POST /api/auth/login → LoginHandler → AuthenticateUserService
    → EmailNormalizer, UserRepo, CredentialRepo, Argon2Hasher,
      PlatformGrantRepo, MembershipRepo, OrganizationRepo, SessionIssuer
```

### 3. Completed Work

| Layer | Files |
|---|---|
| Domain | `errors.go` (+ErrInvalidCredentials), `executor.go` (+Rows interface), `repository.go` (+FindByUser) |
| Application | `authentication/authenticate_user.go`, `authenticated_principal.go`, `session_issuer.go` |
| PGX adapter | `argon2.go` (+Verify), `executor.go` (Query→Rows), `membership_repository.go` (+FindByUser), `session.go` |
| HTTP | `login_handler.go`, `dto.go` (+login types), `runtime.go` (+login route) |
| Config | `config.go` (+SessionSigningKey, +SessionTTL) |
| Wiring | `main.go` (auth service + session issuer) |
| Tests | `login_handler_test.go` (8), `authenticate_user_test.go` (5) |
| Updated fakes | `setup_test.go` (+Verify, +FindByUser, +Rows), `setup_vertical_test.go` (Query→Rows) |

### 4. Key Decisions

- **Session token:** HMAC-SHA256, format `base64(payload).base64(signature)`, no JWT dependency
- **Same 401 message:** "the email or password is incorrect" for unknown email AND wrong password
- **Default context:** Organization context (owner) preferred over platform context (super_admin)
- **Cookie:** HttpOnly, SameSite=Lax, Path=/

### 5. Validation Status

| Check | Result |
|---|---|
| `go vet ./...` | PASS |
| `go test ./...` | PASS (80/80) |
| QA Docker: setup → login → 200 + cookie + principal | PASS |
| Wrong password → 401 "the email or password is incorrect" | PASS |
| Unknown email → 401 "the email or password is incorrect" | PASS |
| Invalid email → 400 | PASS |
| Missing fields → 400 | PASS |
| GET /api/auth/login → 405 JSON | PASS |
| Data integrity (no writes during login) | PASS |

### 6. Current Blocker

None.

### 7. Remaining Work

Out of scope for DON-008.1: `/api/auth/me`, logout, refresh, authorization middleware,
protected endpoints, rate limiting, frontend login.

### 8. Checkpoint Status

RESOLVED
