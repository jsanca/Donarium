# Server

Donarium backend — a modular monolith in Go, backed by PostgreSQL.

## Architecture decision

The server is built as a **single deployable** with internal modules
separated by domain boundaries. Each domain lives under `internal/`
as a plain Go package. There is a single `go.mod` at the server root.

This avoids premature distribution into independent Go modules while
keeping the code organized for future extraction if needed.

## Directory layout

```
server/
├── cmd/donarium/              # Application entrypoint
├── internal/
│   ├── identity/              # Authentication and accounts
│   ├── organizations/         # Multi-tenant workspaces
│   ├── properties/            # Property and unit management
│   ├── leases/                # Lease agreements
│   ├── payments/              # Payment obligations and records
│   ├── maintenance/           # Maintenance requests and tracking
│   ├── documents/             # Document storage and retrieval
│   ├── notifications/         # User notifications
│   └── platform/              # Shared infrastructure
│       ├── config/
│       ├── database/
│       ├── http/
│       └── observability/
├── migrations/                # Database migrations
├── tests/
│   ├── integration/
│   └── fixtures/
├── Dockerfile                 # Multi-stage Go build
├── go.mod                     # Single module (no sub-modules)
├── Makefile                   # (commands are at repository root)
├── .golangci.yml              # Linter configuration
└── README.md                  # This file
```

## Prerequisites

- Go 1.25+
- Docker and Docker Compose
- golangci-lint (version controlled in CI, not yet pinned locally)

## Getting started

Start PostgreSQL:

```sh
make postgres-up
```

Check status:

```sh
docker compose ps
docker compose exec postgres pg_isready
```

Stop:

```sh
make postgres-down
```

## Commands

All commands are run from the repository root via `make`:

| Command | Description |
|---|---|
| `make postgres-up` | Start PostgreSQL |
| `make postgres-down` | Stop PostgreSQL |
| `make postgres-logs` | Tail PostgreSQL logs |
| `make lint` | Run golangci-lint |
| `make test` | Run tests |
| `make build` | Build server binary |
| `make help` | Show available commands |

## Current status

This is the **server skeleton** — the structural foundation for the Go
monolith. No Go source files exist yet. Every directory is prepared for
the domain it will host. PostgreSQL runs locally via Docker Compose.

## What comes next

- `cmd/donarium/main.go` — application entrypoint with HTTP server
- `internal/platform/config/` — configuration loading
- `internal/platform/database/` — PostgreSQL connection and pooling
- `internal/platform/http/` — router and middleware setup
- Domain modules — identity, properties, leases, etc.
- `migrations/` — SQL migration files
- Tests — integration and unit
