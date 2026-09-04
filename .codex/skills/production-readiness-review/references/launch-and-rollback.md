# Launch and Rollback

## Launch Sequence (Example)

```text
T-24h  Freeze unrelated changes; confirm rollback artifact
T-1h   Notify stakeholders; verify staging smoke
T0     Deploy canary / 10% traffic (if supported)
T+15m  Check error rate, latency, business KPI vs baseline
T+30m  Full rollout if metrics within threshold
T+2h   Hypercare monitoring active
T+24h  Hypercare exit review
```

## Smoke Tests (Minimum)

```bash
curl -fsS "${BASE_URL}/health/readiness"
curl -fsS -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/critical-path"
```

Add one write/read cycle on a non-production-safe path in staging before prod.

## Rollback Triggers (Define Before Launch)

- Error rate > X% for Y minutes vs baseline
- p95 latency > X ms on checkout (example business path)
- Failed migration or data integrity check
- Critical security incident linked to release

## When Rollback Is Not Safe

Schema migrations that drop columns or irreversibly transform data require:

- Forward-fix playbook instead of revert
- Extended hypercare and backup verification before launch
- Explicit go/no-go approver for migration class

Document this in the readiness review — do not imply rollback is always one click.

## Hypercare

- Named primary and secondary owner for 24–72 hours
- Daily summary: incidents, metrics, open defects
- Decision point to remove flags or complete rollout
