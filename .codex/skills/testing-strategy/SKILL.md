---
name: testing-strategy
description: Design and review test strategy across unit, integration, contract, and end-to-end layers with CI quality gates. Use when planning test coverage for a feature, improving CI reliability, deciding what to mock vs integrate, or establishing release verification standards.
---

# Testing Strategy

## Workflow

1. Understand the change: user-facing behavior, APIs, data migrations, infra, or refactor-only.
2. Identify risks: correctness, regression, compatibility, performance, security, and operability.
3. Map tests to the test pyramid — favor fast, deterministic tests; use E2E sparingly for critical paths.
4. Define what must be mocked vs tested with real dependencies (DB, queue, HTTP).
5. Specify CI gates: required jobs, coverage expectations (if used), flaky test policy, and PR vs main differences.
6. Plan verification for rollback, feature flags, and canary releases when applicable.
7. Document gaps explicitly — not every layer needs maximum coverage.

## References

- Read `references/pyramid-and-ci-gates.md` for layer selection and pipeline gates.
- Read `references/risk-based-test-planning.md` for prioritization and regression targeting.

## Strategy Checklist

- Critical business paths have automated regression tests.
- Unit tests cover pure logic and edge cases.
- Integration/slice tests cover persistence, HTTP, and messaging boundaries.
- Contract tests protect API consumers and providers.
- E2E tests cover few high-value journeys — stable and parallelizable.
- CI fails on test failure; flaky tests are quarantined and fixed.
- Test data and environments are reproducible locally and in CI.

## Output

Deliver a testing strategy with:

- **Scope and quality goals**
- **Risk matrix** — what could break and severity
- **Recommended test layers** per risk with examples
- **CI gates** — jobs, commands, blocking vs advisory
- **Gaps and deferrals** — what is not tested and why
- **Next actions** — specific tests to add first
