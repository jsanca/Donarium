# Engineering Log

This file is the compact, current index of material engineering work. Detailed task, report, review, and checkpoint records live under `agents/`; this index links their relationship rather than repeating their evidence.

| Task | Description | Status | Depends On | Task File | Report | Review | Fix / Checkpoint | Knowledge |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| TASK-ID | concise outcome | Complete | — | [task](agents/tasks/TASK-ID.md) | [report](agents/reports/TASK-ID.md) | — | — | [current concept](../knowledge/area/concept.md) |
| EP-001 | First Property Experience — approved plan (r3) | Plan approved; nodes materialized, awaiting Role adoption | DON-007R (soft) | [plan](plans/DONARIUM-PA-EXP-001-first-property-experience-engineering-plan-proposal.md) | — | — | — | — |
| EP-001.01 | Post-login shell + router + empty-state routing | Materialized | EP-001.06 | [task](agents/tasks/EP-001.01-post-login-shell-and-router.md) | — | — | — | — |
| EP-001.02 | Property domain + `POST/GET /api/properties` | Materialized | — | [task](agents/tasks/EP-001.02-property-domain-and-endpoints.md) | — | — | — | — |
| EP-001.03 | Party model + `PropertyStakeholder` + wizard step 2 | Materialized | EP-001.02 | [task](agents/tasks/EP-001.03-party-model-and-property-stakeholder.md) | — | — | — | — |
| EP-001.04 | Portfolio + Property Home read model | Materialized | EP-001.01, .02, .03, .12 | [task](agents/tasks/EP-001.04-portfolio-and-property-home-read-model.md) | — | — | — | — |
| EP-001.05 | Wizard step 3 — initial tenancy (property-side) | Materialized | EP-001.02 | [task](agents/tasks/EP-001.05-wizard-step-3-initial-tenancy.md) | — | — | — | — |
| EP-001.06 | Bilingual i18n materialization (ES/EN, frontend) | Materialized | — | [task](agents/tasks/EP-001.06-bilingual-i18n-materialization.md) | — | — | — | — |
| EP-001.07 | Deep-link `/properties/:id` authorization (404 for non-authorized) | Materialized | EP-001.02 | [task](agents/tasks/EP-001.07-deep-link-authorization-non-invitation.md) | — | — | — | — |
| EP-001.08 | Independent boundary review | Materialized | EP-001.03, .12, .13 | [task](agents/tasks/EP-001.08-boundary-review.md) | — | — | — | — |
| EP-001.09 | Verification of outcome for the experience | Materialized | EP-001.01, .02, .03, .04, .05, .06, .07, .12, .13, .14 | [task](agents/tasks/EP-001.09-verification-of-outcome.md) | — | — | — | — |
| EP-001.10 | Migration + operational review | Materialized | EP-001.02, .03, .12, .13 (draft SQL) | [task](agents/tasks/EP-001.10-migration-and-operational-review.md) | — | — | — | — |
| EP-001.11 | Knowledge curation | Materialized | EP-001.03, .05, .12, .13 | [task](agents/tasks/EP-001.11-knowledge-curation.md) | — | — | — | — |
| EP-001.12 | Register Offline Payment (narrow log) | Materialized | EP-001.02 | [task](agents/tasks/EP-001.12-register-offline-payment-narrow-log.md) | — | — | — | — |
| EP-001.13 | Invitation domain + endpoints (URL-only; delivery decoupled) | Materialized | EP-001.02, .03 | [task](agents/tasks/EP-001.13-invitation-domain-and-endpoints.md) | — | — | — | — |
| EP-001.14 | Access-Existing-Property client experience (URL-triggered) | Materialized | EP-001.01, .13 | [task](agents/tasks/EP-001.14-access-existing-property-client-experience.md) | — | — | — | — |

Use `—` where a relationship does not exist. Keep cells brief; the linked durable record carries evidence, limitations, unresolved issues, and validation. Add a knowledge link when work establishes or changes reusable current understanding.
