# DON-007Q.2 — Execute Initialize Donarium QA

- **Status:** Complete
- **Owner:** Elito
- **Role:** QA Tester
- **Executed:** 2026-07-21 (America/Denver)
- **Code revision tested:** `8868748` (working tree contained pre-existing uncommitted work)
- **Initial verdict:** **FAIL**
- **Current verdict after targeted revalidation:** **PASS WITH OBSERVATIONS**

## Context

This report records execution of [TC-001 — Initialize Donarium](../../../knowledge/test-cases/TC-001-initialize-donarium.md) against the real backend and PostgreSQL. It follows [UC-001](../../../knowledge/use-cases/UC-001-initialize-donarium.md), the business rules, and [ADR-001](../../../knowledge/decisions/ADR-001-installation-bootstrap.md) / [ADR-002](../../../knowledge/decisions/ADR-002-application-level-bootstrap-invariant.md).

No application code, contracts, or tests were changed. No commit was created.

## Environment

| Item | Evidence |
| --- | --- |
| API under test | Dedicated Docker service `donarium-api-qa`, `http://127.0.0.1:18080` |
| PostgreSQL | Dedicated Docker service `donarium-postgres-qa`, local port `15432` |
| Isolation | Dedicated `donarium-qa-network` and `donarium-postgres-qa-data`; distinct from development ports `8080` and `5432` |
| Clean state | `make qa-reset` removed and recreated the dedicated QA volume before execution; initial status was `{"initialized":false}` and all five entity counts were zero |
| Migrations | API startup applied migrations automatically; inspection found `migration_count=5` |
| Server logs | Captured through `docker compose -f compose.qa.yml logs`; scans found zero occurrences of the rejected raw value and zero `password=` / `password:` fields |
| Health | Liveness returned `{"status":"ok"}` and readiness returned database `up` |

Only the QA resources named above were reset or queried. No development, production, shared, or external database was used. Connection strings and credentials are intentionally omitted.

## Execution Summary

| Scenario | Result | Summary |
| --- | --- | --- |
| TC-001-01 | PASS | Fresh status returned `200` / JSON `initialized:false`; all counts zero. |
| TC-001-02 | PASS | Setup returned `201`; five expected records, normalized email and contextual roles persisted. |
| TC-001-03 | PASS | Status returned `200` / JSON `initialized:true`. |
| TC-001-04 | PASS | Second setup returned `409`; all five counts remained one. |
| TC-001-05 | PASS | Invalid email returned `400`; all five counts remained zero. |
| TC-001-06 | PASS WITH OBSERVATION | Weak password returned `400`; all five counts remained zero and raw secret scan passed. |
| TC-001-07 | PASS | Every required-field variant returned its expected `400`; all counts remained zero. |
| TC-001-08 | **FAIL** | Router returned `405` and `Allow`, but omitted required JSON content type and error body. |
| TC-001-09 | BLOCKED | No safe, real-PostgreSQL failure injector exists without changing code; see deferred work. |
| TC-001-10 | PASS | After API restart against the same QA database, status remained initialized and a second setup still returned `409`. |
| TC-001-11 | NOT RUN — deferred | Explicitly deferred by ADR-002; not executed. |

## Targeted Revalidation — TC-001-08 and TC-001-09

**Executed:** 2026-07-21 (America/Denver), against the rebuilt isolated QA
stack. Scope was deliberately limited to the two previously non-passing
scenarios; the earlier evidence for all other scenarios remains unchanged.

| Scenario | Initial result | Revalidation result | Evidence |
| --- | --- | --- | --- |
| TC-001-08 | FAIL | PASS | Both wrong-method requests returned `405`, the correct `Allow`, `Content-Type: application/json`, and `{"error":"method not allowed"}`. |
| TC-001-09 | BLOCKED | PASS | `TestTC_001_09_RepositoryDecoratorRollback` passed against real QA PostgreSQL; final counts of all five tables were zero. |

### TC-001-08 — Unsupported HTTP methods (revalidated)

- **Precondition:** Rebuilt, healthy QA API and dedicated PostgreSQL.
- **Commands:** `GET /api/setup`; `POST /api/setup/status`.
- **Observed:**
  - `GET /api/setup` → `405 Method Not Allowed`, `Allow: POST`, `Content-Type: application/json`, `{"error":"method not allowed"}`.
  - `POST /api/setup/status` → `405 Method Not Allowed`, `Allow: GET`, `Content-Type: application/json`, `{"error":"method not allowed"}`.
- **Result:** PASS. DQ2-001 is resolved.

