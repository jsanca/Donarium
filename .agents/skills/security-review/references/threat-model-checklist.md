# Threat Model Checklist

## Trust Boundaries to Map

```text
[Untrusted client] -> [API] -> [Service layer] -> [Database]
                      -> [Third-party API]
[Admin user] -> [Admin API] -> [Production data]
[CI pipeline] -> [Deploy target] -> [Secrets store]
```

## STRIDE Prompts (Per Boundary)

| Category | Question |
| --- | --- |
| Spoofing | Can an attacker impersonate a user or service? |
| Tampering | Can requests or data be modified in transit or at rest? |
| Repudiation | Are security-relevant actions auditable? |
| Information disclosure | Can errors, logs, or APIs leak sensitive data? |
| Denial of service | Can an unauthenticated user exhaust resources? |
| Elevation of privilege | Can a low-privilege actor access admin or other tenants' data? |

## High-Risk Areas (Always Check)

- Authentication and session/token handling
- Authorization on every mutating path (including IDOR via `{id}` routes)
- File upload, import, and export features
- Webhooks and callback URLs (SSRF risk)
- Admin and support tooling
- Background jobs acting with elevated credentials
- Multi-tenant data isolation

## OWASP-Aligned Quick Pass

- Injection (SQL, command, template, LDAP)
- Broken authentication / session management
- Sensitive data exposure
- XML/JSON deserialization issues
- Security misconfiguration (default creds, debug endpoints in prod)
- Missing function-level access control
- Cross-site issues for browser-rendered content

Ecosystem-specific skills (e.g. `java-security-hardening`) cover implementation depth — this skill focuses on review framing and cross-cutting controls.
