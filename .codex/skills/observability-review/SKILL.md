---
name: observability-review
description: Review logging, metrics, traces, dashboards, alerts, and SLOs for production operability. Use when preparing a service for production, debugging unclear incidents, reducing alert noise, or designing SLOs and on-call runbooks.
---

# Observability Review

## Workflow

1. Identify critical user journeys and system dependencies for the service under review.
2. Assess the three pillars: **logs** (structured, searchable), **metrics** (aggregatable signals), **traces** (request flow across components).
3. Map golden signals — latency, traffic, errors, saturation — per service and key endpoints.
4. Review existing dashboards and alerts: actionable, owned, and linked to runbooks.
5. Propose SLOs/SLIs where missing; align alert thresholds to user impact, not arbitrary CPU%.
6. Check correlation: trace IDs in logs, consistent service/resource attributes, and safe log content.
7. Recommend smallest additions to close blind spots before the next incident.

## References

- Read `references/logs-metrics-traces.md` for signal quality and correlation patterns.
- Read `references/slos-and-alerting.md` for SLO design and alert hygiene.

## Review Checklist

- Logs are structured (JSON or parseable fields) with level appropriate to environment.
- Request/correlation/trace IDs propagate across sync and async boundaries.
- Metrics cover rate, errors, and duration for critical operations.
- Dashboards exist for on-call triage within one click from alert.
- Alerts are symptomatic (user impact) where possible — not only cause-based noise.
- Sensitive data is not logged.
- Sampling and cardinality are controlled for high-cardinality labels.

## Output

Deliver an observability review with:

- **Critical paths** reviewed
- **Gaps** — what cannot be seen today during an incident
- **Recommendations** — logs, metrics, traces, dashboards, alerts (prioritized)
- **SLO proposals** — SLI, target, error budget policy (if applicable)
- **Verification** — how to validate signals after implementation
