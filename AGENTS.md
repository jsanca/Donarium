# Donarium Working Agreement

Shared understanding for contributors and agents working in this repository.

## Stack

- **Backend**: Go 1.25 monolith at `server/`. Single `go.mod`. Uses chi (HTTP), pgx (PostgreSQL), golang.org/x/crypto (argon2).
- **Frontend**: React 19 SPA at `client/`. Vite 7, Tailwind CSS 4, TypeScript 5.8, react-hook-form + zod, i18next, motion.
- **Infrastructure**: PostgreSQL 17.4 via Docker Compose (`compose.yaml`).

## Architecture

- `server/cmd/donarium/main.go` is the single entrypoint. It loads env config, connects PostgreSQL, runs migrations, wires all dependencies, and starts the HTTP server.
- The server is a modular monolith. Active code lives in two internal packages:
  - **`identity/`** — users, credentials, organizations, memberships, platform grants (domain types at root; use cases in `application/`; handlers in `http/`; pgx persistence in `pgx/`)
  - **`platform/`** — cross-cutting: `config/`, `database/` (pool + embedded migrations), `http/health/`, `runtime/` (app lifecycle)
- Other domain dirs (`organizations/`, `properties/`, `leases/`, etc.) are `.gitkeep` placeholders awaiting their vertical slices.
- Migrations are **embedded Go files** at `server/internal/platform/database/migrations/` and run automatically on startup. Five migration pairs exist (users, credentials, organizations, memberships, platform_grants). All SQL migrations use `.up.sql`/`.down.sql` naming.
- The client is organized by feature under `app/features/`. i18n is active with Spanish translations at `shared/i18n/es.ts`.

## Commands

All from repository root unless noted:

### Local dev

| Command | What it does |
|---|---|
| `make postgres-up` | Start PostgreSQL in Docker |
| `make postgres-down` | Stop PostgreSQL and server |
| `make server-run` | Run server locally via `go run` (needs PostgreSQL and `.env`) |
| `make server-build` | Compile server binary to `server/donarium` |
| `make server-up` | Start full stack (PostgreSQL + server) via Docker Compose |
| `make lint` | Run golangci-lint v2 via pinned Docker image (`.golangci.yml` at `server/`) |
| `make test` | Run `go test ./...` from `server/` |
| `make health-live` / `make health-ready` | Hit `/health/live` and `/health/ready` on local server |

### QA environment

| Command | What it does |
|---|---|
| `make qa-up` | Start QA stack on separate ports (PG:15432, API:18080) via `compose.qa.yml` |
| `make qa-down` | Stop QA environment (preserves data) |
| `make qa-reset` | Destroy QA data and recreate clean |
| `make qa-status` / `make qa-logs` | Check QA health / tail QA logs |

### Client (`client/`)

| Command | What it does |
|---|---|
| `npm run dev` | Vite dev server |
| `npm run build` | `tsc -b && vite build` (typecheck then bundle) |

No standalone `npm run lint` or `npm run typecheck` exists. Type checking is part of `npm run build`.

## Environment and setup

- **`.env` is required** for the server. Copy `.env.example` to `.env`. Default values work for local dev.
- Key env vars: `DONARIUM_ENV` (local/qa/staging/production, default `local`), `HTTP_PORT` (default `8080`), `POSTGRES_*`, `SESSION_SIGNING_KEY` (required in staging/prod, auto-derives dev key for local/qa), `SESSION_TTL` (default `24h`).
- **`POSTGRES_PASSWORD` is required** — the server refuses to start without it.
- `SESSION_SIGNING_KEY` must be ≥32 chars in staging/production and must NOT use the dev default.
- All server env vars are read via `config.Load()` at startup — no dotenv library is used; the OS environment is the source of truth.

## Testing

- **All Go tests require PostgreSQL running.** Run `make postgres-up` first.
- Test packages use `TEST_DATABASE_URL` env var (defaults to `postgres://donarium:donarium@localhost:5432/donarium?sslmode=disable`).
- Tests that fail to connect to PostgreSQL **silently exit 0** (`os.Exit(0)` in `TestMain`) — they do not fail the run. Check that PostgreSQL is up before trusting test results.
- Integration tests live in `server/tests/integration/`. Unit tests are co-located with their packages (e.g. `server/internal/identity/pgx/repository_test.go`).
- QA environment uses separate DB (`donarium_qa`) on port 15432 — tests against QA can target it with `TEST_DATABASE_URL=postgres://donarium_qa:donarium_qa@localhost:15432/donarium_qa?sslmode=disable`.

## What to ignore

- **`settings.local.json`** — permissions from the Arbitrier project. Does not describe Donarium.
- **`.agents/skills/`, `.claude/skills/`, `.codex/skills/`** — Java-focused skills shared from another project. Do not load them. The stack is Go + TypeScript.
- **`.gitignore`** is a Go template. It also covers `client/` build output. No Go workspace files (`go.work`) are in use.

## Design and product decisions

- **Design constitution is binding**: `knowledge/design/DonariumDesignConstitution.md`. Avoid generic admin dashboards, excessive tables, corporate blue, neon gradients, glassmorphism.
- Use cases: `knowledge/use-cases/`. Roadmap: `ROADMAP.md`.
- Engineering records: `docs/agents/` — task definitions, reports, and checkpoints from prior slices.

## Working philosophy

- **UX before code.** Start from what a person needs to understand or achieve.
- **Domain before infrastructure.** Establish the problem and its language before selecting technical machinery.
- **Vertical slices.** Deliver complete, observable outcomes instead of disconnected layers.
- **Documentation first.** Record intent and constraints before implementation.
- **Small iterations.** Prefer the smallest useful increment, then learn from it.
- **No speculative architecture.** Add structure only when a real need demands it.
- **Accessibility** and **internationalization** are baseline, not afterthoughts.

<!-- OSK:BEGIN -->

## OSK Workspace

Read:

- `docs/PROJECT.md`
- `docs/OSK.md`

<!-- OSK:END -->
