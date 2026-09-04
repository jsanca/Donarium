## Recovery Checkpoint

### 1. Original Objective

DON-007.3: Implement the application service for UC-001 Initial Owner Setup.
Canonical service (business workflow, no tx), transactional wrapper, password
hashing, email normalization, and comprehensive tests.

### 2. Completed Work

**Application interfaces:**
- `identity/application/password.go` — PasswordHasher interface
- `identity/application/email.go` — EmailNormalizer interface
- `identity/application/transaction.go` — TransactionManager interface

**Services:**
- `identity/application/setup.go` — CanonicalSetupService with full business
  workflow: validate → check initialized → normalize email → validate password
  → check duplicates → hash → create 5 entities
- `identity/application/transactional_setup.go` — TransactionalSetupService:
  wraps canonical in WithinTransaction

**Infrastructure:**
- `identity/pgx/transaction.go` — pgx TransactionManager (Begin/Commit/Rollback)
- `identity/pgx/argon2.go` — Argon2id PasswordHasher (time=3, mem=64MB, threads=2)
- `identity/normalizer.go` — DefaultEmailNormalizer (trim, lowercase, validate format)
- `identity/application/setup_test.go` — 9 tests (success, already init, dup email,
  invalid password, invalid email, transactional commit, rollback, tx error, atomicity)

### 3. Files Changed

| File | Change |
|---|---|
| `identity/application/password.go` | Created — PasswordHasher interface |
| `identity/application/email.go` | Created — EmailNormalizer interface |
| `identity/application/transaction.go` | Created — TransactionManager interface |
| `identity/application/setup.go` | Created — CanonicalSetupService + DTOs |
| `identity/application/transactional_setup.go` | Created — TransactionalSetupService |
| `identity/application/setup_test.go` | Created — 9 application tests |
| `identity/pgx/transaction.go` | Created — pgx TransactionManager |
| `identity/pgx/argon2.go` | Created — Argon2id PasswordHasher |
| `identity/normalizer.go` | Created — DefaultEmailNormalizer |
| `server/go.mod` / `server/go.sum` | Updated — argon2 dependency |

### 4. Current Repository State

- UC-001 Initial Owner Setup fully implemented in application layer
- Canonical service has no transaction management
- Transactional service wraps canonical in WithinTransaction
- Password hashing via Argon2id (configurable parameters)
- Email normalization centralized
- 60/60 tests pass (23 domain + 9 app + 15 integration + 7 health + 6 runtime)
- Lint clean, build clean, Docker Compose healthy
- No HTTP, no authentication, no login

### 5. Validation Status

| Check | Result |
|---|---|
| `go vet ./...` | PASS |
| `go test ./...` | PASS (60/60) |
| `golangci-lint run` | PASS (0 issues) |
| `go build ./cmd/donarium/` | PASS |
| `docker compose up -d --build` | PASS |
| No HTTP | PASS |
| No commits | PASS |

### 6. Current Blocker

None.

### 7. Remaining Work

None for 007.3. Ready for DON-007.4 (REST layer).

### 8. Checkpoint Status

RESOLVED
