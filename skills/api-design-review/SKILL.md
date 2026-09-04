---
name: api-design-review
description: Review REST and HTTP API design for resource modeling, contracts, validation, errors, pagination, versioning, and backward compatibility. Use when designing new endpoints, reviewing OpenAPI specs, evaluating breaking changes, or standardizing error and pagination patterns.
---

# API Design Review

## Workflow

1. Identify API consumers: first-party UI, mobile, partners, internal services, or public developers.
2. Review resource naming, HTTP methods, status codes, and idempotency for each operation.
3. Check request/response schemas: required fields, validation rules, null semantics, and enum stability.
4. Evaluate error model consistency — clients should handle failures predictably.
5. Review pagination, filtering, sorting, and rate-limit headers for list endpoints.
6. Assess versioning and compatibility: additive vs breaking changes, deprecation path, and contract tests.
7. Confirm security at the contract layer: auth scopes, sensitive fields, and leakage in errors.

## References

- Read `references/rest-contract-patterns.md` for resource design, status codes, and idempotency.
- Read `references/errors-versioning-compatibility.md` for error shapes, evolution, and breaking change rules.

## Review Checklist

- Nouns for resources; verbs expressed by HTTP methods.
- `POST` create returns `201` with `Location` when appropriate.
- Errors use a stable, documented shape across endpoints.
- Pagination is bounded; defaults and max limits are documented.
- Breaking changes require version bump or explicit deprecation window.
- OpenAPI (or equivalent) matches implementation and CI contract tests.
- Sensitive data is not returned or logged unnecessarily.

## Output

Deliver an API review with:

- **Endpoints reviewed** and consumer context
- **Contract issues** with severity and example request/response
- **Recommended changes** (minimal diff first)
- **Compatibility impact** — breaking vs safe additive
- **Verification** — contract tests, consumer updates, rollout plan
