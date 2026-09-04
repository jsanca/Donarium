## Recovery Checkpoint

### 1. Original Objective

DON-007Q.2-F1: Fix defect DQ2-001 — Chi router 405 responses must return JSON
(`Content-Type: application/json`, `{"error":"method not allowed"}`) instead of
plain text. Both `GET /api/setup` and `POST /api/setup/status` affected.

### 2. Root Cause

Route registrations used `router.Post()` and `router.Get()`, which let Chi
handle method validation internally. Chi's default 405 returns plain text.
Handler-level 405 checks from DON-007.9 were never reached.

### 3. Completed Work

| File | Change |
|---|---|
| `internal/identity/http/runtime.go` | `Post` → `HandleFunc`, `Get` → `HandleFunc` |
| `internal/platform/runtime/platform.go` | `r.Get` → `r.Handle` (health routes) |
| `internal/platform/runtime/application.go` | Added `r.MethodNotAllowed(handler)` + `ErrorResponse` type |
| `internal/platform/runtime/application_test.go` | Updated test module + 2 new 405 tests |

### 4. Validation Status

| Check | Result |
|---|---|
| `go vet ./...` | PASS |
| `go test ./...` | PASS (66/66) |
| `GET /api/setup` (QA Docker) | 405, Allow: POST, JSON |
| `POST /api/setup/status` (QA Docker) | 405, Allow: GET, JSON |
| `POST /health/live` (QA Docker) | 405, Allow: GET, JSON |
| `POST /health/ready` (QA Docker) | 405, Allow: GET, JSON |

### 5. Current Blocker

None. DQ2-001 is resolved.

### 6. Remaining Work

None.

### 7. Checkpoint Status

RESOLVED
