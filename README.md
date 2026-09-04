# Donarium

> A calm, clear, and trustworthy home for rental management.

## Vision

Donarium is a modern platform for managing rental homes with care for both the
people who operate them and the people who live in them. It helps independent
landlords, property managers, organizations, and tenants understand what
matters, act with confidence, and spend less time navigating administrative
friction.

## What is Donarium?

Donarium is a product for the complete rental experience: organizing a
portfolio, understanding a property, establishing a lease, managing payment
obligations, resolving maintenance needs, and keeping important records within
reach.

It is not conceived as another CRUD dashboard. The product should make a
complex human and operational domain feel legible, considered, and dependable.

## Why it exists

Rental management often forces people into fragmented tools, unclear status,
and impersonal workflows. Donarium exists to replace that friction with a
single experience that respects the significance of a home and the
responsibilities attached to it.

## Core Principles

- **Clarity over clutter.** Important state and next actions should be easy to understand.
- **Simplicity with depth.** Workflows stay focused without hiding consequential detail.
- **A human visual experience.** The product feels warm, calm, trustworthy, and contemporary.
- **Domain-led design.** Language and real user outcomes shape the product.
- **Accessible by default.** Every experience is designed for broad, practical use.

## Project Status

**Initial Owner Setup — implemented, with follow-up hardening required.** The
repository contains the platform foundation, a visual login experience, and the
first backend vertical slice: initialization of the first owner, organization,
membership, credentials, and platform grant. The architecture review identified
bootstrap-concurrency, documentation-alignment, and PostgreSQL-test isolation
follow-ups that must be resolved before Authentication begins.

## Architecture

Donarium is a modular monolith organized around the rental domain. The Go
backend exposes initial setup through a thin Chi HTTP layer, an application
service and explicit transaction boundary, identity repositories backed by
pgx, and PostgreSQL migrations. The React client is a feature-oriented SPA.
New capabilities continue to arrive as end-to-end vertical slices.

## Repository Layout

```text
.
├── client/          # React 19, TypeScript, Vite and Tailwind SPA
├── server/          # Go modular monolith and PostgreSQL migrations
├── knowledge/       # Product, use-case, rule and design records
├── docs/agents/     # Engineering tasks, reports, reviews and checkpoints
├── AGENTS.md        # Working agreement and developer commands
├── README.md        # Product vision and project entry point
└── ROADMAP.md       # Experience-oriented delivery path
```

The active Identity module contains the Initial Owner Setup slice. Other domain
directories are intentionally placeholders until their corresponding slice
begins.

## Technology Stack

| Area | Current technology |
| --- | --- |
| Backend | Go 1.25 modular monolith |
| HTTP | Chi |
| Data access | pgx for PostgreSQL |
| Database | PostgreSQL 17 via Docker Compose |
| Password hashing | Argon2id |
| Frontend | React 19, TypeScript, Vite and Tailwind CSS |
| Client foundations | i18next, React Hook Form, Zod and Motion |

## Development Philosophy

We work in small, observable increments. Each increment starts from a user
experience and domain need, then introduces only the design, code, and
infrastructure needed to make that outcome real. The complete working
agreement is in [AGENTS.md](AGENTS.md).

## Roadmap

The roadmap is organized by the experiences Donarium will create, not by
infrastructure milestones. See [ROADMAP.md](ROADMAP.md) for the current order
and the prerequisite follow-ups before Authentication.

## Contributing

Before contributing, read this README, the roadmap, and [AGENTS.md](AGENTS.md).
Propose changes as small vertical slices with a clear actor, visible outcome,
and respect for the project's working philosophy.

## License

License terms have not yet been selected.
