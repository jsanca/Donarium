# Implementation Report: DON-005 — Runtime Composition with Chi

## Context

DON-004 left a working runtime with config loading, pgxpool, `log/slog`,
and two health endpoints on `http.ServeMux`. The architectural direction
for Donarium is explicit runtime composition: a contract for modular
route registration, an application runtime that composes module runtimes,
and a composition root that wires dependencies without touching routes.

## Summary

Refactored the server bootstrap to use `go-chi/chi/v5` as the router,
introduced a `ModuleRuntime` interface with `RegisterRoutes(chi.Router)`,
created `ApplicationRuntime` to own the Chi router and HTTP server
lifecycle, and extracted health routes into `PlatformRuntime`. `main.go`
is now a pure composition root — no route registration, no mux creation,
no server construction details. Health behavior is unchanged (verified
via Docker Compose). 11/11 tests pass, lint and vet clean.

## Deliverables

| File | Description |
|---|---|
| `server/internal/platform/runtime/module.go` | `ModuleRuntime` interface |
| `server/internal/platform/runtime/application.go` | `ApplicationRuntime` — Chi router, server lifecycle, module composition |
| `server/internal/platform/runtime/platform.go` | `PlatformRuntime` — health route registration |
| `server/internal/platform/runtime/application_test.go` | 4 composition tests (fake modules, httptest) |
| `server/cmd/donarium/main.go` | Refactored — composition root only |
| `server/internal/platform/http/server.go` | Removed (logic in ApplicationRuntime) |
| `server/go.mod` | Added `github.com/go-chi/chi/v5 v5.2.1` |

## Architectural Decisions

1. **Chi over http.ServeMux.** The standard library router works for trivial
   cases, but Chi provides better middleware composition, route grouping,
   and mountable sub-routers needed for future domain modules.

2. **ModuleRuntime interface on chi.Router.** The contract receives the full
   `chi.Router` interface, giving each module access to `Get`, `Post`,
   `Route`, `Mount`, and `Use`. The interface is intentionally minimal —
   just `RegisterRoutes`. No lifecycle hooks (Start/Stop/Init) were added
   since they aren't needed yet.

3. **ApplicationRuntime owns the server lifecycle.** `NewApplication` creates
   the Chi router and registers modules. `Run()` creates the `http.Server`
   (with timeouts) and listens. `Shutdown(ctx)` performs graceful HTTP
   shutdown. The pool is closed by `main.go` via `defer`, maintaining single
   ownership.

4. **PlatformRuntime receives pool, not checker.** PlatformRuntime accepts
   `*pgxpool.Pool` and creates its own unexported `poolChecker`. This keeps
   the main composition root simple (one dependency to pass) while allowing
   PlatformRuntime to construct the right checker internally.

5. **No middleware yet.** Chi is used bare. Logging, recovery, request ID,
   and CORS middleware are deferred to when needed. This avoids changing
   observable behavior beyond the router swap.

6. **No domain runtimes created.** The task explicitly avoids placeholder
   files for identity, properties, leases, etc. The contract is ready;
   each runtime will be created with its first vertical slice.

## Implementation Notes

- `server.go` was removed entirely — the DBReadinessChecker became
  `poolChecker` in `platform.go`. The Run/Shutdown logic moved to
  `ApplicationRuntime` methods.
- The `Handler()` method on ApplicationRuntime exposes the Chi router as
  `http.Handler` for testing without starting a real TCP server.
- `TestApplicationRuntime_RunCreatesServer` uses port `:0` for an
  ephemeral port, then calls Shutdown immediately to verify the lifecycle.

## Validation

| Check | Result |
|---|---|
| `go vet ./...` | PASS |
| `go test ./...` | PASS (11/11: 7 health + 4 composition) |
| `golangci-lint run` (v2.7.2) | PASS (0 issues) |
| `go build ./cmd/donarium/` | PASS |
| `docker compose config` | PASS |
| `docker compose up -d --build` | PASS |
| `/health/live` (DB up) | PASS — 200 `{"status":"ok"}` |
| `/health/ready` (DB up) | PASS — 200 `{"status":"ready","checks":{"database":"up"}}` |
| `/health/live` (DB down) | PASS — 200 |
| `/health/ready` (DB down) | PASS — 503 `{"status":"not_ready","checks":{"database":"down"}}` |
| `docker compose down` | PASS — clean |

