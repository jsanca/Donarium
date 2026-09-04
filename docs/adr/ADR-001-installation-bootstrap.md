# ADR-001 — Installation Bootstrap

- **Estado:** Aceptado
- **Fecha:** 2026-07-19

## Contexto

Una instalación nueva no tiene Organization, User ni Credentials. Un flujo de
login presupone que ya existe una identidad, por lo que no puede crear el
primer acceso de forma segura ni expresar quién obtiene el control inicial de
la instalación.

## Decisión

Donarium tendrá un flujo explícito de inicialización anterior al login:
[UC-001 — Initialize Donarium](../use-cases/UC-001-initialize-donarium.md).

El flujo se permite una sola vez por Installation y, en una sola transacción,
crea la Organization inicial, el primer User, sus Credentials, su Membership y
su PlatformGrant. La Membership recibe `OrganizationRole.OWNER`; el
PlatformGrant separado recibe `PlatformRole.SUPER_ADMIN`. Una vez completado,
los intentos posteriores se rechazan y el login solo autentica identidades ya
creadas. El comportamiento de la invariante frente a solicitudes concurrentes
se define en [ADR-002 — Application-Level Bootstrap Invariant](ADR-002-application-level-bootstrap-invariant.md).

## Consecuencias

- El primer acceso es explícito, auditable y no depende de datos seed.
- La aplicación puede decidir de forma inequívoca si mostrar inicialización o
  login sin mezclar responsabilidades.
- La política de contraseña y el tratamiento seguro de credenciales se aplican
  desde el primer User.
- La autorización de organización y la autorización de plataforma permanecen
  separadas desde el bootstrap: `OWNER` no implica `SUPER_ADMIN`, ni viceversa.
- El estado de bootstrap exige una frontera transaccional clara en la futura
  implementación.

## Alternativas descartadas

### Usuario administrador seed

Descartado porque introduce credenciales predecibles o distribución insegura de
secretos en producción.

### Crear el primer usuario desde el login

Descartado porque mezcla autenticación con bootstrap y no puede garantizar de
forma explícita la creación única ni el rol contextual inicial.
