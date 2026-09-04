# Review and Verification

## Review AI Code Like a PR

Check in order:

1. **Scope** — only requested files/behavior changed?
2. **Correctness** — logic matches requirements and edge cases?
3. **Security** — auth, injection, secrets, PII logging?
4. **Tests** — meaningful assertions, not smoke-only?
5. **Conventions** — matches repo naming, layers, and build tools?
6. **Dependencies** — justified and aligned with existing stack?

## Red Flags in AI Output

- Disabled security, validation, or tests "temporarily"
- Hardcoded API keys or credentials
- Generic tutorial code ignoring project structure
- Huge refactors bundled with one-line bug fix
- Comments attributing AI tools in commits or user-facing text
- `catch (Exception e) {}` or swallowed failures

## Verification Ladder

```text
1. Targeted unit/integration test for changed behavior
2. Module or project test command
3. Lint / static analysis if configured
4. Manual smoke for UI or API contract change
5. Staging verification for high-risk releases
```

Always run at least step 1–2 before marking done.

## Documenting Recurring Fixes

When the same agent mistake repeats (wrong JDK, wrong import namespace, wrong test style):

- Add a rule under `.cursor/rules/` or team skill
- Add a regression test if behavior was wrong in production path
- Shorten future prompts by pointing to the rule/skill path

## Human Accountability

AI assists; humans merge. The engineer who merges owns production behavior, security, and operability.
