# Risk-Based Test Planning

## Risk Matrix Template

```text
| Risk                          | Likelihood | Impact | Test focus                    |
|-------------------------------|------------|--------|-------------------------------|
| Payment double-charge         | Medium     | High   | Idempotency integration test  |
| Wrong tenant sees data        | Low        | High   | Authz integration + contract  |
| UI typo on label              | High       | Low    | Manual or snapshot (optional) |
| Migration breaks old clients  | Medium     | Medium | Contract + backward compat    |
```

Prioritize tests where **likelihood × impact** is highest.

## Feature Test Plan Template

```markdown
### Feature: Export orders CSV

**Behaviors to prove**
- Authorized user exports own tenant orders only
- Empty result produces valid CSV header
- Large export paginates without OOM (load test optional)

**Tests**
- Unit: CSV row formatting
- Integration: repository query + tenant filter
- API: 403 for other tenant, 200 with fixture data
- E2E: none (covered by API integration)

**Out of scope**
- Email delivery of export link (separate feature)
```

## Regression Selection

When CI time is limited, run tests affected by:

- Changed packages/modules (build tool graph)
- Changed OpenAPI / protobuf contracts
- Changed migrations or shared fixtures

Document the selection command (`./gradlew test --tests ...`, path filters in GitHub Actions, etc.).

## Release Verification Beyond Automated Tests

- Smoke test on staging after deploy
- Manual exploratory pass for UX-critical flows (time-boxed)
- Monitor error rate and latency for N minutes post-release

Testing strategy should name **who runs** manual steps and **what blocks** promotion.
