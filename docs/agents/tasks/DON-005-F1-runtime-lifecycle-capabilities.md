# DON-005-F1 — Runtime Lifecycle Capabilities

**Status:** COMPLETE
**Owner:** Deep Pro
**Role:** Backend Engineer — Go
**Parent:** DON-005

## Objective

Introduce explicit lifecycle interfaces (`Runner`, `Shutdowner`) alongside
`io.Closer`, make `ApplicationRuntime` implement them, and enforce clean
resource ownership in `main.go` — Shutdown before Close before pool.Close.

## Changes

| File | Change |
|---|---|
| `server/internal/platform/runtime/lifecycle.go` | Created — `Runner`, `Shutdowner` interfaces |
| `server/internal/platform/runtime/application.go` | `Run()` returns nil on `ErrServerClosed`, added `Close()`, compile-time assertions |
| `server/cmd/donarium/main.go` | Defer `Close()` + `pool.Close()`, errCh always receives Run result |
| `server/internal/platform/runtime/application_test.go` | `RunReturnsNilOnShutdown`, `CloseAfterShutdown`, `ImplementsLifecycle` |

## Validation

- [x] `go vet ./...` — PASS
- [x] `go test ./...` — PASS (13/13)
- [x] `golangci-lint run` — 0 issues
- [x] `go build ./cmd/donarium/` — PASS
- [x] Docker Compose + health endpoints — PASS
- [x] No commits