### TC-001-09 — Transaction rollback after a write (revalidated)

- **Precondition:** Dedicated migrated QA PostgreSQL; no development database involved.
- **Command:** `TEST_DATABASE_URL=<redacted> go test ./tests/integration -run '^TestTC_001_09_RepositoryDecoratorRollback$' -count=1 -v` from `server/`.
- **Observed:** `TestTC_001_09_RepositoryDecoratorRollback` passed. Its repository decorator writes the Organization through the real repository, then induces a deterministic failure inside `TransactionalSetupService`'s transaction.
- **Persistence:** Direct post-test inspection returned `0` for users, credentials, organizations, memberships, and platform grants.
- **Result:** PASS. OQ2-002 is resolved.

The API readiness endpoint returned database `up` after the revalidation.

## Scenario Results and Evidence

Commands below are representative and redact request secrets. Timestamps for this execution fall on 2026-07-21 in America/Denver.

### TC-001-01 — Initial state

- **Precondition:** `make qa-reset`; dedicated migrated database.
- **Command:** `GET /api/setup/status`.
- **Observed:** `200 OK`; `Content-Type: application/json`; body `{"initialized":false}`.
- **Persistence:** users, credentials, organizations, memberships, and platform grants each counted `0`.
- **Result:** PASS.

### TC-001-02 — Successful initialization

- **Precondition:** Same clean QA database.
- **Command:** `POST /api/setup` with the TC-001 base payload (password redacted).
- **Observed:** `201 Created`; `Content-Type: application/json`; response contained non-empty UUID `userId` and `organizationId`, and no credential fields.
- **Persistence:** exactly one record in each of the five tables. The User email was normalized to `owner@northstar.example`; Organization was `Northstar Rentals` / `northstar-rentals`; Membership role was `owner`; PlatformGrant role was `super_admin`; credential value had the `$argon2id$` prefix.
- **Result:** PASS.

### TC-001-03 — Initialized status

- **Precondition:** TC-001-02 completed without reset.
- **Command:** `GET /api/setup/status`.
- **Observed:** `200 OK`; `Content-Type: application/json`; body `{"initialized":true}`.
- **Result:** PASS.

### TC-001-04 — Repeated setup

- **Precondition:** TC-001-02 completed.
- **Command:** `POST /api/setup` with the alternate TC-001 payload (password redacted).
- **Observed:** `409 Conflict`; `Content-Type: application/json`; body `{"error":"system is already initialized"}`.
- **Persistence:** all five counts stayed `1`.
- **Result:** PASS.

### TC-001-05 — Invalid email

- **Precondition:** Fresh dedicated QA database after `make qa-reset`.
- **Command:** `POST /api/setup` with invalid email and otherwise valid payload (password redacted).
- **Observed:** `400 Bad Request`; `Content-Type: application/json`; body `{"error":"email is not valid"}`.
- **Persistence:** all five counts were `0`.
- **Result:** PASS.

### TC-001-06 — Weak password

- **Precondition:** Dedicated QA database remained empty; server log capture active.
- **Command:** `POST /api/setup` with a password rejected by policy (value redacted).
- **Observed:** `400 Bad Request`; `Content-Type: application/json`; body `{"error":"password does not meet requirements"}`.
- **Persistence:** all five counts were `0`.
- **Log evidence:** zero matches for the rejected raw value and zero `password=` / `password:` fields. The structured `error` value included the public validation message; it did not include the supplied secret.
- **Result:** PASS WITH OBSERVATION.

### TC-001-07 — Missing required fields

- **Precondition:** Dedicated QA database empty.
- **Command:** Five `POST /api/setup` variants, each omitting one required field.
- **Observed:** Every variant returned `400 Bad Request` with the expected JSON body:

| Omitted field | Body |
| --- | --- |
| `displayName` | `{"error":"displayName is required"}` |
| `email` | `{"error":"email is required"}` |
| `password` | `{"error":"password is required"}` |
| `organizationName` | `{"error":"organizationName is required"}` |
| `organizationSlug` | `{"error":"organizationSlug is required"}` |

- **Persistence:** all five counts were `0` after the variants.
- **Result:** PASS.

### TC-001-08 — Unsupported HTTP methods (initial execution)

- **Precondition:** No dependency on initialization state.
- **Commands:** `GET /api/setup`; `POST /api/setup/status`.
- **Observed:** Both returned `405 Method Not Allowed`. `Allow` was respectively `POST` and `GET`. Neither response included `Content-Type: application/json` nor a response body.
- **Expected:** `405`, respective `Allow`, JSON content type, and `{"error":"method not allowed"}`.
- **Persistence:** no writes were introduced; the counts remained those of the pre-existing initialized scenario (`1` each).
- **Result:** Initial FAIL. Resolved and revalidated as PASS above.

