# Implementation Report: DON-007Q.1 — Initialize Donarium Test Case Design

## Context

UC-001 requiere un diseño QA independiente de las pruebas de implementación.
El diseño debe poder ejecutarse sobre un servidor y PostgreSQL dedicados sin
depender de memoria conversacional.

## Summary

Se creó TC-001 para cubrir el flujo principal, los alternativos A1–A4, la
persistencia de cinco registros, el contrato HTTP, la protección de password y
la persistencia después de reiniciar. La concurrencia se documentó como deuda
diferida por ADR-002.

## Deliverables

| Artefacto | Descripción |
| --- | --- |
| [TC-001](../../../knowledge/test-cases/TC-001-initialize-donarium.md) | Caso de prueba funcional, API y persistencia. |
| [Checkpoint](../checkpoints/CHECKPOINT-DON-007Q.1-initialize-test-case-design.md) | Estado de recuperación resuelto. |

## Architectural Decisions

TC-001 respeta ADR-002: no trata la concurrencia como un fallo requerido del
ciclo actual. La ejecución exige una base dedicada para no repetir el riesgo de
pruebas destructivas identificado en DON-007R.

## Implementation Notes

No se modificó código ni se ejecutó ningún test. El rollback se describe como
una composición de prueba aislada con un fallo inducido posterior a una
escritura, no como un detalle de una prueba existente.

## Validation

| Comprobación | Resultado |
| --- | --- |
| Enlaces locales del test case | PASS |
| Cobertura explícita BR-01 a BR-09 | PASS o marcado como inspección/diferido |
| Requests, servidor o PostgreSQL | NOT RUN — fuera de alcance |

## Tests

No ejecutados por instrucción explícita de la tarea.

## Tradeoffs

El contrato actual usa mensajes textuales en `error`, no códigos de error
separados. TC-001 verifica ese contrato sin añadir un campo que la API no
expone.

## Open Questions

- ¿Qué mecanismo de inyección de fallo se adoptará para ejecutar TC-001-09 sin
  afectar la composición de producción?

## Follow-ups

- Aprovisionar entorno PostgreSQL dedicado y ejecutar TC-001-01 a TC-001-10.

## References

- [UC-001](../../../knowledge/use-cases/UC-001-initialize-donarium.md)
- [Business Rules](../../../knowledge/business-rules.md)
- [ADR-001](../../../knowledge/decisions/ADR-001-installation-bootstrap.md)
- [ADR-002](../../../knowledge/decisions/ADR-002-application-level-bootstrap-invariant.md)
- [DON-007R review](../reviews/DON-007R-initial-owner-setup-architecture-review.md)
