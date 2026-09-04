# Server packages

Go package layout and dependency direction for `server/` (module
`donarium/server`). Derived from `go build ./...` and source inspection.

## Entrypoint and composition root

- `cmd/donarium/main.go` — the only entrypoint. Loads config, connects
  PostgreSQL, runs embedded migrations, wires all repositories/services/handlers,
  and composes `PlatformRuntime` + `IdentityRuntime` + `PropertiesRuntime` into
  `runtime.NewApplication`.

## Bounded contexts

### Identity — `internal/identity/`

- Domain types at package root: `user.go`, `credential.go`, `organization.go`,
  `membership.go`, `platform_grant.go`, `role.go`, `executor.go`,
  `repository.go`, `errors.go`, `service.go`, plus `password_policy.go`,
  `normalizer.go`.
- `application/` — use cases: `setup.go`, `transactional_setup.go`, `email.go`,
  `password.go`, `transaction.go`; and `application/authentication/`
  (`authenticate_user.go`, `principal_resolver.go`, `session_issuer.go`,
  `authenticated_principal.go`).
- `authorization/` — route guards and principal-context middleware.
- `http/` — handlers and middleware: `runtime.go`, `handler.go`,
  `login_handler.go`, `me_handler.go`, `auth_middleware.go`, `session_cookie.go`.
- `pgx/` — PostgreSQL persistence + adapters (`*_repository.go`, `argon2.go`,
  `session.go`, `transaction.go`, `executor.go`).

Layering (dependency direction is inward, top → bottom):

```text
http/  ──>  application/  ──>  domain (root)  <──  pgx/
```

The `http/` adapter imports `pgx` only to compose a `DBExecutor` from the pool
(adapter coupling, not domain contamination).

### Properties — `internal/properties/`

- Domain at package root: `property.go` (Property, Classification, RentalCadence,
  Address), `party.go` (Party), `stakeholder.go` (PropertyStakeholder,
  StakeholderRole), `repository.go` (Repository + StakeholderRepository ports),
  `executor.go` (anti-corruption `DBExecutor`), `errors.go`.
- `application/` — `service.go` (RegisterProperty / ListAccessible / GetByID).
- `http/` — `handler.go`, `dto.go`, `runtime.go` (routes under `/api/properties`).
- `pgx/` — `repository.go`, `stakeholder_repository.go`, `transaction.go`,
  `executor.go`.

The Properties domain references `users`/`organizations` only at the **SQL**
level (stakeholder access join via `memberships`); there is no Go import of
`identity`, preserving the Go-layer boundary.

### Platform — `internal/platform/`

- `config/` — env parsing (`config.Load`).
- `database/` — `postgres.go` (pool), `migrate.go` (embedded migrations).
- `http/health/` — liveness/readiness handlers.
- `runtime/` — `application.go` (chi router + HTTP server), `module.go`
  (`ModuleRuntime` interface), `platform.go`, `lifecycle.go`.
- `observability/` — reserved (empty).

## Runtime composition

```text
main.go
  ├─ config.Load()
  ├─ database.NewPool() ─ database.RunMigrations()
  ├─ identity  (repos + CanonicalSetup + TransactionalSetup + AuthenticateUser
  │             + PrincipalResolver + HMACSessionIssuer + AuthenticationMiddleware)
  ├─ properties (Repository + StakeholderRepository + TransactionManager
  │             + Service + Handler + Runtime)
  └─ runtime.NewApplication(cfg,
       PlatformRuntime(health), IdentityRuntime(setup/login/me),
       PropertiesRuntime(properties))
           └─ chi.Router { RegisterRoutes(m) for each module }
```

HTTP routes (mounted by each `*Runtime.RegisterRoutes`):

- `GET /health/live`, `GET /health/ready` — PlatformRuntime.
- `POST /api/setup`, `GET /api/setup/status` — IdentityRuntime.
- `POST /api/auth/login`, `GET /api/auth/me` (protected) — IdentityRuntime.
- `POST /api/properties`, `GET /api/properties`,
  `GET /api/properties/{id}` (protected) — PropertiesRuntime.

## Migrations

`internal/platform/database/migrations/` (`//go:embed` in `migrate.go`), run on
startup. Numeric-prefixed `.up.sql`/`.down.sql` pairs:

- `001_users`, `002_credentials`, `003_organizations`, `004_memberships`,
  `005_platform_grants` — identity/bootstrap.
- `006_properties` — property aggregate.
- `007_property_stakeholders` — Party/PropertyStakeholder + access rule.
