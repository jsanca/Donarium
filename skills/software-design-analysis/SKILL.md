---
name: software-design-analysis
description: Analyze software problems, requirements, constraints, and design options before coding. Use when clarifying scope, decomposing features, comparing approaches, defining acceptance criteria, assessing non-functional needs, or preparing input for architecture review, RFCs, or implementation plans.
---

# Software Design Analysis

## Workflow

1. **Frame the problem** — goal, users, success criteria, constraints, and explicit non-goals.
2. **Gather context** — existing system behavior, integrations, team boundaries, timelines, and compliance or security constraints.
3. **Decompose** — capabilities, use cases, or bounded chunks; identify dependencies and sequencing.
4. **Define requirements** — functional needs, non-functional attributes (performance, reliability, security, operability), and measurable acceptance criteria.
5. **Generate options** — at least two viable approaches; avoid false dichotomies (build vs buy, sync vs async, monolith vs split only when relevant).
6. **Analyze tradeoffs** — cost, complexity, risk, time-to-value, operability, and reversibility for each option.
7. **Recommend** — smallest step that meets the goal; call out spikes, prototypes, or decisions needing stakeholder input.
8. **Hand off** — if the deliverable is a written artifact, use `technical-documentation-authoring` for structure and formatting.

## References

- Read `references/problem-framing-and-requirements.md` for problem statements, scope, and requirement quality.
- Read `references/design-options-and-tradeoffs.md` for option analysis, quality attributes, and decision records.

## Analysis Checklist

- Problem statement is specific and testable — not a solution disguised as a requirement.
- Scope lists **in** and **out**; assumptions and unknowns are explicit.
- Non-functional requirements name **metrics or thresholds** where possible (latency, RPO/RTO, audit, retention).
- Each option states who owns what, what changes in production, and how to roll back.
- Recommendation prefers incremental delivery over big-bang rewrites unless constraints forbid it.
- Open questions are assigned to a role (product, security, platform, legal) — not left vague.

## Output

Deliver a design analysis with:

- **Problem and goals** — one paragraph plus bullet success criteria
- **Context** — current state, constraints, stakeholders
- **Requirements** — functional and non-functional (with acceptance criteria)
- **Options considered** — brief description each
- **Tradeoff matrix** — comparison on complexity, risk, time, operability
- **Recommendation** — chosen path, rationale, and first milestone
- **Risks and mitigations** — ordered by likelihood × impact
- **Open questions** — decisions blocking implementation
- **Next steps** — spikes, docs to write (BRD/HLD/RFC/ADR), or reviews to schedule

When the user asked for a **document**, produce content structured per `technical-documentation-authoring` references.

## Related Skills

- `system-architecture-review` — deep architecture boundary and scalability review
- `api-design-review` — HTTP/API contract focus
- `migration-planning` — version and platform migration sequencing
- `technical-documentation-authoring` — plan and doc structure/formatting
