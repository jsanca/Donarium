# Architecture

Current system structure, boundaries, and runtime relationships.

> Curated from repository evidence (migrations, Go packages, module wiring) as of
> 2026-09-02. This page describes **current** structure. Decision rationale lives
> in [`../adr/`](../adr/README.md); delivery history in `../engineering/`.

## Boundaries

Donarium is a modular monolith with two deployment layers and a single
PostgreSQL database.

- **`server/`** — Go 1.25 monolith (chi HTTP, pgx, argon2). Single entrypoint
  `cmd/donarium/main.go`.
- **`client/`** — React 19 SPA (Vite 7, Tailwind 4, TypeScript 5.8).
- **Database** — PostgreSQL 17.4, migrations embedded and run at startup.

Three bounded contexts are present in Go. Identity and Properties deliberately
do not import each other; each owns its own anti-corruption `DBExecutor` port.

| Bounded context | Go package root | Concern |
| --- | --- | --- |
| Identity | `server/internal/identity/` | Setup + authentication + authorization (users, credentials, organizations, memberships, platform grants) |
| Properties | `server/internal/properties/` | Property aggregate, Party/PropertyStakeholder, access rule |
| Platform (cross-cutting) | `server/internal/platform/` | config, database (pool + embedded migrations), health, runtime composition |

See the per-layer views for detail.

## Package views

- [Server packages](architecture/server-packages.md)
- [Client structure](architecture/client-structure.md)

## Database

- [Entity relationships](architecture/database-er.md)
