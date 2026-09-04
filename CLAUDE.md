# Claude Workspace Guide

Start with [docs/PROJECT.md](docs/PROJECT.md) to understand the project, then read [docs/OSK.md](docs/OSK.md) before creating or changing workspace documentation.

Use canonical project documentation as the authority. Do not place shared project knowledge exclusively in this file.

<!-- OSK:BEGIN -->

## OSK Workspace

Read:

- `docs/PROJECT.md`
- `docs/OSK.md`

<!-- OSK:END -->

## Tool-Specific Instructions

Add only instructions required by Claude here. Keep shared project guidance in `docs/`.

### Operational source of truth

`AGENTS.md` at the repo root is the working agreement. Read it before running commands or reasoning about architecture — it lists the stack, all `make` targets, env vars, test rules, and directories to ignore. Do not re-derive that information from files here.

### Gotchas that will bite you if ignored

- **Go tests silently exit 0 when PostgreSQL is not reachable.** A green `make test` does not prove the tests ran. Always `make postgres-up` first, or check test output for actual test names.
- **`.env` and `POSTGRES_PASSWORD` are required** for the server to boot; there is no dotenv loader — the OS environment is authoritative. Copy `.env.example` before `make server-run`.
- **Migrations are embedded Go files** under `server/internal/platform/database/migrations/` and run on startup. New migrations must ship as paired `.up.sql`/`.down.sql`.
- **Ignore these directories** — they are shared from unrelated projects and describe a different stack (Java) or unrelated permissions: `.agents/skills/`, `.claude/skills/`, `.codex/skills/`, `settings.local.json`.
- **Design constitution is binding**: `docs/knowledge/design/DonariumDesignConstitution.md`. No generic admin dashboards, corporate blue, neon gradients, or glassmorphism.

### Common commands (quick reference — see `AGENTS.md` for the full list)

- `make postgres-up` → start PG (required before tests or `server-run`)
- `make server-run` → run Go server locally
- `make test` → `go test ./...` in `server/` (needs PG)
- `make lint` → golangci-lint v2 via pinned Docker image
- Client: `cd client && npm run dev` / `npm run build` (build also typechecks; no separate `lint`/`typecheck` script)
- Run a single Go test: `cd server && go test ./internal/identity/pgx/ -run TestName`
