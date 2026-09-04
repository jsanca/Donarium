# Test Pyramid and CI Gates

## Layer Guide

| Layer | Speed | Typical scope | When to use |
| --- | --- | --- | --- |
| Unit | Fastest | Pure functions, domain rules | Always for business logic |
| Slice / integration | Medium | DB, HTTP slice, message handler | Boundary behavior |
| Contract | Medium | API provider/consumer | Shared APIs, mobile, partners |
| E2E | Slowest | Full stack user journey | Few critical paths |

## Example CI Pipeline Gates

```text
PR:
  - lint / static analysis
  - unit tests (all modules)
  - affected integration tests
  - contract tests (if API changed)

Main:
  - full unit + integration suite
  - container build + smoke health check
  - optional nightly E2E against staging
```

Block merge on PR gates that match production risk — do not run 45-minute E2E on every PR unless infrastructure supports it reliably.

## Per-Technology Quality Gate Skills

When implementing or fixing CI for a specific stack, use the matching skill:

| Stack | Skill |
| --- | --- |
| Java (Gradle/Maven) | `java-quality-gates` |
| .NET | `dotnet-quality-gates` |
| PHP | `php-quality-gates` |
| React | `react-quality-gates` |
| Angular | `angular-quality-gates` |

These skills cover static analysis, formatting, build/test commands, and coverage thresholds. This document covers **strategy**; stack skills cover **tooling**.

## Mock vs Real Dependencies

**Mock** when:

- External API is slow, costly, or nondeterministic
- Testing error paths that are hard to trigger in sandboxes

**Use real** (Testcontainers, embedded DB, wire mock server) when:

- SQL/ORM behavior is the risk
- Serialization or schema migration is the risk
- Contract assumptions must hold across versions

## Coverage Policy (If Used)

Coverage is a **signal**, not a goal by itself.

- Set thresholds on critical modules, not blanket 80% everywhere
- Require tests for bug fixes (regression test mandatory)
- Exclude generated code and trivial DTO mappers from gates if they add noise

## Flaky Test Policy

1. Quarantine with ticket link — do not ignore silently
2. Fix or delete within agreed SLA
3. Never retry indefinitely without investigating root cause
