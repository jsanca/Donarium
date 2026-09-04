---
name: migration-planning
description: Plan cross-stack modernizations including version upgrades, framework migrations, datastore moves, and strangler fig cutovers. Use when scoping a migration program, sequencing work across teams, or assessing risk and rollback for large technical change.
---

# Migration Planning

## Workflow

1. Document **current state**: versions, dependencies, deployment model, integrations, and business constraints.
2. Define **target state** and success criteria — measurable, not "latest version."
3. Inventory **coupling points**: shared libraries, DB schemas, APIs, batch jobs, and external consumers.
4. Choose migration pattern: big bang, phased module, strangler fig, parallel run, or feature-flagged cutover.
5. Sequence work into increments that each deliver value or reduce risk — each increment must be deployable.
6. Plan verification: automated tests, shadow traffic, parity checks, and rollback/forward-fix for each phase.
7. Assign owners, timeline, comms, and decommission criteria for legacy paths.

## References

- Read `references/patterns-and-sequencing.md` for strangler, parallel run, and increment sizing.
- Read `references/risk-rollback-comms.md` for risk register, rollback limits, and stakeholder updates.

## Planning Checklist

- Each phase has clear entry/exit criteria.
- Consumers of changed APIs/contracts are identified.
- Data migration strategy includes backfill, validation, and reconciliation.
- Rollback or forward-fix documented per phase — not assumed.
- Team capacity and skill gap accounted for.
- Decommission plan exists — migrations fail when old path lives forever.

## Output

Produce a migration plan with:

- **Current / target state** summary
- **Pattern chosen** and rejected alternatives
- **Phased roadmap** — scope, duration, dependencies per phase
- **Risk register** — likelihood, impact, mitigation
- **Verification** — tests, metrics, parity checks per phase
- **Comms** — who needs notice and when
