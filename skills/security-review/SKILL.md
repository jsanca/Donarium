---
name: security-review
description: Perform cross-stack security reviews of applications, APIs, data handling, authentication, dependencies, and deployment configuration. Use when reviewing PRs with security impact, preparing for release, responding to audit findings, or threat modeling a feature.
---

# Security Review

## Workflow

1. Define scope: feature, service, endpoint set, infra change, or full application surface.
2. Identify assets (data, credentials, availability) and trust boundaries (public internet, internal network, admin, third parties).
3. Walk STRIDE-style threats at boundaries: spoofing, tampering, repudiation, information disclosure, denial of service, elevation of privilege.
4. Review authentication, authorization, input validation, output encoding, secrets, logging, and dependency exposure.
5. Check deployment and config: TLS, CORS, headers, IAM, network policy, and secret storage.
6. Prioritize findings by exploitability and impact — not alphabetical order.
7. Recommend fixes with smallest effective control first; note compensating controls where full fix is deferred.

## References

- Read `references/threat-model-checklist.md` for boundary questions and STRIDE prompts.
- Read `references/controls-and-verification.md` for common controls and verification steps.

## Review Checklist

- Every protected action requires authentication and explicit authorization.
- Input validated at trust boundaries; output encoded where rendered.
- Secrets not in source, logs, URLs, or client-visible errors.
- Dependencies scanned; critical CVEs have upgrade or mitigation plan.
- Rate limiting and abuse cases considered for public endpoints.
- Sensitive data classified; retention and access follow least privilege.
- Security regressions have tests or repeatable verification where feasible.

## Output

Produce a security review with:

- **Scope and assumptions**
- **Findings** — severity (critical/high/medium/low), location, exploit scenario, recommendation
- **Quick wins** — fixes shippable immediately
- **Deferred items** — with compensating controls and owner
- **Verification** — tests, scans, manual steps to confirm fixes

Do not include live secrets, credentials, or unmasked regulated personal data in the review output.
