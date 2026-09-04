## Recovery Checkpoint

### 1. Original Objective

DON-007Q.2-F2: Create a test-only seam that demonstrates real PostgreSQL
rollback after at least one write. Unblock TC-001-09.

### 2. Seam Design

**Chosen pattern:** Repository decorator around `OrganizationRepository.Create`.

The decorator delegates to the real repository (real INSERT), then returns a
deterministic error. The `TransactionManager.WithinTransaction` catches the
error and rolls back via `defer tx.Rollback(ctx)`.

**Write order in `persistSetup`:**
user INSERT → credential INSERT → org INSERT (real, inside decorator) → decorator fails → rollback

Three real PostgreSQL writes executed; all undone by the transaction rollback.

### 3. Completed Work

| File | Change |
|---|---|
| `tests/integration/setup_vertical_test.go` | `failingOrgRepo` decorator + `TestTC_001_09_RepositoryDecoratorRollback` |

### 4. Validation Status

| Check | Result |
|---|---|
| `go vet ./...` | PASS |
| `go test ./...` | PASS (67/67) |
| Real writes executed before failure | CONFIRMED |
| All 5 tables = 0 rows after rollback | CONFIRMED |
| No production code changes | CONFIRMED |
| No test flags or endpoints | CONFIRMED |

### 5. Current Blocker

None. TC-001-09 is unblocked.

### 6. Remaining Work

None.

### 7. Checkpoint Status

RESOLVED
