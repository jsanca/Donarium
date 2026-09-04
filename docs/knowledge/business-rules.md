# Business Rules

Estas reglas son normativas para los casos de uso de Donarium. Los nombres de
los conceptos se mantienen consistentes con [UC-001](use-cases/UC-001-initialize-donarium.md).

## BR-01 — Inicialización única

Una Installation solo puede completar la inicialización una vez. Un intento
posterior debe rechazarse sin crear ni modificar el estado de bootstrap.

## BR-02 — Email normalizado

Todo email se normaliza antes de validarse, compararse o persistirse. La misma
representación normalizada se usa para identificar al User.

## BR-03 — Email único

Un email normalizado puede identificar como máximo un User dentro de la
Installation.

## BR-04 — Password tratada como dato sensible

La contraseña se recibe como `RawPassword`, se valida y se transforma en
`Hash`. `RawPassword` no se persiste, no se registra en logs, no se incluye en
errores ni se devuelve desde el caso de uso.

## BR-05 — Roles contextuales

Los roles expresan autorización en su contexto correspondiente. `OWNER` es un
`OrganizationRole` de la Membership del User dentro de una Organization.
`SUPER_ADMIN` es un `PlatformRole` de un PlatformGrant separado del User.

## BR-06 — Transacción del caso de uso

La creación de User, Credentials, Organization, Membership y PlatformGrant es
atómica. Si una parte no puede completarse, ninguno de los cinco registros del
estado inicial queda persistido.

## BR-07 — PasswordPolicy

Una `PasswordPolicy` debe validar `RawPassword` antes de generar o persistir
su Hash. La política concreta puede evolucionar, pero nunca se omite para el
primer User.

## BR-08 — Primer usuario = SUPER_ADMIN + OWNER

El User creado por la inicialización recibe una Membership de la Organization
inicial con `OrganizationRole.OWNER`.

El mismo User recibe un PlatformGrant separado con
`PlatformRole.SUPER_ADMIN`.

## BR-09 — Sin usuarios seed en producción

Una instalación de producción no contiene usuarios, credenciales ni accesos
preconfigurados. El único primer acceso se crea mediante UC-001.
