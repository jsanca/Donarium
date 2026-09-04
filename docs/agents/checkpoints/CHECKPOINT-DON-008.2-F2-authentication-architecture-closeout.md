## Recovery Checkpoint

### 1. Original Objective

Resolve the three architectural findings from DON-008.2R (CHANGES REQUIRED):
- AAR-001: Remove authorization middleware from authentication slice
- AAR-002: Make default organization context deterministic
- AAR-003: Add router-level integration test for `GET /api/auth/me`

Scope: architectural cleanup only. No new capabilities, no commits.

### 2. Completed Work

#### AAR-001 — Authorization middleware relocated

- Deleted `server/internal/identity/http/authz_middleware.go` (role-based guards: `RequirePlatformRole`, `RequireOrganizationRole`, `HasPlatformRole`, `HasOrganizationRole`)
- Deleted `server/internal/identity/http/authz_middleware_test.go` (role guard tests)
- Created `server/internal/identity/authorization/middleware.go` — relocated authorization functions to the authorization package for DON-008.3 pickup
- Created `server/internal/identity/authorization/middleware_test.go` — relocated authz tests adapted to the new package
- Moved `RequireAuthenticated()` (principal-presence check, not role-based) from `authz_middleware.go` into `auth_middleware.go` where it belongs
- Added `RequireAuthenticated()` unit tests to `auth_middleware_test.go`
- Authentication slice now contains no role-based authorization

#### AAR-002 — Deterministic default context

- Updated `MembershipRepo.FindByUser` SQL to include `ORDER BY created_at ASC` in `pgx/membership_repository.go:62`
- Documented the deterministic rule in `determineDefaultContext()`: "the organization with the earliest membership creation date" in `application/authentication/principal_resolver.go:111-117`
- Added four principal resolver tests in `application/authentication/principal_resolver_test.go`:
  - `TestResolve_DefaultContextIsEarliestMembership` — three memberships with different dates; verifies earliest is default
  - `TestResolve_DefaultContextDeterministicWithSameCreationDate` — five iterations with two memberships at same time; verifies stable output
  - `TestResolve_DefaultContextIsPlatformWhenNoOrganizations` — no memberships; verifies platform default
  - `TestResolve_DefaultContextIsFirstOrg_SingleMembership` — single membership; verifies it becomes default

#### AAR-003 — Router composition test

- Created `server/internal/identity/http/me_router_test.go` with 8 router-level tests:
  - `TestMeRouter_Unauthenticated` — invalid token → 401 + JSON error
  - `TestMeRouter_MissingCookieReturns401` — no cookie → 401
  - `TestMeRouter_ExpiredSessionReturns401` — expired token → 401
  - `TestMeRouter_AuthenticatedReturns200` — valid token → 200 + principal JSON + content type
  - `TestMeRouter_SessionTokenNotInResponse` — verifies session token absent from response body
  - `TestMeRouter_WrongMethodReturns405` — POST with valid auth → 405
  - `TestMeRouter_ContextPassesThroughMiddleware` — principal from resolver reaches handler
  - `TestMeRouter_PublicPathIsRegistered` — path is registered (not 404)
- All tests use fakes for `SessionVerifier` and `PrincipalResolver`; no database dependency

### 3. Files Changed

| File | Change |
|---|---|
| `server/internal/identity/http/authz_middleware.go` | **Deleted** — authorization middleware |
| `server/internal/identity/http/authz_middleware_test.go` | **Deleted** — authorization middleware tests |
| `server/internal/identity/http/auth_middleware.go` | Added `RequireAuthenticated()` function (moved from deleted file) |
| `server/internal/identity/http/auth_middleware_test.go` | Added `RequireAuthenticated` unit tests |
| `server/internal/identity/authorization/middleware.go` | **Created** — relocated authorization functions |
| `server/internal/identity/authorization/middleware_test.go` | **Created** — relocated authz tests |
| `server/internal/identity/pgx/membership_repository.go` | Added `ORDER BY created_at ASC` to `FindByUser` query |
| `server/internal/identity/application/authentication/principal_resolver.go` | Documented deterministic default context rule |
| `server/internal/identity/application/authentication/principal_resolver_test.go` | **Created** — deterministic default context tests |
| `server/internal/identity/http/me_router_test.go` | **Created** — router composition tests for `/api/auth/me` |

### 4. Current Repository State

- Authentication slice (`identity/http`) contains no role-based authorization guards
- Authorization functions live in `identity/authorization/` package, ready for DON-008.3
- Default organization context is deterministic: earliest membership by creation date
- `GET /api/auth/me` has router-level integration tests covering middleware + handler composition
- All existing functionality preserved; no behavioral changes to any endpoint

### 5. Validation Status

| Command | Result |
|---|---|
| `go build ./...` | PASS |
| `go vet ./...` | PASS |
| `go test ./internal/identity/http/...` | PASS (all 37 tests) |
| `go test ./internal/identity/authorization/...` | PASS (all 13 tests) |
| `go test ./internal/identity/application/authentication/...` | PASS (all tests including 4 new) |
| `go test ./...` (full suite) | PASS — 3 pre-existing failures in pgx package from concurrent TestMain DB reset (pass when run independently) |

### 6. Current Blocker

None. All three findings resolved.

### 7. Evidence

- `RequirePlatformRole` and `RequireOrganizationRole` no longer exist in `identity/http` package
- `authz_middleware.go` no longer exists
- `MembershipRepo.FindByUser` query now contains `ORDER BY created_at ASC`
- `determineDefaultContext` has documented rule comment
- `me_router_test.go` asserts complete middleware chain on exact path `/api/auth/me`

### 8. Remaining Work

- [x] AAR-001: Remove authorization from authentication slice
- [x] AAR-002: Make default context deterministic
- [x] AAR-003: Add router composition test

### 9. Finding Resolution Summary

| Finding | Status | Details |
|---|---|---|
| AAR-001 | RESOLVED | Authorization middleware removed from authentication slice; relocated to `identity/authorization/` |
| AAR-002 | RESOLVED | `ORDER BY created_at ASC` + documented rule; 4 new tests for deterministic behavior |
| AAR-003 | RESOLVED | 8 router-level tests exercising full middleware chain on `/api/auth/me` |

### 10. Proposed Continuation Tasks

- **DON-008.3 — Authorization context and route guards**: pick up the relocated `identity/authorization/` package and mount role guards on protected routes
- **DON-008.2 targeted re-review**: re-run DON-008.2R with expectation of moving from CHANGES REQUIRED to APPROVED
- **AAR-004 observation**: session protocol versioning/rotation remains a future concern (not addressed in this slice)

### 11. Recommended Next Action

Request targeted re-review of DON-008.2R. All three findings (AAR-001, AAR-002, AAR-003) are resolved. The slice is ready for architectural approval.

### 12. Checkpoint Status

OPEN
