# Problem Framing and Requirements

## Problem Statement Template

A strong problem statement answers **who**, **what pain**, **when**, and **why now** — without naming the implementation.

```markdown
## Problem
Checkout teams wait 2–3 days for manual fraud review on 8% of orders, causing cart abandonment after payment authorization.

## Goal
Reduce post-auth review queue time to under 4 hours for 95% of flagged orders within one release cycle.

## Non-goals
- Replacing the existing payment provider
- Real-time ML scoring in v1
```

## Scope Boundaries

| In scope | Out of scope |
| --- | --- |
| Async review workflow for flagged orders | International payment methods |
| Status API for order tracking | Admin UI redesign |

## Requirement Types

### Functional

- Observable behavior the system must provide
- Written as **capabilities** or **rules**, not technology choices

**Weak:** "Use Kafka for events."  
**Strong:** "When an order is flagged, downstream review receives the order within 60 seconds with idempotent delivery."

### Non-functional

| Attribute | Example acceptance criterion |
| --- | --- |
| Performance | p95 API latency < 200ms at 500 RPS |
| Availability | 99.9% monthly for checkout API |
| Security | PII encrypted at rest; no card data in logs |
| Operability | Deploy rollback < 15 minutes; health checks on all services |
| Maintainability | New engineer can run locally in < 30 minutes (documented) |

### Constraints

- Regulatory (PCI, GDPR, SOC2)
- Platform (must run on existing EKS cluster)
- Team (two engineers for eight weeks)
- Compatibility (must not break mobile clients on v2 API)

## Acceptance Criteria Quality

Use **Given / When / Then** or checklist form that is verifiable in test or demo:

```markdown
- [ ] Given a flagged order, when review completes approve, then order status becomes `CAPTURED` within 5s.
- [ ] Given duplicate webhook delivery, when processor runs twice, then payment is captured once.
```

Avoid:

- Subjective adjectives without measure ("fast", "scalable", "user-friendly")
- Duplicate requirements stated as both functional and NFR
- Hidden dependencies ("as today" without describing today)

## Questions to Ask Before Design

1. What happens if we do nothing?
2. What is the smallest outcome that proves value?
3. Which decisions are reversible in one sprint?
4. Who approves scope changes and security exceptions?
5. What data is created, stored, and deleted — and for how long?

## Handoff to Documentation

When the analysis is ready to publish, map content to document types in `technical-documentation-authoring` → `references/document-types-and-templates.md`.
