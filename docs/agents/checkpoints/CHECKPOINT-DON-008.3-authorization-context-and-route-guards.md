## Recovery Checkpoint

### 1. Original Objective

DON-008.3: Build authorization layer — composable middleware consuming
AuthenticatedPrincipal from context. Separate 401 (auth) from 403 (authz).

### 2. Implemented

| Component | Description |
|---|---|
| `RequireAuthenticated()` | Checks principal in context → 401 if missing |
| `RequirePlatformRole(role)` | Checks platform roles → 403 if missing |
| `RequireOrganizationRole(role)` | Checks org contexts → 403 if missing |
| `HasPlatformRole(principal, role)` | Pure function |
| `HasOrganizationRole(principal, role)` | Pure function |
| Route protection | `/api/auth/me` uses both auth + authz |

### 3. Validation

| Check | Result |
|---|---|
| `go vet ./...` | PASS |
| `go test ./...` | PASS |
| 401 (no cookie) vs 403 (no role) | CONFIRMED |
| Authz never touches cookies/tokens/repos | CONFIRMED |
| QA: /me with cookie → 200 | PASS |
| QA: /me without cookie → 401 | PASS |

### 4. Checkpoint Status

RESOLVED
