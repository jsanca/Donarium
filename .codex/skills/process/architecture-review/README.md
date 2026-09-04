# Architecture Review Skill

## Purpose

Standardize how architectural reviews are performed across projects. This skill captures the review methodology refined during Arbitrier, Donarium, and related projects.

## Invocation

To invoke this skill, reference it by name:

```
Apply: .claude/skills/process/architecture-review/SKILL.md
```

The skill is loaded automatically when a task description includes "architecture review" or when the reviewer role (e.g., Elito) is assigned to a review task.

## Expected Inputs

The reviewer needs access to:

- **Task definition** — the slice or feature being reviewed (`docs/agents/tasks/<task-id>.md`)
- **Use cases** — relevant use-case documents (`knowledge/use-cases/`)
- **Design documents** — architecture decisions, domain models (`knowledge/design/`)
- **Source code** — the implementation under review
- **Related reports** — prior implementation reports and reviews

## Expected Outputs

1. **Architecture review report** at `docs/agents/reviews/<task-id>-review-<reviewer-role>.md`
   - Verdict (APPROVED or CHANGES REQUIRED)
   - Findings with severity, evidence, and recommendations
   - Risk assessment

2. **Recovery checkpoint** at `docs/agents/checkpoints/CHECKPOINT-<task-id>.md`

3. **Engineering log entry** appended to `docs/agents/ENGINEERING_LOG.md`

## Relationship with Other Skills

| Skill | Relationship |
|---|---|
| `engineering-reporting` | Architecture reviews use the review report template and follow its evidence rules. |
| `execution-timebox` | Reviews must stop at 45 minutes and produce a recovery checkpoint. |
| `security-review` | Complements architecture review; security reviews can reference architecture findings. |

## Review Constraints

- **Do not modify production code.** The review evaluates existing code without changing it.
- **Do not modify tests.** Tests are inspected, not rewritten.
- **Do not commit.** Reviews produce reports and checkpoints only.
- **Respect the timebox.** See `execution-timebox` skill for hard-stop rules.

## Directory Structure

```
architecture-review/
├── SKILL.md                        # Full skill definition
├── README.md                       # This file
└── examples/
    ├── approved-review.md           # Annotated APPROVED review example
    └── changes-required-review.md   # Annotated CHANGES REQUIRED review example
```
