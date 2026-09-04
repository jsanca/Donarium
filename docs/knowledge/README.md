# Project Knowledge

## Purpose

Store durable, current understanding of what the project is, how its domain works, and the concepts needed to change it safely. This is an entry point, not an exhaustive index: navigate from this root to a relevant area, then to the concept and its authority or evidence.

## Navigation

- Start with [`docs/PROJECT.md`](../PROJECT.md) for concise project context, then the workspace guide [`docs/OSK.md`](../OSK.md) for where new information belongs.
- Current system structure: [`architecture/`](architecture/index.md).
- Decisions and rationale: [`docs/adr/`](../adr/README.md).
- Delivery history / engineering evidence: [`docs/engineering/`](../engineering/index.md) (indexed by `ENGINEERING_LOG.md`).
- Committed direction: [`docs/roadmap/ROADMAP.md`](../roadmap/ROADMAP.md).

## Areas

| Area | Contents |
| --- | --- |
| [`architecture/`](architecture/index.md) | Current package structure (server/client), boundaries, database entity relationships |
| [`use-cases/`](use-cases/) | UC-001 Initialize, UC-002 Authenticate |
| [`business-rules.md`](business-rules.md) | BR-01…BR-09 (bootstrap-scoped normative rules) |
| [`design/`](design/) | Design Constitution + wireframes |
| [`test-cases/`](test-cases/) | QA test-case designs |
| [`testing/`](testing/) | QA environment notes |

## What belongs here

Domain definitions, actors, entities, workflows, business rules, terminology, conceptual architecture, external-system relationships, and other enduring project concepts.

**Example:** A durable business rule ("an Order has three terminal states") lives here; the task report that discovered it stays in `../engineering/`.

## What does not belong here

Task reports, implementation logs, review results, temporary investigation notes, Architecture Decision Records (those live in `../adr/`), or unverified conclusions (keep those in `../engineering/` or `../roadmap/future/`).

## Updating knowledge

- Prefer updating an existing canonical page over adding a parallel page.
- Link durable claims to their authority (a source location, ADR, or engineering evidence).
- When a change reveals a reusable current fact, update the canonical concept page and link the relevant report/review/ADR.
- Retain *why* in an ADR; retain *what happened* in `../engineering/`.
