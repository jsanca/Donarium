# SLOs and Alerting

## SLI / SLO Example

```text
SLI: proportion of successful HTTP requests to POST /api/v1/orders
     completed in < 800ms over 30-day window

SLO: 99.5% of requests meet SLI

Error budget: 0.5% ≈ 3.6 hours/month of bad minutes at steady traffic
```

Define SLOs for user-visible paths — not every internal admin endpoint.

## Alert Quality Rules

| Good alert | Bad alert |
| --- | --- |
| Pages when users likely impacted | Pages on every deploy blip |
| Has runbook link | "Something looks weird" |
| Clear severity and owner | Duplicates across 3 tools |
| Uses SLO burn rate or error budget | Static CPU > 70% always |

## Symptomatic vs Cause-Based

- **Symptomatic**: checkout error rate > 2% for 5m — on-call investigates
- **Cause-based**: pod restarted — notify only if paired with user impact or repeated

Reduce noise by routing cause-based signals to dashboards or low-priority channels.

## Runbook Link Template

```text
Alert: OrdersAPIHighErrorRate
Dashboard: https://grafana.example.com/d/orders-api
Runbook: docs/runbooks/orders-api-errors.md
First steps:
  1. Check recent deploys and feature flags
  2. Inspect dependency health (payments, DB)
  3. Sample failing trace IDs from APM
Escalate: #payments-oncall if provider errors dominate
```

## Post-Incident Observability Debt

After each incident, ask:

- Which question took longest to answer?
- Which signal was missing?
- Which alert fired too late or not at all?

Track as observability backlog items with owners.
