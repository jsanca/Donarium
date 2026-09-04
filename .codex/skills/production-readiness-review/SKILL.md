---
name: production-readiness-review
description: Review release and production readiness for operability, rollback, monitoring, capacity, runbooks, and launch verification. Use before major launches, production cutovers, on-call handoffs, or go/no-go checkpoints.
---

# Production Readiness Review

## Workflow

1. Define the release: feature launch, infra migration, version upgrade, or new service.
2. Confirm functional completeness — separate "works in dev" from "safe in production."
3. Review operability: health checks, metrics, alerts, logs, runbooks, and on-call ownership.
4. Validate deployment mechanics: CI/CD, immutable artifacts, config/secrets, migrations, and rollback path.
5. Assess capacity and failure modes: load expectations, autoscaling, dependency limits, and blast radius.
6. Verify security and compliance gates required for this release class.
7. Record go/no-go criteria, open blockers, and post-launch monitoring plan.

## References

- Read `references/readiness-checklist.md` for go/no-go categories and blockers.
- Read `references/launch-and-rollback.md` for launch steps, rollback, and hypercare.

## Readiness Categories

- **Functionality** — acceptance criteria met; known gaps documented
- **Reliability** — SLO impact understood; timeouts and retries configured
- **Observability** — dashboards, alerts, and log queries ready
- **Security** — review complete; secrets rotated if needed
- **Data** — migrations tested; backup/restore validated if high risk
- **Operations** — runbook, owner, escalation path defined
- **Rollback** — revert procedure tested or explicitly not possible with mitigation

## Output

Produce a production readiness review with:

- **Release summary** and target date/environment
- **Go / no-go** recommendation with blockers
- **Checklist results** per category (pass / fail / N/A)
- **Rollback plan** — steps, time estimate, data implications
- **Launch verification** — smoke tests and metrics to watch
- **Hypercare** — duration, owner, daily check schedule
