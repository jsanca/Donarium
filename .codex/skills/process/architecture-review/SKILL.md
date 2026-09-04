---
name: architecture-review
description: Review software architecture for integrity, boundaries, contracts, and long-term maintainability. Use when evaluating a vertical slice, PR with architectural impact, completed feature, or before architectural approval. Compatible with Go, Java, TypeScript, and multi-language projects.
---

# Architecture Review

## Mission

An architecture review validates that a software system's structure, boundaries, and contracts are correct, maintainable, and aligned with its intended design. It detects architectural drift before it becomes technical debt.

The review is **not** a code review, a style audit, or a functional test. It answers one question: *Will this architecture hold as the system grows?*

## Responsibilities

An Architecture Reviewer evaluates the following architectural properties.

### Dependency Direction

- Do dependencies flow toward stable abstractions?
- Is the direction of imports consistent with the intended layering?
- Does application code depend on domain ports, not infrastructure?
- Are framework, HTTP, database, and transport concerns kept out of domain and application layers?

### Architectural Boundaries

- Are bounded contexts clearly separated?
- Do domain packages have well-defined public surfaces?
- Are internal implementation details unexported or inaccessible?
- Is each module's responsibility singular and coherent?

### Contracts

- Are public APIs, HTTP routes, and domain method signatures stable and intentional?
- Do error responses follow a consistent envelope?
- Are error semantics distinct at each architectural level (domain, application, HTTP)?

### Statelessness

- Does each request carry its own state without hidden server-side session caches?
- Are tokens or sessions verified per-request rather than assumed valid?
- Is mutable shared state absent or explicitly justified?

### Determinism

- Does the same persisted state produce the same observable behavior every time?
- Are database queries ordered when order affects application behavior?
- Is "first" or "default" selection based on an explicit, documented rule?

### Adapter Isolation

- Are persistence, HTTP, messaging, and external service adapters behind interfaces?
- Is infrastructure wiring confined to composition roots (main, DI modules)?
- Do adapters translate domain types to infrastructure types without leaking either?

### Error Propagation

- Are domain errors distinct from infrastructure errors?
- Do application services wrap repository failures appropriately?
- Does the HTTP boundary map errors to correct status codes without exposing internals?

### Testability

- Do ports and adapters have seams for test doubles?
- Are integration points (database, HTTP client, message broker) replaceable in tests?
- Does the design allow testing domain logic without infrastructure?

### Security Implications

- Are credentials, signing keys, and secrets absent from responses and logs?
- Are sessions and tokens verified before trust is extended?
- Is constant-time comparison used for cryptographic verification?

### Extensibility

- Can new functionality be added without modifying existing stable code?
- Is the design open to vertical slices without cross-cutting rewrites?
- Are versioning and migration paths considered for evolving protocols?

## Non-Responsibilities

An Architecture Reviewer does **not** evaluate:

- Formatting, whitespace, or naming preferences
- Coding style or linting rules (unless they mask structural problems)
- Minor refactors or DRY opportunities without architectural impact
- Test coverage percentage (review test *design*, not quantity)
- Performance micro-optimizations
- Personal taste or preferred libraries/frameworks
- UI aesthetics or visual design
- Product roadmap decisions or feature prioritization

## Severity Taxonomy

Every finding receives a severity classification.

| Severity | ID Pattern | Definition | Expected Disposition |
|---|---|---|---|
| **BLOCKER** | AAR-0XX (0-9) | Architecture is incorrect or unsafe. Cannot ship this slice. | Must be resolved in this slice. |
| **MAJOR** | AAR-1XX | Violates an architectural boundary, contract, or agreed rule. | Should be resolved before closeout. |
| **MINOR** | AAR-2XX | Acceptable but increases risk or leaves a coverage gap. | Follow-up in a subsequent slice. |
| **OBSERVATION** | AAR-3XX | Not a defect. Noted for future consideration or awareness. | No action required now. |

### Severity Characteristics

**BLOCKER**: The system behaves incorrectly under known conditions. Examples: credentials exposed in responses, session verification bypassed, data corruption under concurrent access, missing encryption where required.

**MAJOR**: An architectural rule is broken but no runtime defect exists yet. Examples: authorization code in authentication slice, non-deterministic default selection, adapter coupling that prevents test isolation.

**MINOR**: The architecture is correct but incomplete or unverified. Examples: missing router-level integration test, unvalidated configuration edge case, untested error boundary.

**OBSERVATION**: A pattern or concern worth recording for future slices. Examples: session protocol versioning not yet designed, potential key rotation complexity, naming convention inconsistency.

## Review Workflow

Execute each step in order. Record findings as they are discovered.

### Step 1: Understand the Slice

