# SKILL-001 — Create Architecture Review Skill

- **Status:** Complete
- **Type:** Skill creation
- **Owner:** Deep (Engineering Process Engineer)

## Objective

Create a reusable Architecture Review skill that standardizes how architectural reviews are performed across all projects (Java, Go, TypeScript, and multi-language).

## Scope

- Create `.claude/skills/process/architecture-review/SKILL.md` with full methodology
- Create annotated example reviews (APPROVED and CHANGES REQUIRED)
- Create README with invocation and relationship documentation
- Copy skill to `.agents/`, `.opencode/`, and `.codex/` skill directories

## Deliverables

| File | Description |
|---|---|
| `.claude/skills/process/architecture-review/SKILL.md` | Main skill definition |
| `.claude/skills/process/architecture-review/README.md` | Skill overview and invocation guide |
| `.claude/skills/process/architecture-review/examples/approved-review.md` | Annotated APPROVED review example |
| `.claude/skills/process/architecture-review/examples/changes-required-review.md` | Annotated CHANGES REQUIRED review example |
| `.agents/skills/process/architecture-review/` | Mirror copy |
| `.opencode/skills/process/architecture-review/` | Mirror copy |
| `.codex/skills/process/architecture-review/` | Mirror copy |

## Skill Contents

### Mission
Validates architectural integrity, protects long-term maintainability, detects architectural drift, and verifies intended boundaries.

### Responsibilities
Covers 10 architectural properties: dependency direction, boundaries, contracts, statelessness, determinism, adapter isolation, error propagation, testability, security implications, and extensibility.

### Non-Responsibilities
Explicitly excludes: formatting, naming, coding style, minor refactors, coverage percentages, micro-optimizations, personal taste, UI aesthetics, and product roadmap decisions.

### Findings Classification
Four severity levels: BLOCKER (AAR-0XX), MAJOR (AAR-1XX), MINOR (AAR-2XX), OBSERVATION (AAR-3XX). Each with clear characteristics and expected dispositions.

### Review Workflow
Nine-step repeatable process: Understand Slice → Map Architecture → Evaluate Boundaries → Validate Contracts → Verify Determinism/Statelessness → Inspect Tests → Assess Risks → Classify Findings → Produce Verdict.

### Approval Rules
APPROVED: zero BLOCKER, zero MAJOR, MINORs have follow-ups. CHANGES REQUIRED: one or more BLOCKER or MAJOR without resolution. OBSERVATIONs never block.

### Report Structure
Standardized sections: Header, Executive Summary, Architecture Assessment, Use Case Compliance, Findings, Positive Findings, Deferred Findings, Risk Assessment, Recommendation, References.

### Examples
Two complete project-agnostic example reviews demonstrating both verdicts with realistic scenarios.

## Files Changed

| File | Change |
|---|---|
| `.claude/skills/process/architecture-review/SKILL.md` | Created |
| `.claude/skills/process/architecture-review/README.md` | Created |
| `.claude/skills/process/architecture-review/examples/approved-review.md` | Created |
| `.claude/skills/process/architecture-review/examples/changes-required-review.md` | Created |
| `.agents/skills/process/architecture-review/*` | Mirrored |
| `.opencode/skills/process/architecture-review/*` | Mirrored |
| `.codex/skills/process/architecture-review/*` | Mirrored |
| `docs/agents/tasks/SKILL-001-architecture-review-skill.md` | Created |
| `docs/agents/checkpoints/CHECKPOINT-SKILL-001-architecture-review-skill.md` | Created |
| `docs/agents/ENGINEERING_LOG.md` | Updated |

## Validation

All 16 files (4 files x 4 directories) verified identical across `.claude/`, `.agents/`, `.opencode/`, and `.codex/`.

No code was modified. No commits were made.
