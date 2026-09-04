# Pipeline and Platform Standards

## Reference Pipeline Stages

```text
1. Source checkout (pinned commit)
2. Lint / static analysis
3. Unit tests
4. Integration tests (with service containers if needed)
5. Build artifact (image or package)
6. Vulnerability scan (deps + image)
7. Publish immutable tag to registry
8. Deploy to staging (automated)
9. Smoke / contract verification
10. Promote to production (gated)
11. Post-deploy metrics check (automated or runbook)
```

Teams may skip stages only with documented exception and compensating control.

## Platform Checklist (Kubernetes-Oriented)

- [ ] Namespace per env/service class with RBAC
- [ ] Resource requests/limits set; HPA where appropriate
- [ ] Ingress/TLS terminated consistently
- [ ] Secrets via External Secrets or cloud provider — not ConfigMap literals
- [ ] NetworkPolicy for sensitive tiers (optional but recommended)
- [ ] PodDisruptionBudget for critical services
- [ ] Centralized logging and metrics agents

Adapt checklist for ECS, Cloud Run, or VM deploy — same principles apply.

## Golden Pipeline Template

Provide a copyable pipeline template repo or workflow file new services fork:

- Standard job names and artifact conventions
- Required scan tools
- Deploy hooks calling platform API
- Smoke test script interface (`SMOKE_BASE_URL`, `SMOKE_TOKEN`)

## Measuring Delivery Maturity

Track over time:

- Deploy frequency
- Lead time for changes
- Change failure rate
- Mean time to restore

Use these metrics to prioritize pipeline investment — not vanity Dockerfile counts.
