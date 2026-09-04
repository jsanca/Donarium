# Implementation Report: DON-006 — Initial Owner Setup Documentation

## Context

Donarium necesita un primer acceso explícito antes de poder autenticar a una
persona. Sin ese flujo, el sistema no puede determinar de forma segura quién
crea la primera organización ni qué permisos recibe.

## Summary

Se documentó UC-001 como el bootstrap único de una Installation. El caso de
uso crea la Organization inicial, el User inicial, su Membership y Credentials
en una única transacción; su Membership recibe `SUPER_ADMIN` y `OWNER`.

## Deliverables

| Artefacto | Descripción |
| --- | --- |
| [UC-001](../../../knowledge/use-cases/UC-001-initialize-donarium.md) | Caso de uso, flujos, reglas y modelo conceptual. |
| [Business Rules](../../../knowledge/business-rules.md) | BR-01 a BR-09. |
| [ADR-001](../../../knowledge/decisions/ADR-001-installation-bootstrap.md) | Bootstrap previo al login. |
| [ROADMAP.md](../../../ROADMAP.md) | Orden de desarrollo actualizado. |

## Architectural Decisions

ADR-001 establece que la inicialización es anterior al login y solo se permite
una vez por Installation. El primer User recibe sus permisos mediante una
Membership contextual, no mediante roles globales.

## Implementation Notes

Este entregable es documental. La futura implementación debe preservar la
frontera transaccional y nunca registrar ni persistir `RawPassword`.

## Validation

| Comprobación | Resultado |
| --- | --- |
| Enlaces locales Markdown | PASS |
| Reglas BR-01 a BR-09 referenciadas desde UC-001 | PASS |
| Cambios de código | No realizados |
| Build y tests | NOT RUN — no aplican a una tarea documental |

## Tests

No aplican; no se modificó código.

## Tradeoffs

El modelo conceptual muestra la secuencia solicitada para facilitar la
discusión inicial. No prescribe todavía entidades, tablas ni interfaces de
implementación.

## Open Questions

- ¿Qué requisitos concretos compondrán la primera `PasswordPolicy`?
- ¿Qué mecanismo expone la aplicación para detectar una Installation aún no
  inicializada sin revelar información innecesaria?

## Follow-ups

- Implementar el servicio de aplicación de UC-001.
- Implementar la persistencia transaccional del bootstrap.

## References

- [Task](../tasks/DON-006-initial-owner-setup-documentation.md)
- [Checkpoint](../checkpoints/CHECKPOINT-DON-006-initial-owner-setup-documentation.md)
- [ADR-001](../../../knowledge/decisions/ADR-001-installation-bootstrap.md)
