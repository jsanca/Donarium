# QA Environment

Isolated, reproducible environment for Donarium functional and integration testing.
Uses real PostgreSQL and the real backend — no mocks, no shared development data.

## Purpose

The QA environment exists exclusively for running functional test cases (TC-001 and
future slices) against a clean, disposable stack. It must never be used against
production data, development databases, or external services.

## Security

- **No real secrets.** All credentials are hardcoded local defaults (`donarium_qa` /
  `donarium_qa`). They are safe to destroy.
- **No shared volumes.** QA PostgreSQL uses its own Docker volume
  (`donarium-postgres-qa-data`). It never mounts the development volume.
- **Port isolation.** QA ports (18080, 15432) differ from development ports (8080, 5432).
- **`qa-reset` destroys data.** `docker compose -f compose.qa.yml down -v` removes
  the QA volume entirely. Verify container names before running:
  ```sh
  docker compose -f compose.qa.yml ps
  ```
- **Never run cleanup against an external URL.** `qa-reset` only touches local
  Docker resources defined in `compose.qa.yml`.

## Quick start

```sh
# Full clean start (destroys any existing QA data)
make qa-reset

# Check that everything is running
make qa-status
```

Expected output from `qa-status`:

```
=== QA Services ===
NAME                  IMAGE                STATUS
donarium-postgres-qa  postgres:17.4-alpine healthy
donarium-api-qa       compose.qa.yml       healthy

=== Health Checks ===
{"status":"ok"}
{"status":"ready","checks":{"database":"up"}}
```

## Endpoints

| Service | URL |
|---|---|
| Donarium API | http://127.0.0.1:18080 |
| PostgreSQL | 127.0.0.1:15432 |
| Health — liveness | GET http://127.0.0.1:18080/health/live |
| Health — readiness | GET http://127.0.0.1:18080/health/ready |
| Setup status | GET http://127.0.0.1:18080/api/setup/status |

## Verifying a clean environment

```sh
curl -s http://127.0.0.1:18080/api/setup/status
# {"initialized":false}
```

After running setup:

```sh
curl -s -X POST http://127.0.0.1:18080/api/setup \
  -H "Content-Type: application/json" \
  -d '{"displayName":"Owner","email":"owner@qa.test","password":"ValidP@ss1","organizationName":"QA Org","organizationSlug":"qa-org"}'
```

Then verify:

```sh
curl -s http://127.0.0.1:18080/api/setup/status
# {"initialized":true}
```

## Database inspection

```sh
docker compose -f compose.qa.yml exec postgres-qa \
  psql -U donarium_qa -d donarium_qa
```

Common queries:

```sql
SELECT COUNT(*) FROM users;
SELECT COUNT(*) FROM organizations;
SELECT * FROM schema_migrations ORDER BY name;
```

## Backend restart

To restart just the API container (data persists in PostgreSQL):

```sh
docker compose -f compose.qa.yml restart donarium-api-qa
```

## Stop and cleanup

| Command | Effect |
|---|---|
| `make qa-down` | Stop all QA containers. **Preserves** PostgreSQL data. |
| `make qa-reset` | Destroy QA containers and volume, then recreate. **All data is lost.** |

## Migration strategy

The backend applies migrations automatically on startup (`database.RunMigrations`
in `cmd/donarium/main.go`). No separate migration service is needed. The
`schema_migrations` table tracks which migrations have run — restarting the
backend reapplies only new ones.

## Topology

```
donarium-qa-network (isolated)
├── postgres-qa           (127.0.0.1:15432, volume: donarium-postgres-qa-data)
└── donarium-api-qa       (127.0.0.1:18080)
    └── depends_on: postgres-qa (healthy)
```

No production, development, or external services are connected.
