# Example: CHANGES REQUIRED Architecture Review

This is an annotated example of a review that resulted in a CHANGES REQUIRED verdict.

---

# AUTH-003R — Authentication Middleware Architecture Review

- **Reviewer:** Elito
- **Date:** 2026-07-22
- **Scope:** AUTH-003 — Session Middleware and Protected Routes
- **Verdict:** **CHANGES REQUIRED**

## Executive Summary

AUTH-003 implements session middleware that reads cookies, verifies tokens, and attaches a principal to the request context. The core middleware chain is correct and stateless. However, authorization role guards are implemented inside the authentication middleware package, a default context selection depends on unordered database results, and the protected route has no router-level integration test. These three issues must be resolved before closeout.

## Architecture Assessment

### Dependency Direction

The `AuthenticationMiddleware` depends on `SessionVerifier` and `PrincipalResolver` — domain ports. It constructs an `AuthenticatedPrincipal` and stores it in request context. It does not import authorization logic, database drivers, or external services. Direction is correct.

The `PrincipalResolverService` depends on `UserRepository`, `PlatformGrantRepository`, `MembershipRepository`, and `OrganizationRepository` — all domain ports. It does not import HTTP, cookies, JSON, or Chi. Direction is correct.

**Issue found**: `authz_middleware.go` in the same HTTP package implements `RequirePlatformRole` and `RequireOrganizationRole`. These are authorization decisions, not authentication. See AAR-101.

### Boundaries

| Layer | Component | Assessment |
|---|---|---|
| Domain | `User`, `Credential`, `Membership`, `OrganizationRole` | Clean. |
| Application | `PrincipalResolverService`, `AuthenticateUserService` | Depends on domain ports. Clean. |
| Application | `AuthenticatedPrincipal`, `SessionClaims` | Data types; no behavior. Acceptable in application layer. |
| HTTP | `AuthenticationMiddleware` | Correct: cookie → verify → resolve → context. |
| HTTP | `RequirePlatformRole`, `RequireOrganizationRole` | **Boundary violation**: authorization in authentication package. See AAR-101. |
| HTTP | `MeHandler` | Clean: reads principal from context, returns JSON. |

### Contracts

- `GET /api/auth/me` returns principal JSON with `userId`, `email`, `displayName`, `organizationContexts`, `defaultContext`.
- Session token (`sessionToken`) is tagged `json:"-"` and excluded from serialization — correct.
- Error envelope is consistent: `{"error":"message"}`.
- `401` for missing/invalid/expired sessions and deleted users.
- `500` for repository operational failures.

**Issue found**: No router-level test exercises the exact public path `/api/auth/me` through the Chi middleware chain. See AAR-301.

### Statelessness

**Confirmed.** No server-side session cache, map, or shared state. Each request verifies the cookie, loads current grants/memberships/organizations from the database, and constructs a request-scoped principal. Tokens carry only `sub`, `iat`, `exp` — no permissions frozen in the token.

### Determinism

**Issue found**: `MembershipRepo.FindByUser` queries memberships without `ORDER BY`. `determineDefaultContext` selects the first returned organization. The same user with multiple memberships may get different default contexts across requests depending on database row ordering. See AAR-102.

### Adapter Isolation

The verifier, resolver, and middleware are tested with fake implementations. The `Clock` interface enables deterministic time in session verification tests. Adapter isolation is sound.

### Error Propagation

| Failure | Domain Error | HTTP Status | Assessment |
|---|---|---|---|
| Token missing | — | 401 | Correct. |
| Invalid signature | `ErrInvalidSession` | 401 | Correct. |
| Expired token | `ErrExpiredSession` | 401 | Correct. |
| User deleted | `ErrInvalidCredentials` | 401 | Correct. |
| DB unreachable | wrapped error | 500 | Correct: operational failure is distinct from auth failure. |

### Security

- HMAC verification uses constant-time comparison before payload decoding.
- Credentials, password hashes, and signing keys are absent from all responses.
- Session cookie is HttpOnly and SameSite=Lax.
- Cookie `Secure` flag is environment-aware (enabled in staging/production).

## Findings

### AAR-101 — Authorization middleware implemented inside authentication slice

