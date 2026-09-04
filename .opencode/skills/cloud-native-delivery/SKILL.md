---
name: cloud-native-delivery
description: Review and design cross-stack cloud delivery architecture including environments, promotion, containers, CI/CD, infrastructure boundaries, and platform concerns. Use when planning deployment topology, standardizing pipelines across services, or reviewing org-wide delivery practices — complementing ecosystem-specific delivery skills.
---

# Cloud Native Delivery

## Workflow

1. Clarify delivery context: team size, service count, runtimes (Java, Node, .NET, etc.), and target platform (Kubernetes, PaaS, VMs).
2. Map environments: dev, CI, staging, production — and promotion rules between them.
3. Review artifact strategy: container images, language-specific packages, infra-as-code, and immutable versioning.
4. Assess pipeline stages: test gates, security scan, deploy approval, smoke verification, and rollback.
5. Evaluate platform concerns: ingress, secrets, config, autoscaling, network policy, and multi-region needs.
6. Align with twelve-factor practices without mandating one cloud vendor.
7. Recommend standards and exceptions — document what is org-wide vs per-service.

## References

- Read `references/environments-and-promotion.md` for env topology and promotion rules.
- Read `references/pipeline-and-platform-standards.md` for CI/CD stages and platform checklist.

## Delivery Checklist

- Every service has a defined build artifact and immutable tag strategy.
- Secrets are injected at deploy time — not baked into artifacts.
- Staging mirrors production constraints sufficiently for meaningful verification.
- Health/readiness semantics are consistent across services.
- Deployments are automated and auditable; manual SSH deploy is exception-only.
- Rollback procedure exists and is practiced.
- Platform limits (CPU, memory, connections, quotas) are documented per service class.

## Output

Deliver a cloud delivery review or plan with:

- **Current topology** — environments, pipelines, platform
- **Gaps** — consistency, security, operability
- **Standards** — recommended org defaults and allowed exceptions
- **Roadmap** — prioritized improvements
- **Verification** — how teams prove compliance (checks, templates, golden pipelines)

For JVM-specific container and Actuator guidance, also use `java-cloud-native-delivery`.

After delivery architecture work, summarize standards adopted, exceptions granted, and next implementation steps per team or service.
