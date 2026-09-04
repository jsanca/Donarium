# DON-007Q.2-F1 — Normalize Router-Level 405 Responses

**Status:** COMPLETE
**Owner:** Deep
**Role:** Backend Fix Engineer
**Defect:** DQ2-001

## Summary

Fixed DQ2-001: Chi router's default 405 responses did not return JSON. The
handler-level 405 checks from DON-007.9 were bypassed because Chi's `Post()` and
`Get()` route registrations intercepted method validation before the handler ran.

## Root Cause

Route registrations used `router.Post("/api/setup", ...)` and `router.Get(...)`.
Chi handles method matching internally: when a path matches but the method doesn't,
it returns 405 with plain text and Allow header — before the handler is called.

Chi v5.2.1's `MethodNotAllowed(func)` has a known limitation: when a custom
handler is set, it discards the `methodsAllowed` parameter, so the custom handler
cannot build the correct Allow header. The default Chi handler sets Allow but
returns plain text.

## Solution

**Two-layer approach:**

1. **Route layer**: Changed `router.Post()`/`router.Get()` to `router.HandleFunc()`/
   `router.Handle()` in all modules. This passes all HTTP methods through to the
   handler, where method checks from DON-007.9 produce consistent JSON 405
   responses with correct Allow headers.

2. **Router layer (safety net)**: Kept `r.MethodNotAllowed(methodNotAllowedHandler)`
   as a `runtime`-level policy. Any future route registered with `Post()`/`Get()`
   that lacks a handler-level method check will still get JSON 405 (with the
   Allow header limitation).

## Implementation

| File | Change |
|---|---|
| `internal/identity/http/runtime.go` | `Post` → `HandleFunc`, `Get` → `HandleFunc` |
| `internal/platform/runtime/platform.go` | `r.Get` → `r.Handle` (2 occurrences) |
| `internal/platform/runtime/application.go` | Added `r.MethodNotAllowed(methodNotAllowedHandler)` + `ErrorResponse` type |
| `internal/platform/runtime/application_test.go` | Updated test module to use `HandleFunc` with handler-level method checks; added `405OnGetSetup` and `405OnPostStatus` tests |
| `docs/agents/reports/DON-007Q.2-F1-normalize-router-405-responses.md` | This report |
| `docs/agents/checkpoints/CHECKPOINT-DON-007Q.2-F1-normalize-router-405-responses.md` | Created |
| `docs/agents/ENGINEERING_LOG.md` | Updated |

## Test Results

```
go test ./...  →  PASS (66/66)
go vet ./...   →  PASS (0 issues)
```

QA Docker validation:

```
GET /api/setup        → 405  Allow: POST  {"error":"method not allowed"}
POST /api/setup/status → 405  Allow: GET   {"error":"method not allowed"}
POST /health/live     → 405  Allow: GET   {"error":"method not allowed"}
POST /health/ready     → 405  Allow: GET   {"error":"method not allowed"}
```

All responses include `Allow`, `Content-Type: application/json`, and JSON body
matching `ErrorResponse{"error":"method not allowed"}`.

## Tradeoffs

- **`HandleFunc` vs custom `MethodNotAllowed`**: Chose `HandleFunc` because
  Chi v5.2.1 does not expose allowed methods to custom 405 handlers. The
  `MethodNotAllowed` handler remains as a safety net for future routes.

## Follow-ups

- If a future Chi release passes `methodsAllowed` to custom handlers, the
  `MethodNotAllowed` handler can be made the primary mechanism and
  handler-level checks can be removed.
