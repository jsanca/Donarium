# Task Briefing for AI Agents

## Effective Task Template

```markdown
## Goal
Fix N+1 query on GET /api/v1/orders for customer list.

## Context
- Spring Boot 4, Java 25, Gradle, JPA
- Follow patterns in OrderController and OrderRepository
- Do not change API response shape

## Constraints
- No new dependencies without justification
- Do not enable open-in-view as the fix
- No logging of customer email or payment data

## Verification
./gradlew test --tests '*Order*'
./gradlew check
```

## Scope Sizing

| Good scope | Too large |
| --- | --- |
| Add validation to one DTO | "Refactor entire backend" |
| Fix one failing test with root cause | "Make app faster" |
| Implement one endpoint with tests | "Build admin portal" |

Split large work:

1. Plan or spike (read-only exploration)
2. Implementation slice with tests
3. Follow-up hardening

## Context to Include

- **Build/runtime**: Java/Node version, framework, package roots
- **Patterns**: "match existing controller → service → repository"
- **Files**: starting points and files not to touch
- **Non-goals**: what to defer explicitly

## Context to Exclude

- Pasting entire codebases when grep/read targets suffice
- Secrets, tokens, production URLs with credentials
- Vague "best practices" without repo-specific anchor

## When to Switch Modes

- **Explore first** when root cause or architecture is unknown
- **Implement** when acceptance criteria and patterns are clear
- **Review only** when user did not ask for code changes
