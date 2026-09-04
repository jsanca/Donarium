# Roadmap

This document records committed project direction. Keep proposals and uncommitted ideas under `future/` rather than presenting them as planned work.

Donarium se desarrolla por experiencias y vertical slices observables. La
secuencia evita organizar el trabajo por infraestructura: cada paso deja al
producto en un estado útil y prepara el siguiente límite de dominio.

## Current Commitments

| Item | Outcome | Status | Target | References |
| --- | --- | --- | --- | --- |
| 1. Platform Foundation | Base platform available to receive domain slices (runtime composition + lifecycle capabilities) | Done | — | DON-003…DON-005 (see `docs/agents/`) |
| 2. UC-001 — Initialize Donarium | First-owner setup: User, Credentials, Organization, Membership (`OWNER`), PlatformGrant (`SUPER_ADMIN`) | Done | — | [UC-001](knowledge/use-cases/UC-001-initialize-donarium.md) |
| 3. Authentication | Authenticate Users, issue signed session tokens, recover access | Done | — | [UC-002](knowledge/use-cases/UC-002—authenticate-user.md) |
| 4. Organization | Manage the organization, its memberships and contextual roles | Not started | — | — |
| 5. Properties | Represent properties and rental units (First Property Experience under delivery) | In progress | EP-001 | [plan](engineering/plans/DONARIUM-PA-EXP-001-first-property-experience-engineering-plan-proposal.md) |
| 6. Leases | Establish and understand occupancy agreements | Not started | — | — |
| 7. Payments | Show obligations, record payments, communicate balances | Not started | — | — |

## Recently Completed

- **UC-001 — Initialize Donarium** (`Organizations`, `Memberships`, `PlatformGrants`, `Credentials`) — the first operator flow preceding login; see `docs/agents/reports/DON-007.5-vertical-integration-validation.md`.
- **Authentication** (`POST /api/auth/login`, `GET /api/auth/me`) — see `docs/agents/reports/DON-008.3-authorization-context-and-route-guards.md`.
- **First Property Experience** — EP-001 plan approved r3; nodes EP-001.01/.02/.03 delivered and reviewed; see `docs/engineering/ENGINEERING_LOG.md`.
