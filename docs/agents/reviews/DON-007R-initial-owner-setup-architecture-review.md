# DON-007R — Initial Owner Setup Architecture Review

## Review Scope

Se revisaron el estado actual del repositorio y los materiales obligatorios:

- `README.md`, `ROADMAP.md` y `AGENTS.md`.
- [UC-001](../../../knowledge/use-cases/UC-001-initialize-donarium.md),
  [Business Rules](../../../knowledge/business-rules.md) y
  [ADR-001](../../../knowledge/decisions/ADR-001-installation-bootstrap.md).
- Los task records, reportes y checkpoints DON-007.1 a DON-007.5-F1.
- El dominio Identity, aplicación, adaptadores pgx, migraciones, HTTP runtime,
  composition root y las pruebas unitarias, HTTP, de repositorio e integración.

Validación ejecutada en esta revisión:

```text
cd server
go test -count=1 ./internal/identity ./internal/identity/application ./internal/identity/http  PASS
go vet ./...                                                                  PASS
```

Las pruebas PostgreSQL no se ejecutaron: sus `TestMain` eliminan tablas de la
base configurada. Se revisó su código en lugar de arriesgar datos locales fuera
del alcance de esta revisión.

## Traceability Summary

La ruta implementada es `POST /api/setup` → `SetupHandler` →
`TransactionalSetupService` → `CanonicalSetupService` → repositorios pgx →
PostgreSQL. La capa HTTP mantiene la validación de transporte y el mapeo de
errores; la aplicación crea User, Credential, Organization, Membership y
PlatformGrant dentro del executor recibido.

| Regla | Implementación | Prueba | Error / HTTP |
| --- | --- | --- | --- |
| BR-01 inicialización única | `OrganizationRepo.ExistsAny` | Secuencial: sí; concurrente: no | `ErrAlreadyInitialized` → 409 |
| BR-02 email normalizado | `DefaultEmailNormalizer` antes de consultas | Unitario parcial | `ErrInvalidEmail` → 400 |
| BR-03 email único | Consulta previa + `uq_users_email` | Repositorio y HTTP | `ErrDuplicateEmail` → 409 |
| BR-04 password sensible | Argon2id y `PasswordHash`; no se retorna | Implementación inspeccionada | Errores sin hash → 400/500 |
| BR-05 / BR-08 roles | Owner en Membership; Super Admin en PlatformGrant | Integración inspeccionada | Sin error específico |
| BR-06 transacción | Decorador + pgx transaction | Éxito sí; fallo real post-escritura no | Error original → 500 cuando corresponde |
| BR-07 PasswordPolicy | `DefaultPasswordPolicy().Validate` | Unitario y HTTP | `ErrInvalidPassword` → 400 |
| BR-09 sin seed | No hay inserciones seed en migraciones | Inspección de migraciones | No aplica |

## Verdict

**CHANGES REQUIRED.** No se recomienda iniciar DON-008 (Authentication) hasta
resolver R-001 y R-002. R-003 debe resolverse junto con esos cambios para que
la validación vertical no pueda pasar sin PostgreSQL real. No hay evidencia de
un riesgo de contraseña en texto plano, de inyección SQL, ni de lógica de
negocio en el handler.

## Findings

### R-001 — MAJOR — BR-01 no está protegido contra dos inicializaciones concurrentes

