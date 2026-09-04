# Architecture Decision Records

## Purpose

Store durable decisions whose alternatives and consequences matter to future engineers.

## What belongs here

ADRs that record context, decision, and consequences for architectural or long-lived technical choices, such as persistence, boundaries, communication, security, deployment, or orchestration strategy.

**Example:** “Choose PostgreSQL instead of Provider X” with alternatives and consequences is an ADR. “The project uses PostgreSQL” is current knowledge in `../knowledge/`.

## Index

| ADR | Decision |
| --- | --- |
| [ADR-001](ADR-001-installation-bootstrap.md) | Installation bootstrap — explicit first-owner setup flow before login |
| [ADR-002](ADR-002-application-level-bootstrap-invariant.md) | Application-level bootstrap invariant — accepted setup race window |
| [ADR-003](ADR-003-client-routing.md) | Client routing — React Router v7 (`createBrowserRouter`) |
| [ADR-004](ADR-004-property-stakeholder-relation.md) | PropertyStakeholder as a per-property relation (not extending `OrganizationRole`) |
| [ADR-005](ADR-005-party-model.md) | Party model — `UserRef` \| `OrganizationRef` \| `ExternalParty` |

## What does not belong here

Task reports, general documentation, or a restatement of the current architecture. Keep current system understanding in `../knowledge/`.
