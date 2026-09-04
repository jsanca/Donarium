## Recovery Checkpoint

### 1. Original Objective

DON-007.1-F1: Refine the Identity domain model before introducing persistence.
Strongly typed IDs, separate PlatformRole/OrganizationRole, PasswordHash type,
immutable PasswordPolicy, constructor validation, explicit Membership identity.

### 2. Completed Work

- `server/internal/identity/role.go` — split into PlatformRole (SUPER_ADMIN) and
  OrganizationRole (OWNER). Removed manager, tenant, maintenance_staff roles.
- `server/internal/identity/user.go` — UserID.IsZero() helper, NewUser returns
  (User, error) with validation (empty email, empty display name)
- `server/internal/identity/credential.go` — PasswordHash semantic type,
  UserID replaces uuid.UUID, CredentialID.IsZero(), NewCredential validates
  empty userID and passwordHash
- `server/internal/identity/organization.go` — CreatedBy uses UserID,
  OrganizationID.IsZero(), slug pattern validation, NewOrganization validates
  empty name/slug/invalid slug/empty createdBy
- `server/internal/identity/membership.go` — UserID + OrganizationID typed IDs,
  OrganizationRole replaces Role, composite identity documented as (UserID,
  OrganizationID), no surrogate key
- `server/internal/identity/password_policy.go` — all fields unexported
  (minLength, requireUpper, etc.), only DefaultPasswordPolicy() constructor
- `server/internal/identity/errors.go` — 20 sentinel errors (added validation
  errors: ErrEmptyEmail, ErrEmptyDisplayName, ErrInvalidSlug, etc.)
- `server/internal/identity/repository.go` — all interfaces use typed IDs
  (UserID, OrganizationID) instead of uuid.UUID
- `server/internal/identity/service.go` — SetupResult uses UserID + OrganizationID
- `server/internal/identity/password_policy_test.go` — extended to 21 tests
  covering password policy + all constructor validations + credential validations
- `server/internal/identity/role_test.go` — PlatformRole.Valid() + OrganizationRole.Valid()

### 3. Files Changed

All identity domain files rewritten. No outside files modified.

### 4. Current Repository State

- No raw uuid.UUID in domain relationships — all typed IDs
- PlatformRole and OrganizationRole separated
- PasswordHash is a semantic type
- Membership identity is explicit composite key
- PasswordPolicy immutable
- All constructors validate basic invariants
- 36/36 tests pass (23 identity + 7 health + 6 runtime)
- Lint 0 issues, build clean

### 5. Validation Status

| Check | Result |
|---|---|
| `go vet ./...` | PASS |
| `go test ./...` | PASS (36/36) |
| `golangci-lint run` | PASS (0 issues) |
| `go build ./cmd/donarium/` | PASS |
| Docker Compose | PASS |

### 6. Current Blocker

None.

### 7. Remaining Work

None for F1. Ready for DON-007.2 (persistence).

### 8. Checkpoint Status

RESOLVED
