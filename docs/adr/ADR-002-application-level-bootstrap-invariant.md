# ADR-002 — Application-Level Bootstrap Invariant

- **Estado:** Aceptado
- **Fecha:** 2026-07-20
- **Relacionado con:** [ADR-001 — Installation Bootstrap](ADR-001-installation-bootstrap.md)

## Contexto

El setup de Donarium se ejecuta una única vez durante una instalación. En el
modelo actual, la aplicación comprueba la ausencia de organizaciones mediante
`ExistsAny()` y devuelve `ErrAlreadyInitialized` cuando ya existe una.

Esta comprobación cubre la inicialización secuencial, pero no existe una
restricción singleton persistente. Dos solicitudes perfectamente concurrentes,
con emails y slugs diferentes, pueden observar el estado vacío antes de que
alguna confirme sus cambios.

## Decisión

Donarium acepta esta pequeña ventana de carrera durante la inicialización para
mantener la simplicidad del sistema en el modelo de despliegue actual.

No se introducen en este momento:

- advisory locks;
- optimistic locking;
- aislamiento `SERIALIZABLE`; ni
- tablas singleton.

La invariante de bootstrap se mantiene en el nivel de aplicación mediante
`ExistsAny()` seguido de `ErrAlreadyInitialized` para intentos posteriores que
observan una instalación ya configurada.

## Consecuencias

### Ventajas

- Arquitectura sencilla.
- Menor complejidad operativa y de mantenimiento.
- Comportamiento suficiente para el modelo de despliegue actual, en el que la
  inicialización es una acción administrativa poco frecuente.

### Riesgos conocidos

- Dos solicitudes exactamente concurrentes podrían inicializar dos estados de
  setup si usan datos distintos.
- La garantía de inicialización única es de aplicación para la ejecución
  habitual, no una invariante persistente ante concurrencia perfecta.

## Evolución futura

Esta decisión se reevaluará si Donarium evoluciona hacia:

- SaaS;
- múltiples nodos;
- inicializaciones concurrentes; o
- despliegues distribuidos.

En ese contexto podrán introducirse mecanismos de persistencia que garanticen
la unicidad de la instalación.
