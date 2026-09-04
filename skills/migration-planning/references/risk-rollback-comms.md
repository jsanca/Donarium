# Risk, Rollback, and Communications

## Risk Register Template

```text
| Risk                         | L | I | Mitigation                          | Owner   |
|------------------------------|---|---|-------------------------------------|---------|
| API break for mobile clients | M | H | Contract tests + min version banner | API team|
| Migration runtime > window   | M | H | Rehearse on prod-sized snapshot     | DBA     |
| Team unfamiliar with new stack | H | M | Spike + pairing in phase 0          | Eng mgr |
```

L = likelihood, I = impact

## Rollback vs Forward-Fix

Document per phase:

- **Rollback**: redeploy previous artifact; DB revert script if safe
- **Forward-fix**: required when migration is irreversible — must have playbook before cutover

Never promise "we can always roll back" for destructive schema changes.

## Parity Verification

For parallel run or strangler:

```text
- Sample N requests per hour; diff response hash or key fields
- Alert if mismatch rate > 0.1%
- Log diff details with correlation ID (no sensitive payload)
```

## Stakeholder Comms

| Audience | Needs |
| --- | --- |
| Engineering | Phase scope, dates, breaking changes |
| Product | User-visible changes, flag schedule |
| Support | Known issues, workaround, escalation |
| External API consumers | Deprecation timeline, migration guide |

## Decommission Criteria

Legacy path may be removed when:

- Traffic = 0 for agreed period (e.g. 30 days)
- Data reconciled and backups verified
- No open Sev1/2 defects on new path
- Runbook updated and on-call trained
