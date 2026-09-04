# DON-003 — Server Foundation

**Status:** COMPLETE (including DON-003-F1)
**Owner:** Deep Pro
**Role:** Backend Engineer — Go
**Target:** 25–35 minutes
**Hard stop:** 45 minutes

## Objective

Create the technical skeleton of the Donarium backend and local
development environment without implementing any Go source files or
application logic.

## Scope

1. Modular monolith directory structure under `server/`
2. PostgreSQL via Docker Compose
3. Multi-stage server Dockerfile
4. Linter configuration (golangci-lint)
5. Development commands (Makefile)
6. Server README
7. Environment variable template

## Deliverables

| File | Status |
|---|---|
| `compose.yaml` | Created |
| `.env.example` | Created |
| `server/Dockerfile` | Created |
| `server/go.mod` | Created |
| `Makefile` | Created — lint uses golangci-lint v2.7.2 via Docker |
| `server/README.md` | Created |
| `server/.golangci.yml` | Created |
| `server/cmd/donarium/.gitkeep` | Created |
| `server/internal/*/.gitkeep` | Created (all domains + platform) |
| `server/migrations/.gitkeep` | Created |
| `server/tests/*/.gitkeep` | Created |
| `docs/agents/checkpoints/CHECKPOINT-DON-003-server-foundation.md` | Created |

## Validation

- [x] `docker compose config` — valid YAML
- [x] `docker compose up -d postgres` — started successfully
- [x] `docker compose ps` — healthy
- [x] `docker compose exec postgres pg_isready` — accepting connections
- [x] `docker compose down` — stopped cleanly
- [x] `make lint` — golangci-lint v2.7.2 (skips gracefully: no Go sources yet)
- [x] `make test` / `make build` — skip gracefully: no Go sources yet
- [x] No `.go` files exist under `server/`
- [x] Modular directory structure created
- [x] No `client/` modified
- [x] No commits made
