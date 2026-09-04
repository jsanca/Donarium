## Recovery Checkpoint

### 1. Original Objective

DON-007.10: Create an isolated, reproducible QA environment for functional
testing against real PostgreSQL. No shared development databases, no mocks.

### 2. Completed Work

| Deliverable | Status |
|---|---|
| `compose.qa.yml` | Created — postgres-qa (port 15432) + donarium-api-qa (port 18080), isolated network and volume |
| `Makefile` QA targets | `qa-up`, `qa-down`, `qa-reset`, `qa-status`, `qa-logs` |
| `.env.qa.example` | Reference connection values |
| `knowledge/testing/qa-environment.md` | User documentation |
| Migration strategy | Option A — backend runs migrations at startup (existing behavior) |
| Validation V1–V6 | All 6 scenarios verified and logged |

### 3. Environment Topology

```
donarium-qa-network
├── donarium-postgres-qa   → 127.0.0.1:15432 (vol: donarium-postgres-qa-data)
└── donarium-api-qa        → 127.0.0.1:18080 (depends_on: postgres-qa healthy)
```

### 4. Files Changed

| File | Change |
|---|---|
| `compose.qa.yml` | Created |
| `Makefile` | Added 5 QA targets |
| `.env.qa.example` | Created |
| `knowledge/testing/qa-environment.md` | Created |
| `docs/agents/reports/DON-007.10-reproducible-qa-environment.md` | Created |
| `docs/agents/ENGINEERING_LOG.md` | Updated |

### 5. Validation Status

| Check | Result |
|---|---|
| V1 — Clean environment (qa-reset) | PASS — both services healthy |
| V2 — Initial status (GET /api/setup/status) | PASS — initialized=false |
| V3 — Persistence during restart | PASS — data survives backend restart |
| V4 — Reset destroys data | PASS — initialized=false after qa-reset |
| V5 — Isolation (ports, names, volumes, networks) | PASS — no overlap with dev |
| V6 — Teardown (qa-down) | PASS — no lingering processes |
| No dev database used | PASS |
| No production secrets exposed | PASS |

### 6. Current Blocker

None.

### 7. Remaining Work

TC-001-09 (transactional rollback injection) is compatible with this environment
but requires a follow-up task for the injection mechanism.

### 8. Checkpoint Status

RESOLVED
