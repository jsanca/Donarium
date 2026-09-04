# Migration Patterns and Sequencing

## Pattern Selection

| Pattern | When it fits | Risk |
| --- | --- | --- |
| **Big bang** | Small system, short maintenance window, strong tests | High blast radius |
| **Phased module** | Monolith with clear modules; one deploy unit | Medium — discipline required |
| **Strangler fig** | Replace capability incrementally behind facade/router | Lower per step; routing complexity |
| **Parallel run** | Compare outputs (billing, search indexing) before cutover | Cost + reconciliation work |
| **Feature flag cutover** | User-visible behavior toggle with metrics | Flag debt if not removed |

## Strangler Example (Text)

```text
Phase 1: Route GET /reports to new service; POST still on legacy
Phase 2: Move POST with read-only parity checks on shadow traffic
Phase 3: Migrate historical data; reconcile counts nightly
Phase 4: Decommission legacy /reports after 30d zero traffic
```

## Increment Sizing Rules

Each increment should:

- Ship to production independently
- Have automated verification
- Reduce legacy surface area OR de-risk the next step
- Fit in one sprint or less when possible

Avoid "Phase 1: build entire new platform" with no production value.

## Dependency Order

Typical order (adjust per context):

1. Observability and feature flags
2. Contract tests and consumer inventory
3. Non-critical read paths
4. Writes with idempotency and reconciliation
5. Data migration and decommission

## Ecosystem-Specific Skills

Use version/framework skills for implementation detail:

- `java-migrate-any-version`, `react-migrate-any-version`, etc.

This skill owns **program sequencing and risk** — not language-specific upgrade commands.
