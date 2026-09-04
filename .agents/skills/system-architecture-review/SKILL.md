---
name: system-architecture-review
description: Review system architecture for boundaries, coupling, scalability, operability, and migration risk. Use when evaluating a new design, reviewing a PR with architectural impact, assessing monolith vs service split, or preparing an architecture decision record.
---

# System Architecture Review

## Workflow

1. Clarify the goal: new capability, refactor, scale event, cost reduction, reliability fix, or migration.
2. Map the current context: users, workloads, data flows, external dependencies, deployment units, and team ownership.
3. Identify boundaries: what belongs inside vs outside each component; where consistency, latency, and failure isolation matter.
4. Evaluate tradeoffs across coupling, cohesion, data ownership, sync vs async communication, and operational complexity.
5. Surface risks: single points of failure, shared mutable state, unclear ownership, hidden synchronous chains, and schema coupling.
6. Recommend the smallest change that meets the goal — avoid premature microservices or speculative abstraction.
7. Document decisions, rejected options, and verification steps.

## References

- Read `references/context-and-boundaries.md` for context mapping, bounded contexts, and ownership questions.
- Read `references/tradeoffs-and-decisions.md` for ADR-style tradeoff analysis and common failure patterns.

## Review Checklist

- Each component has a clear responsibility and owner.
- Data has a single source of truth per bounded context.
- Failure modes are isolated where the business requires it.
- Scaling strategy matches actual bottlenecks (CPU, DB, queue, external API).
- Security and compliance boundaries are explicit.
- Operability (deploy, observe, rollback) is designed in — not bolted on later.
- Migration path exists if the design replaces or straddles legacy systems.

## Output

Produce an architecture review with:

- **Scope** — what was reviewed and assumptions
- **Current state** — diagram or bullet flow of components and dependencies
- **Findings** — severity (blocker / major / minor), evidence, and impact
- **Recommendations** — ordered by priority with smallest viable step first
- **Tradeoffs** — options considered and why one was preferred
- **Open questions** — decisions needing product, security, or ops input
- **Verification** — spikes, load tests, or proofs of concept to run before commit