## Tests

### Existing (preserved): 7 health handler tests

All 7 handler tests passed without modification — the health handler
functions (`LivenessHandler`, `ReadinessHandler`) were unaffected.

### New: 4 composition tests

| Test | Verifies |
|---|---|
| `TestApplicationRuntime_RegistersModuleRoutes` | Module receives route registration, registered route responds 200 via httptest |
| `TestApplicationRuntime_MultipleModules` | Both modules receive registration |
| `TestApplicationRuntime_NoModules` | Empty application handles unknown routes (404/405) |
| `TestApplicationRuntime_RunCreatesServer` | Run/Shutdown lifecycle works with ephemeral port |

## Tradeoffs

- **chi.Router in the contract signature.** This couples every module to
  Chi. If the router were ever swapped, every module would change. For a
  monolith, this is acceptable — router changes are extremely rare.
- **PlatformRuntime creates its own checker.** An alternative would be
  passing `health.ReadinessChecker` directly and letting main create it.
  The current approach keeps main's surface area smaller at the cost of
  PlatformRuntime coupling to pgxpool. If additional checkers are added
  (Redis, etc.), PlatformRuntime should accept a slice of checkers.

## Open Questions

- When should middleware (request logging, request ID, recovery) be
  introduced? With the first domain slice or as a separate platform task?

## Follow-ups

- **DON-006:** First domain slice (Identity) using the ModuleRuntime contract
- Add Chi middleware for request logging and recovery
- Consider mounting health under `/health` sub-router for clarity

## DON-005-F1 — Runtime Lifecycle Capabilities

### Context

The DON-005 base gave us a working composition model but lacked explicit
lifecycle contracts. ApplicationRuntime had `Run()` and `Shutdown(ctx)` as
concrete methods with no interface backing. Resource cleanup ordering in
main.go was implicit — Shutdown was called but Close was never invoked on
the runtime, and the errCh only forwarded non-nil errors.

### Changes

1. **Lifecycle interfaces.** `Runner` and `Shutdowner` in `lifecycle.go`.
   `io.Closer` from the standard library (not duplicated).

2. **Run returns nil on graceful shutdown.** Uses `errors.Is(err, http.ErrServerClosed)`
   to treat server close as clean termination. The errCh in main always receives
   the Run result — nil signals clean exit.

3. **Close method.** Calls `server.Close()`. Nil-safe (checks `a.server != nil`).
   Safe to call after Shutdown. Does not touch the database pool.

4. **Compile-time assertions.** Three assertions verify interface compliance at
   build time — if any method signature changes, compilation fails.

5. **main.go cleanup ordering.** Defers registered in creation order:
   `defer pool.Close()` then `defer appRuntime.Close()`. On exit, LIFO runs
   Close() before pool.Close(). Shutdown is called explicitly before the
   function returns, ensuring graceful drain before hard close.

6. **Tests.** `RunReturnsNilOnShutdown` verifies Run returns nil after graceful
   shutdown. `CloseAfterShutdown` confirms Close is safe post-Shutdown.
   `ImplementsLifecycle` exercises all three interface casts.

### Architectural Decisions

- **Why io.Closer instead of a custom Closer.** The standard library already
  defines `Close() error`. Duplicating it adds nothing and creates confusion
  about which Closer a type implements.
- **No lifecycle hooks on ModuleRuntime.** Only runtimes that own active
  resources (workers, consumers, schedulers) need lifecycle. Health routes
  are stateless. Adding empty methods to every module would violate the
  interface segregation principle.
- **Run always sends to channel.** The previous code only forwarded non-nil
  errors, meaning a clean shutdown (nil) was silently dropped. Now main
  observes all terminations and logs appropriately.

## References

- Task: `docs/agents/tasks/DON-005—RuntimeCompositionWithChi.md`
- Checkpoint: `docs/agents/checkpoints/CHECKPOINT-DON-005-runtime-composition-with-chi.md`
- DON-004 report: `docs/agents/reports/DON-004-health-runtime-bootstrap.md`
- ROADMAP.md — Phase 2 (Authentication Experience)
