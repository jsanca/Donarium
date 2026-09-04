# Environments and Promotion

## Typical Environment Ladder

```text
Developer local -> CI (ephemeral) -> Shared staging -> Production
                      |
                 Preview env (optional, per PR)
```

## Promotion Rules (Example Policy)

| From | To | Requirements |
| --- | --- | --- |
| CI | Staging | Unit + integration pass; image scanned |
| Staging | Production | Smoke pass; approval from release owner |
| Hotfix branch | Production | Expedited review + post-deploy verification |

## Configuration by Environment

```text
dev:      local overrides allowed; synthetic data
staging:  prod-like resources; anonymized or synthetic data
prod:     secrets from manager; no debug endpoints
```

Document which config keys differ per environment and who approves prod changes.

## Multi-Service Coordination

When deploying several services:

- Define deploy order if contracts changed (consumer-first vs provider-first)
- Use compatibility windows for API breaks
- Coordinate feature flags across services when needed

## Anti-Patterns

- Staging that skips auth or uses toy datasets — false confidence
- Production-only config drift discovered during incident
- `:latest` as sole production tag across all services
- Manual config edits on servers without IaC or audit trail
