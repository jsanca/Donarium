# UC-001 — Initialize Donarium

## Objetivo

Configurar una instalación nueva de Donarium creando la primera organización y
su propietario administrador. Este es el único flujo que habilita el acceso
inicial; no es un flujo de login ni crea datos de demostración.

## Actores

| Actor | Responsabilidad |
| --- | --- |
| Primer Owner | Proporciona los datos de su organización y sus credenciales iniciales. |
| Donarium | Valida los datos, aplica la política de contraseña y crea el estado inicial de forma atómica. |

## Precondiciones

- La instalación está disponible y puede persistir el estado inicial.
- La instalación no ha sido inicializada previamente.
- El primer Owner proporciona su nombre visible, el nombre y slug de la
  organización, un email y una contraseña que cumplen los requisitos de entrada.
- La contraseña llega al caso de uso como `RawPassword` y se mantiene solo en
  memoria durante el procesamiento.

## Flujo principal

1. El Primer Owner inicia **Initialize Donarium**.
2. Donarium abre la frontera transaccional del caso de uso.
3. Donarium valida los datos requeridos y verifica que la instalación aún no
   está inicializada.
4. Donarium normaliza el email proporcionado, valida su unicidad y aplica la
   `PasswordPolicy`.
5. Donarium transforma `RawPassword` en `PasswordHash` sin persistir ni registrar la
   contraseña original.
6. Donarium crea el `User` inicial y sus `Credentials` con el hash.
7. Donarium crea la `Organization` inicial, identificada por su slug.
8. Donarium crea la `Membership` que vincula al User con la Organization y le
   asigna `OrganizationRole.OWNER`.
9. Donarium crea el `PlatformGrant` del mismo User y le asigna
   `PlatformRole.SUPER_ADMIN`.
10. Donarium confirma los cinco registros como una única transacción y marca la
    instalación como inicializada.

## Flujos alternativos

### A1 — Instalación ya inicializada

Si la instalación ya tiene su estado inicial, Donarium rechaza el flujo sin
crear ni alterar organizaciones, usuarios, membresías o credenciales.

### A2 — Email inválido o ya existente

Si el email no puede normalizarse, no tiene formato válido o ya identifica a
un User, Donarium informa el error de validación y no persiste cambios.

### A3 — Contraseña no conforme

Si `RawPassword` no cumple la `PasswordPolicy`, Donarium informa el requisito
incumplido, descarta el valor recibido y no persiste cambios.

### A4 — Fallo durante la creación

Si cualquier creación o persistencia falla, Donarium revierte la transacción.
La instalación conserva el estado que tenía antes de iniciar el flujo.

## Postcondiciones

- Existe exactamente una Organization inicial para la instalación.
- Existe un User inicial identificado por su email normalizado.
- Existe una Membership entre ese User y la Organization inicial con
  `OrganizationRole.OWNER`.
- Existe un PlatformGrant para el mismo User con `PlatformRole.SUPER_ADMIN`.
- Las Credentials del User contienen un Hash; la contraseña original no queda
  persistida ni registrada.
- La instalación queda marcada como inicializada y el flujo no puede repetirse.

## Reglas de negocio

- [BR-01 — Inicialización única](../business-rules.md#br-01--inicialización-única)
- [BR-02 — Email normalizado](../business-rules.md#br-02--email-normalizado)
- [BR-03 — Email único](../business-rules.md#br-03--email-único)
- [BR-04 — Password tratada como dato sensible](../business-rules.md#br-04--password-tratada-como-dato-sensible)
- [BR-05 — Roles contextuales](../business-rules.md#br-05--roles-contextuales)
- [BR-06 — Transacción del caso de uso](../business-rules.md#br-06--transacción-del-caso-de-uso)
- [BR-07 — PasswordPolicy](../business-rules.md#br-07--passwordpolicy)
- [BR-08 — Primer usuario = SUPER_ADMIN + OWNER](../business-rules.md#br-08--primer-usuario--super_admin--owner)
- [BR-09 — Sin usuarios seed en producción](../business-rules.md#br-09--sin-usuarios-seed-en-producción)

## Modelo conceptual

```text
Installation
     │ inicia el bootstrap de
     ▼
Organization
     ▲                    ▲
     │ OrganizationRole   │ Membership vincula
     │ OWNER              │
Membership ──────────── User ───────────── Credentials
                            │                  │
                            │                  └── PasswordHash
                            ▼
                     PlatformGrant
                            │
                            └── PlatformRole SUPER_ADMIN
```

La Membership es la relación de un User con una Organization y contiene solo
roles de organización. PlatformGrant pertenece al User y contiene roles de
plataforma. Credentials pertenece al User y solo conserva su `PasswordHash`.

## Ejemplo de resultado inicial

Para una organización `northstar` y el User `owner@northstar.example`, la
inicialización crea una Membership con `OrganizationRole.OWNER` y un
PlatformGrant separado con `PlatformRole.SUPER_ADMIN`.
