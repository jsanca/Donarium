# DON-007Q.2-F2 — Real PostgreSQL Rollback QA Seam

**Status:** COMPLETE
**Owner:** Deep
**Role:** Testability Engineer
**Defect:** TC-001-09 (unblocked)

## Summary

Created a test-only repository decorator that injects a deterministic failure
after a real PostgreSQL write inside a transaction. The `TransactionManager`
rolls back the entire operation, verified by querying all five tables with a
fresh connection and confirming zero rows.

No flags, no endpoints, no production code changes.

## Seam Design

**Pattern:** Decorator around `OrganizationRepository.Create`.

```go
type failingOrgRepo struct {
    identity.OrganizationRepository  // delegate
}

func (r *failingOrgRepo) Create(ctx context.Context, db identity.DBExecutor,
    org identity.Organization) error {
    // 1. Real write — INSERT INTO organizations
    if err := r.OrganizationRepository.Create(ctx, db, org); err != nil {
        return err
    }
    // 2. Induced failure
    return errors.New("induced failure after org creation for rollback test")
}
```

**Why this seam:**

- `OrganizationRepository.Create` is called third in `persistSetup` (after
  user and credential). Both preceding writes execute real SQL INSERTs before
  this decorator triggers the failure.
- The decorator is a single-file type in `tests/integration/` — zero impact
  on production code.
- All other collaborators are real: `*pgxpool.Pool`, `TransactionManager`,
  `UserRepo`, `CredentialRepo`, `MembershipRepo`, `PlatformGrantRepo`,
  `Argon2Hasher`, `EmailNormalizer`.

## Write-then-fail sequence

| Step | Entity | Real write? | Result |
|---|---|---|---|
| 1 | User | INSERT INTO users | Succeeds |
| 2 | Credential | INSERT INTO credentials | Succeeds |
| 3 | Organization | INSERT INTO organizations | Succeeds (real) |
| 3b | Decorator returns error | — | `induced failure after org creation` |
| 4 | Membership | INSERT skipped | — |
| 5 | Platform grant | INSERT skipped | — |
| — | TransactionManager | `defer tx.Rollback(ctx)` | All 3 writes undone |

## Evidence

```
=== RUN   TestTC_001_09_RepositoryDecoratorRollback
--- PASS (0.43s)
```

Post-rollback counts from fresh pool connection:

| Table | Expected | Actual |
|---|---|---|
| users | 0 | 0 |
| credentials | 0 | 0 |
| organizations | 0 | 0 |
| memberships | 0 | 0 |
| platform_grants | 0 | 0 |

## Why test-only

The `failingOrgRepo` type exists only in `tests/integration/setup_vertical_test.go`.
It is never referenced from production code. The `main.go` composition root
wires real repositories exclusively. No `FAIL_AFTER` flags, no query parameters,
no test-only code paths in the runtime.

## Limits

- The decorator fails on `OrganizationRepository.Create`, which is the 3rd of
  5 calls in `persistSetup`. It proves that 2 preceding writes are rolled back,
  but does not test failure after membership or platform grant creation.
- The failure is injected in a decorator (application layer), not at the SQL
  level. A SQL-level failure (e.g., constraint violation) would behave
  identically — pgx's `tx.Rollback` undoes all statements in the transaction.

## Files Changed

| File | Change |
|---|---|
| `tests/integration/setup_vertical_test.go` | Added `failingOrgRepo` decorator + `TestTC_001_09_RepositoryDecoratorRollback` |
| `docs/agents/reports/DON-007Q.2-F2-real-postgresql-rollback-qa-seam.md` | This report |
| `docs/agents/checkpoints/CHECKPOINT-DON-007Q.2-F2-real-postgresql-rollback-qa-seam.md` | Created |
| `docs/agents/ENGINEERING_LOG.md` | Updated |

## Validation

| Check | Result |
|---|---|
| `go vet ./...` | PASS |
| `go test ./...` | PASS (67/67) |
| Real PostgreSQL writes before failure | CONFIRMED (user + credential + org INSERTs executed) |
| Transaction rollback | CONFIRMED (all 5 tables = 0 rows post-test) |
| No production code changes | CONFIRMED |
| No test flags or endpoints | CONFIRMED |
| Passwords not logged | CONFIRMED |

## TC-001-09 Status

**BLOCKED → UNBLOCKED.** A reproducible, real-PostgreSQL rollback test now
exists demonstrating zero rows after a post-write failure.
