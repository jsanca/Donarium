## Recovery Checkpoint

### 1. Original Objective

DON-008.1-F3: Resolve AAR-003. Distinguish "entity not found" from operational
failures in authentication persistence layer.

### 2. Fixed Patterns

| Location | Before | After |
|---|---|---|
| `FindByEmail` error | Always → 401 | `ErrUserNotFound` → 401; other → 500 |
| `FindByUserID` error | Always → 401 | `ErrCredentialNotFound` → 401; other → 500 |
| `FindByUser` (grant) error | Always → `nil, nil` (swallow) | `ErrMembershipNotFound` → empty; other → 500 |
| `FindByID` (org) error | Always → `continue` (swallow) | All errors → 500 |

### 3. Validation Status

| Check | Result |
|---|---|
| `go vet ./...` | PASS |
| `go test ./...` | PASS (126/126) |
| User not found → 401 | PASS |
| Credential not found → 401 | PASS |
| Operational error finding user → 500 (not 401) | PASS |
| Operational error finding credential → 500 (not 401) | PASS |
| Grant not found → continues | PASS |
| Operational grant → propagated | PASS |
| Operational memberships → propagated | PASS |
| Org not found → propagated (data inconsistency) | PASS |
| Operational org → propagated | PASS |
| No grant + no membership → 401 | PASS |

### 4. Current Blocker

None. AAR-003 is resolved.

### 5. Checkpoint Status

RESOLVED
