# Controls and Verification

## Control Patterns (Smallest Effective Fix)

| Risk | Control |
| --- | --- |
| IDOR on `{resourceId}` | Server-side ownership check on every read/write |
| Credential stuffing | Rate limit + MFA for sensitive accounts |
| Secret in repo | Rotate secret, remove from history, use secret manager |
| Verbose errors | Problem Details with generic `detail` in prod |
| Missing auth on admin route | Default-deny security config + integration test |
| Vulnerable dependency | Upgrade, pin, or virtual patch with documented exception |
| SSRF on webhook URL | Allowlist domains, block metadata IPs, no raw redirects |

## Authorization Test Matrix

Document expected result for each role:

```text
| Action              | Anonymous | User (owner) | User (other) | Admin |
|---------------------|-----------|--------------|--------------|-------|
| GET /orders/{id}    | 401       | 200          | 403/404      | 200   |
| DELETE /orders/{id} | 401       | 403          | 403          | 200   |
```

Automate at least one negative case per protected endpoint class.

## Verification Commands (Examples)

```bash
# Dependency scan (Java)
./gradlew dependencyCheckAnalyze

# Container scan (example)
trivy image registry.example.com/app:tag

# Contract/security smoke
curl -i -X DELETE "https://staging.example.com/api/v1/orders/other-users-id" \
  -H "Authorization: Bearer ${USER_TOKEN}"
```

## Finding Severity Guide

- **Critical** — exploitable remotely with high impact, no strong compensating control
- **High** — exploitable with conditions or high impact on confidentiality/integrity
- **Medium** — defense-in-depth gap or requires chained preconditions
- **Low** — hardening, informational, or best practice

Always pair severity with **exploit scenario** — not just checklist presence.
