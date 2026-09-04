## Recovery Checkpoint

### 1. Original Objective

Implement the first runtime of the Donarium Go backend: `main.go` with config
loading, PostgreSQL connection via pgxpool, HTTP server with `/health/live`
and `/health/ready` endpoints, graceful shutdown, and slog logging. Also
update Docker Compose, Dockerfile, and Makefile for the runnable server.

### 2. Completed Work

- `server/cmd/donarium/main.go` — entrypoint with config load, DB connect,
  HTTP server start, signal handling, graceful shutdown
- `server/internal/platform/config/config.go` — env var loading with defaults,
  validation, DatabaseURL helper (with redacted variant for logging)
- `server/internal/platform/database/postgres.go` — pgxpool connection with
  timeout and ping
- `server/internal/platform/http/server.go` — net/http server with timeouts
  (ReadHeaderTimeout, ReadTimeout, WriteTimeout, IdleTimeout), route
  registration, Run/Shutdown helpers
- `server/internal/platform/http/health/response.go` — LivenessResponse,
  ReadinessResponse types
- `server/internal/platform/http/health/handler.go` — LivenessHandler (200),
  ReadinessHandler (200/503 based on ReadinessChecker interface), method
  enforcement, JSON Content-Type
- `server/internal/platform/http/health/handler_test.go` — 8 unit tests: liveness
  (200, JSON, rejects POST), readiness (ready 200, not_ready 503, JSON,
  rejects POST). Uses fakeChecker implementing ReadinessChecker.
- `server/go.mod` — added `github.com/jackc/pgx/v5 v5.7.4`
- `server/go.sum` — created via `go mod tidy`
- `compose.yaml` — added `server` service (builds from server/Dockerfile,
  depends on postgres healthy, env vars, internal network)
- `server/Dockerfile` — removed "not buildable" note, fixed HEALTHCHECK to
  use `/health/live`, port via env var
- `.env.example` — added HTTP_PORT, POSTGRES_SSLMODE, DATABASE_CONNECT_TIMEOUT,
  SHUTDOWN_TIMEOUT
- `Makefile` — added server-build, server-run, server-up, server-down,
  server-logs, health-live, health-ready; lint/test/build now work with Go sources

### 3. Files Changed

| File | Change |
|---|---|
| `server/cmd/donarium/main.go` | Created — entrypoint |
| `server/internal/platform/config/config.go` | Created — env config loading |
| `server/internal/platform/database/postgres.go` | Created — pgxpool connection |
| `server/internal/platform/http/server.go` | Created — HTTP server setup |
| `server/internal/platform/http/health/response.go` | Created — response types |
| `server/internal/platform/http/health/handler.go` | Created — health handlers |
| `server/internal/platform/http/health/handler_test.go` | Created — 8 unit tests |
| `server/go.mod` | Updated — pgx/v5 dependency |
| `server/go.sum` | Created — dependency checksums |
| `compose.yaml` | Updated — added server service |
| `server/Dockerfile` | Updated — buildable, fixed HEALTHCHECK |
| `.env.example` | Updated — added server env vars |
| `Makefile` | Updated — server and health commands |
| `server/*/.gitkeep` | Removed from dirs now containing Go files |

### 4. Current Repository State

- Server compiles, builds, and runs
- 8/8 health handler tests pass
- Lint (golangci-lint v2.7.2) reports 0 issues
- Docker Compose builds and runs both postgres and server
- /health/live returns 200 with `{"status":"ok"}`
- /health/ready returns 200 with `{"status":"ready","checks":{"database":"up"}}` when DB is up
- /health/ready returns 503 with `{"status":"not_ready","checks":{"database":"down"}}` when DB is down
- No business logic, no migrations, no frameworks
- Safe to continue

### 5. Validation Status

| Check | Result |
|---|---|
| `go fmt ./...` | PASS |
| `go vet ./...` | PASS |
| `go test ./...` | PASS (8/8) |
| `golangci-lint run` | PASS (0 issues) |
| `go build ./cmd/donarium/` | PASS |
| `docker compose config` | PASS |
| `docker compose up -d --build` | PASS |
| `curl /health/live` (DB up) | PASS — 200 `{"status":"ok"}` |
| `curl /health/ready` (DB up) | PASS — 200 `{"status":"ready","checks":{"database":"up"}}` |
| `docker compose stop postgres` | PASS |
| `curl /health/live` (DB down) | PASS — 200 `{"status":"ok"}` |
| `curl /health/ready` (DB down) | PASS — 503 `{"status":"not_ready","checks":{"database":"down"}}` |
| `docker compose down` | PASS — clean |
| No Go container running | PASS |
| `client/` not modified | PASS |

### 6. Current Blocker

None. All planned work is complete.

### 7. Evidence

```
$ go test ./...        → ok (8/8)
$ golangci-lint run     → 0 issues
$ go build ./cmd/donarium/ → ok
$ docker compose up -d  → both services healthy
$ curl localhost:8080/health/live  → 200 {"status":"ok"}
$ curl localhost:8080/health/ready → 200 {"status":"ready","checks":{"database":"up"}}
$ docker compose stop postgres
$ curl localhost:8080/health/live  → 200 {"status":"ok"}
$ curl localhost:8080/health/ready → 503 {"status":"not_ready","checks":{"database":"down"}}
```

### 8. Remaining Work

None — task is complete.

### 9. Proposed Continuation Tasks

- **DON-005: Identity domain** — user model, authentication, login endpoint
- **DON-006: SQL migrations** — Flyway or golang-migrate, initial schema
- **DON-007: Integration tests** — DB-dependent tests with testcontainers

### 10. Recommended Next Action

Mark task complete. Proceed to DON-005 (Identity domain).

### 11. Checkpoint Status

RESOLVED
