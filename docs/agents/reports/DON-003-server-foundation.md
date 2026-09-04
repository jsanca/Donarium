# Implementation Report: DON-003 — Server Foundation

## Context

Donarium is entering implementation. The backend will be a modular Go
monolith backed by PostgreSQL, with the frontend (React/Vite) built in
parallel under `client/`. This task establishes the skeleton before any
application logic exists.

## Summary

Created the full technical skeleton for the Donarium Go backend and
local development environment. All directories are in place, PostgreSQL
runs via Docker Compose with health checks and persistent storage, the
Dockerfile defines the multi-stage build strategy, golangci-lint is
configured, and dev commands work from the root Makefile. No Go source
files were created.

## Deliverables

| File | Description |
|---|---|
| `compose.yaml` | PostgreSQL 17.4-alpine, health check, persistent volume, internal network |
| `.env.example` | Local database credentials |
| `Makefile` | `postgres-up/down/logs`, `lint` (golangci-lint v2.7.2 via Docker), `test`, `build`, `help` |
| `server/go.mod` | Single module `donarium/server`, Go 1.25 |
| `server/Dockerfile` | Multi-stage: Go 1.25 build + Alpine 3.22 runtime, non-root user |
| `server/.golangci.yml` | v2: errcheck, govet, ineffassign, staticcheck, unused + gofmt/goimports |
| `server/README.md` | Architecture decision, layout, commands, status |
| `server/cmd/donarium/.gitkeep` | Entrypoint placeholder |
| `server/internal/{9 domains + 4 platform}/.gitkeep` | Domain boundaries |
| `server/migrations/.gitkeep` | Future SQL migrations |
| `server/tests/{integration,fixtures}/.gitkeep` | Test structure |
| `docs/agents/tasks/DON-003-server-foundation.md` | Task record |
| `docs/agents/reports/DON-003-server-foundation.md` | This report |
| `docs/agents/checkpoints/CHECKPOINT-DON-003-server-foundation.md` | Checkpoint (RESOLVED) |

## Architectural Decisions

1. **Single module at `server/go.mod`.** No sub-modules per domain.
   Package boundaries express domain separation; `go.mod` boundaries
   are deferred until a real need emerges.

2. **Makefile at repository root.** The `compose.yaml` is at root,
   and commands like `postgres-up` naturally operate there. Server
   commands (`lint`, `test`, `build`) delegate to `server/`. This
   keeps a single entry point for developers.

3. **Go 1.25 baseline.** Matches the Go version available in the
   environment. Dockerfile uses the same version explicitly.

4. **PostgreSQL 17.4-alpine.** Explicitly pinned version with Alpine
   base for small image size.

5. **Module name `donarium/server`.** Neutral — no remote organization
   is defined yet. Follows the repo name convention. Documented as a
   conscious choice.

6. **No `.go` files.** `go.mod` is metadata, not source. This keeps
   the skeleton clean for the first implementation slice.

7. **golangci-lint v2.7.2 pinned via Docker.** No system dependency on
   golangci-lint. The Makefile runs it through a versioned Docker image,
   avoiding `tools.go` (which would introduce a `.go` file).

## Implementation Notes

- The `.gitignore` already contained `.env` — no update was needed.
- PostgreSQL container was started, verified healthy (`pg_isready`),
  and torn down. No containers remain running.
- `make lint`, `make test`, `make build` detect the absence of Go
  source files and skip gracefully rather than failing. When Go sources
  exist, `make lint` runs golangci-lint v2.7.2 via Docker.
- The Dockerfile documents it is "not buildable yet" — no dummy
  `main.go` was created.
- `client/` was not modified. It already exists as a React/Vite project
  with its own `node_modules/` and build tooling.

## Validation

| Check | Result |
|---|---|
| `docker compose config` | PASS — valid YAML |
| `docker compose up -d postgres` | PASS — started |
| `docker compose ps` | PASS — healthy |
| `docker compose exec postgres pg_isready` | PASS — accepting connections |
| `docker compose down` | PASS — stopped and removed |
| `find server -name '*.go'` | PASS — no output (zero `.go` files) |
| `make help` | PASS — all commands listed |
| `make lint` | PASS — "No Go source files yet. Skipping." |
| `make test` | PASS — "No Go source files yet. Skipping." |
| `make build` | PASS — "No main.go yet. Skipping." |
| Secrets in `.env.example` | PASS — only local development defaults |
| `client/` modified | PASS — no changes |

## Tests

No automated tests exist. The task explicitly excludes Go source files.
Testing infrastructure directories are prepared under `server/tests/`.

## Tradeoffs

- **Root Makefile vs. server/Makefile.** Root was chosen because
  `compose.yaml` lives at root. If the compose configuration moves
  under `server/`, the Makefile should move with it.
- **golangci-lint via Docker vs. system install vs. `tools.go`.** Docker
  was chosen to pin the version without introducing a `.go` file.
  `tools.go` will be considered once Go sources exist.
- **No `.env` file created.** Only `.env.example` exists. Developers
  copy it to `.env` locally; `.env` is in `.gitignore`.

## Open Questions

- Should the `compose.yaml` eventually live under `server/` or stay at
  root when the server binary is added as a service?

## Follow-ups

- **DON-004:** `cmd/donarium/main.go`, HTTP server scaffold, platform
  config loading, database connection.

## References

- Task: `docs/agents/tasks/DON-003-server-foundation.md`
- Checkpoint: `docs/agents/checkpoints/CHECKPOINT-DON-003-server-foundation.md`
- ROADMAP.md — Phase 2 (Authentication Experience) is the next domain target
- Design constitution: `knowledge/design/DonariumDesignConstitution.md`