- **Severity:** MAJOR
- **Description:** `authz_middleware.go` in the HTTP authentication package implements `HasPlatformRole`, `HasOrganizationRole`, `RequirePlatformRole`, and `RequireOrganizationRole`. These are role-based authorization guards.
- **Evidence:** `server/internal/auth/http/authz_middleware.go` contains `RequirePlatformRole` that checks `principal.PlatformRoles` and returns `403`.
- **Impact:** Authorization code in the authentication package conflates two distinct architectural concerns. Future protected routes that mount these guards would embed authorization into every authentication-adjacent component. It also creates an unreviewed API surface before any protected business endpoint exists.
- **Recommendation:** Move authorization guards to a separate `authorization/` package. Keep only `AuthenticationMiddleware` and the principal-presence check in the authentication slice.
- **Disposition:** Required before AUTH-003 closeout.

### AAR-102 — Default organization context is nondeterministic

- **Severity:** MAJOR
- **Description:** `MembershipRepo.FindByUser` queries memberships without `ORDER BY`. `determineDefaultContext` selects `orgCtxs[0]` — the first returned row. Database row ordering is not guaranteed.
- **Evidence:** `FindByUser` SQL: `SELECT ... FROM memberships WHERE user_id = $1` — no `ORDER BY`. `determineDefaultContext` uses `orgCtxs[0].OrganizationID`.
- **Impact:** A user with multiple organization memberships may see different default contexts on different requests, even with identical persisted state. The stateless principal reconstruction violates determinism.
- **Recommendation:** Add `ORDER BY created_at ASC` and document the rule: "the earliest membership by creation date is the default context."
- **Disposition:** Required before AUTH-003 closeout.

### AAR-301 — Public route not integration-tested

- **Severity:** MINOR
- **Description:** `MeHandler` and `AuthenticationMiddleware` are tested with direct handler calls, but no test constructs the full Chi router and exercises `GET /api/auth/me` through the middleware chain.
- **Evidence:** All existing tests call `handler.Me(rec, req)` directly without going through `chi.Router`.
- **Impact:** Route registration, middleware ordering, path normalization, and method-not-allowed behavior are untested.
- **Recommendation:** Add a focused router-level test that constructs the real middleware chain and exercises the exact public path with fakes.
- **Disposition:** Follow-up in AUTH-003-F1.

### AAR-302 — Expiration boundary not tested

- **Severity:** MINOR
- **Description:** The verifier uses `now.After(exp)` for expiration checks. The exact boundary case `now == exp` is not specified or tested.
- **Evidence:** `HMACSessionHandler.Verify` tests cover `now > exp` and `now < exp`, but not `now == exp`.
- **Impact:** Ambiguous behavior at the expiration boundary. Minor; affects only the millisecond when a token transitions from valid to expired.
- **Recommendation:** Document the boundary policy (inclusive or exclusive) and add a boundary test.
- **Disposition:** Follow-up in a future session-protocol slice.

## Positive Findings

- Stateless request-by-request reconstruction is correctly implemented.
- HMAC verification validates signature before decoding payload content.
- Constant-time comparison is used for cryptographic verification.
- `PrincipalResolver` is reused by both login and read-side middleware.
- Middleware has a clean happy path and short-circuits failures without calling next.
- Credentials, keys, and tokens are absent from all responses.
- Typed context keys prevent collisions with other middleware.
- Operational failures (DB down) are distinct from authentication failures (wrong password).
- Cookie security flags are environment-aware.

## Risk Assessment

| Dimension | Assessment |
|---|---|
| Maintainability | Good component separation; authorization code in this slice adds avoidable confusion. |
| Extensibility | Verifier and resolver ports are reusable; authorization guards should be separate. |
| Security | HMAC verification and stateless resolution are sound; no permissions in tokens. |
| Testability | Strong seams for verifier/resolver/clock; router-level test coverage is missing. |

## Recommendation

**CHANGES REQUIRED.** Resolve AAR-101 (move authorization guards out of authentication slice) and AAR-102 (add `ORDER BY` and document deterministic rule). Add the router-level integration test from AAR-301. After resolution, request a targeted re-review.

MINOR findings AAR-301 and AAR-302 have documented follow-ups and are not blocking.

## References

- [AUTH-003 task definition](../../tasks/AUTH-003-session-middleware.md)
- [UC-002 — Authenticate User](../../../knowledge/use-cases/UC-002-authenticate-user.md)
- [Authentication design document](../../../knowledge/design/authentication.md)
