# Context and Boundaries

## Questions to Ask First

- Who are the actors (users, systems, admins, batch jobs)?
- What are the read vs write paths and their latency expectations?
- Which data is authoritative, cached, derived, or replicated?
- What happens when a dependency is slow or unavailable?
- Who operates each component on call?

## Simple Context Map (Text Format)

```text
[Browser/App]
    -> [API Gateway / BFF]
        -> [Order Service] -> [Orders DB]
        -> [Payment Service] -> [Payment Provider API]
        -> [Notification Service] -> [Email Queue]

Shared risk: Payment provider timeout blocks synchronous checkout if not decoupled.
```

## Boundary Smells

| Smell | Why it matters |
| --- | --- |
| Two services write the same table | No clear source of truth; race conditions |
| Synchronous chain across 4+ services | Latency and failure amplification |
| "Shared library" with business rules | Hidden coupling across deploy units |
| DB accessed directly by many services | Schema changes become organization-wide |
| No owner for integration contract | Breaking changes ship without notice |

## Bounded Context Heuristic

Split when **business language**, **change cadence**, or **scaling profile** diverge — not when code size feels large.

Keep together when:

- Same team owns the workflow end-to-end
- Strong transactional consistency is required
- Extracting would mostly move calls, not reduce complexity

Split when:

- Independent scaling or release cadence is needed
- Failure isolation has clear business value
- Data ownership can be cleanly separated with explicit contracts
