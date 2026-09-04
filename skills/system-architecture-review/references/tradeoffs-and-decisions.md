# Tradeoffs and Decisions

## ADR-Style Summary Template

```markdown
## Decision
Use an outbox table + async consumer for payment capture after order creation.

## Context
Checkout must stay fast; payment provider latency is p95 800ms and occasionally times out.

## Options
1. Synchronous payment in order request — rejected: ties availability to provider.
2. Saga with compensating transactions — viable but heavy for current team size.
3. Outbox + async worker — chosen: simpler ops, idempotent consumer, clear retry.

## Consequences
+ Faster checkout response
+ Retries isolated to worker
- Eventual payment state visible to users (needs UI status)
- Requires idempotency keys and dead-letter handling
```

## Common Architecture Failure Patterns

**Distributed monolith** — many services, one shared database, synchronized releases.

**Chatty integration** — fine-grained API calls in a loop where batch or event stream fits.

**Cache as consistency fix** — caching without invalidation strategy when reads must be fresh.

**Platform before product** — generic workflow engine before two concrete workflows exist.

## Scalability Questions

- What is the expected RPS and payload size in 6–12 months?
- Which resource saturates first under 2× load?
- Is the bottleneck compute, DB connections, lock contention, or external quota?
- Can the design scale horizontally without sticky session assumptions?

## When to Recommend a Spike

Recommend a time-boxed spike when:

- Latency or throughput claims are untested
- A new datastore or queue is introduced
- Splitting a monolith has multiple viable cut lines
- Compliance or security constraints block a obvious design

Spike output should answer one question with measurable evidence — not build production code.