Read the task definition, use case, and related reports. Identify the slice boundary: what was supposed to be built, what was explicitly excluded, and what existing architecture it extends.

### Step 2: Map the Architecture

Trace the execution path from entry point to persistence and back:

```
Entry point → Adapter → Application Service → Domain Ports → Adapter → Infra
```

Identify every component, its dependencies, and which layer it belongs to.

### Step 3: Evaluate Boundaries

For each component, ask:
- Is it in the right layer?
- Does it import only stable abstractions from lower layers?
- Does it expose only what callers need?

### Step 4: Validate Contracts

- Check HTTP routes, request/response shapes, and error envelopes.
- Verify that error codes are appropriate and consistent.
- Confirm that secrets and internal state are absent from responses.

### Step 5: Verify Determinism and Statelessness

- Trace each request path for hidden state, caches, or singletons.
- Check database queries for ordering when ordering matters.
- Verify that "first", "default", or "current" selections are explicit.

### Step 6: Inspect Tests

- Are ports testable with fakes or doubles?
- Do tests exercise architectural boundaries (middleware chains, error paths)?
- Does the test design match the architectural design?

### Step 7: Assess Risks

For each finding, evaluate:
- What breaks if this is wrong?
- How hard is it to fix later?
- Is it already causing problems or only likely to?

### Step 8: Classify Findings

Assign severity using the taxonomy above. Each finding must include:
- **ID** (AAR-XXX)
- **Severity** (BLOCKER / MAJOR / MINOR / OBSERVATION)
- **Description** — what was found
- **Evidence** — file paths, line numbers, patterns
- **Impact** — architectural consequence
- **Recommendation** — concrete fix
- **Disposition** — required now, follow-up, or deferred

### Step 9: Produce Verdict

Apply the approval rules below. Record positive findings alongside issues.

## Approval Rules

### APPROVED

Issue when:
- Zero BLOCKER findings.
- Zero MAJOR findings.
- All MINOR findings have documented follow-up tasks.
- Architecture matches the intended design.
- Boundaries are intact.
- Contracts are consistent.

### CHANGES REQUIRED

Issue when:
- One or more BLOCKER findings exist.
- One or more MAJOR findings exist without a documented resolution plan.
- The architecture contradicts its own stated design.
- A core architectural property (statelessness, direction, determinism) is violated.

### Observations Alone Do Not Block Approval

OBSERVATION findings are informational. They never cause a CHANGES REQUIRED verdict. They are recorded for future slices.

MINOR findings alone do not block approval if follow-up tasks are created and the architecture is otherwise sound.

## Report Structure

Every architecture review report must include:

### Header
- **Reviewer:** Name or role
- **Date:** Review date
- **Scope:** What was reviewed (slice ID, feature, component)
- **Verdict:** APPROVED or CHANGES REQUIRED

### Executive Summary
One paragraph stating whether the architecture is sound and what the primary concerns are.

### Architecture Assessment
Per-property evaluation of the architecture:
- Dependency direction
- Boundaries and layering
- Contracts (HTTP, domain, error)
- Statelessness
- Determinism
- Adapter isolation
- Error propagation
- Testability
- Security

### Use Case Compliance (when applicable)
Table mapping each use-case concern to an assessment of whether it is met.

### Findings
Each finding with: ID, severity, description, evidence, impact, recommendation, disposition.

### Positive Findings
Architectural decisions or patterns that are correct and worth preserving.

### Deferred Findings
Issues identified but explicitly out of scope for this review.

### Risk Assessment
Table of risk dimensions (maintainability, extensibility, security, testability) with brief assessments.

### Recommendation
Clear statement: APPROVED, CHANGES REQUIRED, or FOLLOW-UP REQUIRED. List what must change to move from CHANGES REQUIRED to APPROVED.

### References
Links to related reports, use cases, task definitions, and design documents.

## Output Artifacts

After the review:
1. Write the review report to `docs/agents/reviews/<task-id>-review-<reviewer-role>.md`
2. Create a checkpoint at `docs/agents/checkpoints/CHECKPOINT-<task-id>.md`
3. Add an entry to `docs/agents/ENGINEERING_LOG.md`
4. Do not modify production code, tests, or configuration

## Short Example

```
## Dependency Direction Assessment

The application service imports domain ports (UserRepository, PasswordHasher,
SessionIssuer). It does not import Chi, HTTP, pgx, JSON, or environment access.
Direction is correct.

LoginHandler owns a *pgxpool.Pool and constructs a DBExecutor before calling
the service. This keeps pgx out of application code but makes the HTTP adapter
the composition point — acceptable adapter coupling, not domain contamination.
```

See `examples/approved-review.md` and `examples/changes-required-review.md` for complete annotated review examples.
