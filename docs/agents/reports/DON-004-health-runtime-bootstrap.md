# Implementation Report: DON-004 — Health & Runtime Bootstrap

## Context

DON-003 left a modular monolith skeleton with a single Go module,
PostgreSQL via Docker Compose, a multi-stage Dockerfile, lint config,
and a Makefile. DON-004 converts that skeleton into a minimal runnable
server: config loading, database connection, HTTP health endpoints,
and graceful shutdown.

## Summary

Implemented the first runtime of Donarium: `cmd/donarium/main.go` with
env-based configuration, pgxpool database connection, HTTP server via
`net/http` with two health endpoints (`/health/live` and `/health/ready`),
graceful shutdown on SIGINT/SIGTERM, and structured logging via `log/slog`.
All 8 unit tests pass, lint is clean (0 issues), the server builds and runs
in Docker Compose alongside PostgreSQL. No frameworks, no business logic.

## Deliverables

| File | Description |
|---|---|
| `server/cmd/donarium/main.go` | Entrypoint: config → DB → HTTP → signals → shutdown |
| `server/internal/platform/config/config.go` | Env var loading, validation, DatabaseURL/redacted |
| `server/internal/platform/database/postgres.go` | pgxpool with connect timeout and initial ping |
| `server/internal/platform/http/server.go` | net/http server with timeouts, route setup |
| `server/internal/platform/http/health/response.go` | JSON response types |
| `server/internal/platform/http/health/handler.go` | LivenessHandler, ReadinessHandler with ReadinessChecker interface |
| `server/internal/platform/http/health/handler_test.go` | 8 unit tests using fakeChecker |
| `server/go.mod` | Added pgx/v5 v5.7.4 |
| `server/go.sum` | Dependency checksums |
| `compose.yaml` | Added server service |
| `server/Dockerfile` | Buildable, HEALTHCHECK fixed |
| `.env.example` | Added HTTP_PORT, POSTGRES_SSLMODE, timeouts |
| `Makefile` | Added server-up/down/build/run, health-live/ready |

## Architectural Decisions

1. **net/http over frameworks.** With only two health endpoints, the standard
   library is sufficient. No router framework (Gin, Echo, Chi, Fiber) was
   introduced.

2. **Go 1.22+ pattern-based routing.** `mux.HandleFunc("GET /health/live", ...)`
   provides method-specific routing without third-party dependencies. Non-GET
   requests return 405 with an Allow header.

3. **ReadinessChecker interface.** `type ReadinessChecker interface { Check(ctx context.Context) error }`.
   Kept intentionally minimal — one method, context-aware. The PostgreSQL
   implementation delegates to `pool.Ping(ctx)`. No extensible health registry
   was built.

4. **DatabaseURL construction.** Built from individual config fields rather
   than a raw connection string. A `DatabaseURLRedacted()` variant is used
   for logging to avoid leaking credentials.

5. **Config via os.Getenv.** No heavy config library. A single struct with
   defaults, required-field validation, and duration parsing. No global
   mutable state.

6. **Graceful shutdown flow.** SIGINT/SIGTERM → stop accepting traffic →
   HTTP shutdown with 30s timeout → close DB pool → exit. `os.Exit` is only
   called from `main`; internal packages return errors.

## Implementation Notes

- The `.health` subdirectory was not created as separate Go files were placed
  directly in `health/` — this matched the task's expected structure.
- `.gitkeep` files were removed from directories that now contain Go source.
- golangci-lint flagged unchecked `json.Encode` returns — resolved with `_ =`.
- A pre-existing Java process on port 8080 was stopped before Docker Compose
  validation to avoid port conflicts.

## Validation

| Check | Result |
|---|---|
| `go vet ./...` | PASS |
| `go test ./...` | PASS (8/8) |
| `golangci-lint run` (v2.7.2) | PASS (0 issues) |
| `go build ./cmd/donarium/` | PASS |
| `docker compose config` | PASS |
| `docker compose up -d --build` | PASS — both services healthy |
| `/health/live` (DB up) | PASS — 200 `{"status":"ok"}` |
| `/health/ready` (DB up) | PASS — 200 `{"status":"ready","checks":{"database":"up"}}` |
| `/health/live` (DB down) | PASS — 200 |
| `/health/ready` (DB down) | PASS — 503 `{"status":"not_ready","checks":{"database":"down"}}` |
| No secrets in responses or logs | PASS |
| `client/` not modified | PASS |

## Tests

8 unit tests in `server/internal/platform/http/health/handler_test.go`:

| Test | Verifies |
|---|---|
| `TestLivenessHandler_Returns200` | Status 200, JSON body `{"status":"ok"}` |
| `TestLivenessHandler_ReturnsJSONContentType` | Content-Type: application/json |
| `TestLivenessHandler_RejectsPOST` | POST returns 405 |
| `TestReadinessHandler_ReadyWhenHealthy` | checker OK → 200, status "ready", database "up" |
| `TestReadinessHandler_NotReadyWhenUnhealthy` | checker error → 503, status "not_ready", database "down" |
| `TestReadinessHandler_ReturnsJSONContentType` | Content-Type: application/json |
| `TestReadinessHandler_RejectsPOST` | POST returns 405 |

All tests use a `fakeChecker` struct implementing `ReadinessChecker`. No
mock framework or real PostgreSQL dependency.

## Tradeoffs

- **ReadinessChecker single method vs. multi-check registry.** One method
  is sufficient for a database ping. Extensibility deferred to when
  additional dependencies (Redis, etc.) are added.
- **pgxpool vs. database/sql.** pgxpool is PostgreSQL-specific and more
  performant than database/sql + pgx driver. The tradeoff is that switching
  databases later would require changing the driver.
- **No middleware layer yet.** CORS, request ID, logging middleware are
  deferred. The task scope is explicitly health endpoints only.

## Open Questions

- When should the health check registry become extensible (multiple
  checkers, per-check timeouts, aggregated status)?
- Should the server listen on a different port inside Docker vs. local
  development (e.g., `SERVER_PORT` vs `HTTP_PORT`)?

## Follow-ups

- **DON-005:** Identity domain — user model, authentication primitives
- Add middleware: request logging, request ID, CORS
- Add server integration test against real PostgreSQL via testcontainers
- Consider `tools.go` for golangci-lint when Go sources already exist

## References

- Task: `docs/agents/tasks/DON-004—Health &RuntimeBootstrap.md`
- Checkpoint: `docs/agents/checkpoints/CHECKPOINT-DON-004-health-runtime-bootstrap.md`
- DON-003 report: `docs/agents/reports/DON-003-server-foundation.md`
- ROADMAP.md — Phase 2 (Authentication Experience)