### TC-001-09 — Transaction rollback after a write (initial execution)

- **Precondition:** Dedicated QA database available.
- **Assessment:** BLOCKED.
- **Evidence:** Existing application tests use fakes; the real-PostgreSQL integration suite has no test-only collaborator that fails after User/Credential writes. The QA-environment report also identifies this as a follow-up.
- **Reason:** Creating an injector or modifying runtime composition would violate this QA task's no-code-change restriction. Simulating a database failure outside a controlled collaborator would not be reproducible or safe.
- **Result:** Initially BLOCKED; resolved and revalidated as PASS above.

### TC-001-10 — Persistence across API restart

- **Precondition:** TC-001-02 completed successfully.
- **Command:** `docker compose -f compose.qa.yml restart donarium-api-qa`, then status and repeated setup requests.
- **Observed:** After readiness, status returned `200` / `{"initialized":true}`. Repeated setup returned `409` / `{"error":"system is already initialized"}`. All five counts remained `1`.
- **Result:** PASS.

### TC-001-11 — Concurrent initialization

- **Result:** NOT RUN — deferred by ADR-002. No concurrency verdict is inferred from this QA run.

## Defects

### DQ2-001 — Route-level 405 responses violate the JSON error contract

**Status:** RESOLVED by targeted revalidation on 2026-07-21.

| Field | Detail |
| --- | --- |
| Severity | MAJOR |
| Scenario | TC-001-08 |
| Expected | JSON `405` responses with `Content-Type: application/json`, `Allow`, and `{"error":"method not allowed"}`. |
| Observed | `GET /api/setup` and `POST /api/setup/status` returned `405` and correct `Allow`, but no content type and an empty body. |
| Evidence | Real API requests against `http://127.0.0.1:18080`; reproduced for both registered setup routes. |
| Affected area | Route registration in `server/internal/identity/http/runtime.go`; current handler-level wrong-method branches are bypassed by Chi route matching. |
| Recommendation | Configure router-level method-not-allowed handling to use the standard JSON error envelope, then add an end-to-end router test for both paths. |

## Observations

- **OQ2-001 — Validation logging is secret-safe but includes the public validation phrase.** The raw rejected value and password field key were absent from captured logs. The `error` field carries the public validation message. This is not a secret leak; keep it under review if logging vocabulary becomes more restrictive.
- **OQ2-002 — TC-001-09 needs a purpose-built isolated failure injector.** RESOLVED: a repository decorator in the integration test now produces a deterministic post-write failure against real PostgreSQL without affecting production runtime.

## Deferred Scenario

TC-001-11 remains deferred under ADR-002. The current product decision accepts the theoretical bootstrap race and this QA execution did not alter that decision.

## Validation

- QA stack reset, rebuilt and started successfully with `make qa-reset`.
- API liveness and readiness were healthy.
- Direct PostgreSQL inspection confirmed five migrations and the scenario-specific entity counts recorded above.
- No code formatting, build-only validation, or unrelated test suite was run as a substitute for the functional QA evidence.

## Tests

Functional/API/persistence scenarios TC-001-01 through TC-001-08 and TC-001-10 were exercised against real PostgreSQL. TC-001-09 was subsequently exercised with real PostgreSQL and passed. TC-001-11 was intentionally not run.

## Tradeoffs

The QA volume was reset only through the explicit QA Compose file to obtain clean state between independent scenarios. This provides trustworthy zero-count assertions while preserving development resources. No test hooks were added merely to force TC-001-09.

## Open Questions

- Should the router's method-not-allowed behavior become a shared API policy before authentication work starts?
- Which isolated real-PostgreSQL failure injection seam should be approved for TC-001-09?

## Follow-ups

1. No follow-up remains for TC-001-08 or TC-001-09.
2. TC-001-11 remains deferred unless ADR-002 changes.

## References

- [TC-001](../../../knowledge/test-cases/TC-001-initialize-donarium.md)
- [UC-001](../../../knowledge/use-cases/UC-001-initialize-donarium.md)
- [Business Rules](../../../knowledge/business-rules.md)
- [ADR-001](../../../knowledge/decisions/ADR-001-installation-bootstrap.md)
- [ADR-002](../../../knowledge/decisions/ADR-002-application-level-bootstrap-invariant.md)
- [QA environment](../../../knowledge/testing/qa-environment.md)
