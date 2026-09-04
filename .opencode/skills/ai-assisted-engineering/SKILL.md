---
name: ai-assisted-engineering
description: Use AI coding agents effectively and safely in day-to-day engineering. Use when delegating implementation to an agent, reviewing AI-generated code, structuring tasks for reliable outcomes, or establishing team practices for AI-assisted development.
---

# AI Assisted Engineering

## Workflow

1. Clarify the task outcome: bug fix, feature, refactor, review, or spike — and what "done" means (tests, commands, acceptance criteria).
2. Provide context the agent cannot infer: repo conventions, build commands, constraints, files to avoid, and security boundaries.
3. Scope work in small, verifiable steps — one behavior change per iteration when risk is high.
4. Require verification before trust: run tests, lint, build, and manual checks on critical paths.
5. Review AI output for correctness, security, scope creep, and missing edge cases — treat output as a draft from a junior contributor.
6. Capture repeatable patterns in skills, rules, or docs when the same guidance is needed every session.
7. Never commit secrets, disable auth, or skip tests because the agent suggested speed over safety.

## References

- Read `references/task-briefing.md` for how to write effective agent prompts and task scopes.
- Read `references/review-and-verification.md` for reviewing AI-generated code and defining done.

## Practice Checklist

- Task includes expected verification command (`make test`, `./gradlew check`, etc.).
- Sensitive paths (auth, payments, PII) are called out explicitly.
- Agent changes are diff-reviewed before merge — not blind accept.
- Regression tests exist for bug fixes and security-sensitive behavior.
- Large refactors are split into reviewable PRs.
- Team rules/skills encode repo-specific conventions agents forget.

## Output

After AI-assisted work, summarize:

- Task goal and scope enforced
- Changes made and files touched
- Verification run (commands and results)
- Issues found in AI output and how they were corrected
- Follow-ups to codify in rules, skills, or tests