- **Category:** Persistencia y transacciones
- **Evidence:** `CanonicalSetupService` decide la inicialización mediante
  `OrganizationRepository.ExistsAny` antes de las inserciones
  ([setup.go](../../../server/internal/identity/application/setup.go#L103-L120)).
  `TransactionManager` abre la transacción predeterminada de pgx, sin una
  invariante singleton ni aislamiento explícito
  ([transaction.go](../../../server/internal/identity/pgx/transaction.go#L21-L38)).
  La migración de organizations solo impone unicidad de `slug`, no una sola
  instalación u organización ([003_organizations.up.sql](../../../server/internal/platform/database/migrations/003_organizations.up.sql#L1-L8)).
- **Impact:** Dos solicitudes iniciales concurrentes con emails y slugs
  distintos pueden observar una tabla vacía y confirmar ambas. Esto viola
  BR-01 y la postcondición de una única Organization inicial; no sobrescribe el
  primer setup, pero permite un segundo setup válido.
- **Recommendation:** Modelar y persistir el estado singleton de Installation,
  o una invariante de base de datos equivalente, y convertir la violación a
  `ErrAlreadyInitialized`/409. Añadir una prueba vertical concurrente con dos
  solicitudes iniciales distintas. Es una corrección de invariante, no una
  recomendación de locking especulativo.
- **Affected files:** `server/internal/identity/application/setup.go`,
  `server/internal/identity/pgx/transaction.go`,
  `server/internal/platform/database/migrations/003_organizations.up.sql`,
  `server/tests/integration/setup_vertical_test.go`.
- **Blocker / fix-defer:** Resolver antes de Authentication.

### R-002 — MAJOR — UC-001 y ADR-001 contradicen el modelo de roles implementado

- **Category:** Trazabilidad funcional y modelo de identidad
- **Evidence:** UC-001 asigna `SUPER_ADMIN` y `OWNER` a la Membership
  ([UC-001](../../../knowledge/use-cases/UC-001-initialize-donarium.md#L33-L39),
  [postcondiciones](../../../knowledge/use-cases/UC-001-initialize-donarium.md#L63-L71)).
  BR-05 y BR-08 hacen la misma afirmación
  ([Business Rules](../../../knowledge/business-rules.md#L27-L48)), al igual
  que ADR-001 ([ADR-001](../../../knowledge/decisions/ADR-001-installation-bootstrap.md#L18-L22)).
  El código separa correctamente el `owner` de OrganizationRole en Membership
  ([membership.go](../../../server/internal/identity/membership.go#L7-L30),
  [role.go](../../../server/internal/identity/role.go#L13-L21)) y el
  `super_admin` de PlatformRole en PlatformGrant
  ([platform_grant.go](../../../server/internal/identity/platform_grant.go#L5-L23),
  [role.go](../../../server/internal/identity/role.go#L3-L11)); el servicio
  crea ambos registros separados
  ([setup.go](../../../server/internal/identity/application/setup.go#L139-L167)).
- **Impact:** La documentación canónica describe una autorización diferente de
  la que consume el sistema. Authentication podría implementar la lectura de
  permisos en Membership y omitir PlatformGrant.
- **Recommendation:** Corregir UC-001, BR-05, BR-06, BR-08 y ADR-001: Owner es
  el rol de Organization en Membership; Super Admin es un PlatformGrant. Añadir
  PlatformGrant al modelo conceptual y a las postcondiciones atómicas.
- **Affected files:** `knowledge/use-cases/UC-001-initialize-donarium.md`,
  `knowledge/business-rules.md`,
  `knowledge/decisions/ADR-001-installation-bootstrap.md`.
- **Blocker / fix-defer:** Resolver ahora, antes de que DON-008 defina
  autorización.

### R-003 — MAJOR — Las pruebas PostgreSQL pueden informar éxito sin ejecutar ninguna prueba

- **Category:** Confiabilidad de pruebas
- **Evidence:** Ambos `TestMain` finalizan con `os.Exit(0)` cuando no se puede
  crear o pinguear PostgreSQL
  ([repository_test.go](../../../server/internal/identity/pgx/repository_test.go#L20-L39),
  [setup_vertical_test.go](../../../server/tests/integration/setup_vertical_test.go#L25-L45)).
  Además, usan por defecto una base `donarium` local y eliminan sus tablas
  ([setup_vertical_test.go](../../../server/tests/integration/setup_vertical_test.go#L47-L65)).
- **Impact:** Un pipeline puede pasar con las pruebas de repositorio y
  verticales omitidas; si PostgreSQL está disponible, las pruebas destruyen el
  esquema de una base de desarrollo con el mismo nombre.
- **Recommendation:** Exigir un `TEST_DATABASE_URL` dedicado y fallar cuando
  no esté disponible; aislar la base de pruebas para que las operaciones
  destructivas no apunten al entorno de desarrollo.
- **Affected files:** `server/internal/identity/pgx/repository_test.go`,
  `server/tests/integration/setup_vertical_test.go`, configuración de CI.
- **Blocker / fix-defer:** Resolver con R-001 antes de depender de validación
  vertical para Authentication.

### R-004 — MINOR — La ruta de rollback real posterior a una escritura no está cubierta

- **Category:** Calidad de pruebas
- **Evidence:** Las pruebas de aplicación usan un `fakeTxManager` que solo
  invoca el callback, sin commit ni rollback
  ([setup_test.go](../../../server/internal/identity/application/setup_test.go#L265-L340)).
  La prueba vertical de consistencia provoca el segundo setup después de que
  `ExistsAny` ya devuelve verdadero, por lo que falla antes de cualquier INSERT
  ([setup_vertical_test.go](../../../server/tests/integration/setup_vertical_test.go#L129-L170),
  [setup.go](../../../server/internal/identity/application/setup.go#L103-L120)).
- **Impact:** El código de `defer tx.Rollback(ctx)` es correcto por inspección,
  pero una regresión que use el pool en una de las cinco escrituras podría
  dejar de estar cubierta por una prueba real.
- **Recommendation:** Incorporar una prueba PostgreSQL que fuerce un error
  determinista después de al menos una escritura y compruebe cero filas en las
  cinco tablas al finalizar.
- **Affected files:** `server/internal/identity/application/setup_test.go`,
  `server/tests/integration/setup_vertical_test.go`.
- **Blocker / fix-defer:** Corregir junto con R-003; no exige rediseñar el
  TransactionManager.

### R-005 — MINOR — Las respuestas 405 rompen el contrato JSON de la API

- **Category:** API REST
- **Evidence:** Los paths de método incorrecto usan `http.Error`, que produce
  texto plano ([handler.go](../../../server/internal/identity/http/handler.go#L33-L38),
  [handler.go](../../../server/internal/identity/http/handler.go#L92-L97));
  400, 409 y 500 usan `ErrorResponse` JSON
  ([handler.go](../../../server/internal/identity/http/handler.go#L110-L136)).
- **Impact:** Un consumidor que espere un error JSON consistente debe tratar
  405 como una excepción.
- **Recommendation:** En una corrección pequeña, usar `writeError` para 405 y
  conservar el header `Allow`; añadir aserciones de body y `Content-Type`.
- **Affected files:** `server/internal/identity/http/handler.go`,
  `server/internal/identity/http/handler_test.go`.
- **Blocker / fix-defer:** Puede resolverse después de R-001 a R-004, antes de
  exponer la API a un cliente externo.

### R-006 — MINOR — README no representa el estado actual del repositorio

- **Category:** Documentación de estado actual
- **Evidence:** README declara Foundation planificada, sin código y sin stack
  decidido ([README.md](../../../README.md#L38-L67)), mientras AGENTS y el
  árbol contienen un monolito Go, cliente React, PostgreSQL y UC-001.
- **Impact:** Un nuevo contribuidor recibe instrucciones incompatibles con el
  repositorio que debe revisar o ejecutar.
- **Recommendation:** Actualizar README como una tarea documental breve para
  reflejar el stack, layout y estado real; no cambiar la visión de producto.
- **Affected files:** `README.md`.
- **Blocker / fix-defer:** Puede esperar hasta que R-001 a R-004 estén
  corregidos, pero debe resolverse antes de ampliar la contribución externa.

## Positive Findings

- El handler es delgado: valida transporte, delega a `SetupPerformer` y mapea
  errores; no contiene reglas de setup.
- Application y el dominio no importan pgx. `DBExecutor` se pasa de forma
  explícita a los repositorios y la adaptación de pgx queda en infraestructura.
- Las cinco escrituras están dentro del callback de `TransactionalSetupService`.
- Las migraciones imponen claves foráneas, email/slug únicos y roles sin valores
  especulativos: `owner` para Membership y `super_admin` para PlatformGrant.
- Credential almacena `PasswordHash`; Argon2id genera un hash con salt y el
  API solo retorna IDs. Las consultas usan parámetros `$1` a `$5`.
- Las respuestas 400, 409 y 500 siguen un DTO JSON estable; los errores
  inesperados no se exponen al cliente.

## Deferred Findings

### FUTURE CONSIDERATION — PasswordPolicy y representación de RawPassword

El comando de aplicación usa `Password string` y la política cuenta bytes
(`len`) aunque clasifica caracteres Unicode. No hay evidencia de pérdida de
seguridad actual: el valor se hashea y no se persiste. Al introducir requisitos
de password internacionales o autenticación, conviene decidir si el mínimo es
por bytes o runas y si un tipo explícito `RawPassword` mejora la frontera.

### OBSERVATION — DBExecutor es una decisión consciente, no una dependencia pgx

La interfaz se ubica en Identity y expone operaciones de executor porque los
repositorios reciben la transacción explícitamente. No importa pgx y el adaptador
resuelve la incompatibilidad de firmas. No se clasifica como defecto en este
slice; debe revisarse solo si más casos de uso fuerzan un contrato de base de
datos más amplio en el dominio.

## Conclusion

UC-001 está mayormente implementado a través de HTTP, aplicación, repositorios
y PostgreSQL; las rutas, el modelo de identidad y la protección de password son
coherentes en código. Sin embargo, no está completamente listo como base para
Authentication: BR-01 admite una carrera razonable, los documentos canónicos
contradicen la separación Membership/PlatformGrant, y la suite vertical puede
omitirse silenciosamente. No existe evidencia de sobrescritura de un setup
existente; existe evidencia de que dos setups distintos podrían coexistir bajo
concurrencia.

La deuda que debe resolverse ahora es R-001 a R-004. R-005 y R-006 pueden
planificarse como follow-ups cortos. Tras resolver R-001 a R-004 y repetir la
validación con PostgreSQL dedicado, se puede iniciar DON-008.

## References

- [UC-001](../../../knowledge/use-cases/UC-001-initialize-donarium.md)
- [Business Rules](../../../knowledge/business-rules.md)
- [ADR-001](../../../knowledge/decisions/ADR-001-installation-bootstrap.md)
- [DON-007 task](../tasks/DON-007%E2%80%94InitialOwnerSetup%20%28VerticalSlice%29.md)
- [DON-007.5 validation report](../reports/DON-007.5-vertical-integration-validation.md)
- [DON-007.5-F1 cleanup report](../reports/DON-007.5-F1-transaction-cleanup.md)
- [Review checkpoint](../checkpoints/CHECKPOINT-DON-007R-initial-owner-setup-review.md)

## Final Verdict

CHANGES REQUIRED
