# Production Readiness Checklist

## Functionality

- [ ] Acceptance criteria signed off with product/owner
- [ ] Feature flags or dark launch path if partial rollout
- [ ] Known defects documented with severity and waiver approver

## Reliability and Performance

- [ ] Load or soak test run for expected peak (or waiver documented)
- [ ] Timeouts and circuit breakers on external dependencies
- [ ] Rate limits for public endpoints where abuse is possible
- [ ] Database migration duration estimated; lock impact understood

## Observability

- [ ] Golden signals dashboard (latency, traffic, errors, saturation)
- [ ] Alerts routed to on-call with runbook links
- [ ] Structured logs include correlation/request IDs
- [ ] Tracing enabled for new cross-service paths

## Security and Compliance

- [ ] Security review complete for release class
- [ ] Secrets in secret manager — not env files in image
- [ ] Audit logging for admin/sensitive actions if required

## Operations

- [ ] On-call owner named for launch window
- [ ] Runbook covers deploy, verify, rollback, common failures
- [ ] Support/comms informed of user-visible changes

## Rollback

- [ ] Previous artifact version pinned and deployable
- [ ] DB migration rollback strategy documented (forward-fix vs revert)
- [ ] Rollback tested in staging or drill completed

Mark **blocker** on any item that would cause data loss, security exposure, or unrecoverable outage without mitigation.
