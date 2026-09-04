# Design Options and Tradeoffs

## Option Analysis Pattern

For each option, document:

| Field | Content |
| --- | --- |
| Name | Short label (e.g. "Outbox + worker") |
| Summary | One sentence |
| Changes | Components, data stores, deployment |
| Pros | Concrete benefits |
| Cons | Concrete costs and risks |
| Effort | T-shirt or time range with assumptions |
| Reversibility | Easy / moderate / hard to undo |

Present **at least two** serious options plus **do nothing** or **defer** when relevant.

## Tradeoff Matrix Example

| Criterion | Option A: sync call | Option B: outbox + async |
| --- | --- | --- |
| Time to implement | Low | Medium |
| Checkout latency | Tied to provider p95 | Decoupled |
| Consistency | Immediate | Eventual |
| Operational load | Low | Queue monitoring required |
| Failure isolation | Poor | Good |

Weight criteria by **business priority**, not personal preference.

## Quality Attribute Scenarios

Describe scenarios that stress the design:

```markdown
### Scenario: Provider timeout
When payment provider exceeds 5s, checkout returns 202 with order id;
worker retries with exponential backoff; user sees pending status.

### Scenario: Review backlog spike
When queue depth > 1000, alert fires; auto-scale workers to max 10;
SLA dashboard shows age of oldest item.
```

## Decision Record (Lightweight ADR)

When a option is chosen, capture:

```markdown
## Decision
Adopt transactional outbox for post-checkout fraud handoff.

## Status
Proposed | Accepted | Superseded

## Context
(Link to problem statement and constraints)

## Decision drivers
- Checkout latency SLO
- Team familiarity with PostgreSQL

## Considered options
1. ...
2. ...

## Outcome
Chosen option B because ...

## Consequences
Positive: ...
Negative: ...
Follow-ups: ...
```

Full formatting guidance: `technical-documentation-authoring` → `references/document-types-and-templates.md`.

## Anti-Patterns in Analysis

| Anti-pattern | Fix |
| --- | --- |
| Analysis paralysis | Time-box; recommend default with explicit spike |
| Single option | Add build/buy/defer alternatives |
| Technology-first | Restate problem in user/outcome terms |
| Ignoring operations | Add deploy, observe, rollback to every option |
| Hidden coupling | Draw data flow; name owners per datastore |

## When to Escalate to Architecture Review

Use `system-architecture-review` when the change affects:

- Service boundaries or team ownership
- Shared databases or cross-domain events
- Platform-wide security or compliance posture
- Multi-region or multi-tenant isolation

Use `api-design-review` when the primary surface is a public or partner HTTP contract.

## Spikes and Prototypes

Recommend a spike when uncertainty blocks a decision:

- Throughput or latency unvalidated
- Unfamiliar infrastructure (new queue, DB, or identity model)
- Legal or security sign-off needs a concrete demo

Spike deliverable: **answer one question** with evidence — not production code.
