# ADR-005 — Party model (UserRef | OrganizationRef | ExternalParty)

- **Estado:** Aceptado
- **Fecha:** 2026-09-02
- **Relacionado con:** [Plan EP-001 — First Property Experience](../engineering/plans/DONARIUM-PA-EXP-001-first-property-experience-engineering-plan-proposal.md) (EP-001.03), ADR-004, `server/internal/properties/party.go`
- **Decide:** `software-engineer` (EP-001.03)
- **Revisa:** `docs/engineering/ENGINEERING_LOG.md` (EP-001.03)

## Contexto

EP-001.03 debe representar “quién” es responsable de una propiedad. Stitch muestra cuatro opciones de relación que combinan identidad personal y organizativa con terceros externos. El modelo debe permitir:

- identificar a un `User` concreto de Donarium,
- identificar a una `Organization` concreta de Donarium,
- referenciar a una persona externa solo por nombre + email de contacto (sin cuenta),
sin crear un “Organization por defecto” sintética ni extender `User` con campos de propiedad.

Alternativas consideradas:

1. **Tres columnas nullable en `properties`** (`owner_user_id`, `manager_user_id`, `owner_org_id`…) — rígido, no escala a múltiples stakeholders por propiedad ni a que el mismo Party tenga ambos roles.
2. **Party polimórfico como valor en `PropertyStakeholder`** — `PartyType` discriminador con variantes `UserRef(UserID)`, `OrganizationRef(OrganizationID)`, `ExternalParty(Name, Email)`. Cada stakeholder referencia exactamente una variante; la persistencia usa columnas discriminadas con checks por tipo.
3. **Tabla única `parties` con FK genérico** — añadiría una entidad intermedia sin comportamiento propio y requeriría gestión de ciclo de vida de `ExternalParty` (que no tiene cuenta).

## Decisión

Se adopta **(2) Party polimórfico**.

- Tipo `Party` en `server/internal/properties/party.go` con `Type PartyType` (`user`, `organization`, `external`) y campos discriminados: `UserID *uuid.UUID`, `OrganizationID *uuid.UUID`, `ExternalName/ExternalEmail string`. `Validate()` exige exactamente la variante correspondiente y, para `External`, nombre 2–100 y email con `mail.ParseAddress`; `NormalizedExternalEmail()` baja a minúsculas.
- `PropertyStakeholder` compone `Party` + `Role` (`owner` | `manager`). La unicidad es `(PropertyID, Party, Role)` per A-12; el mismo Party puede ser ambos roles en la misma propiedad.
- Persistencia discriminada en `property_stakeholders` con `party_type`, `party_user_id`, `party_org_id`, `party_external_name/email` y checks `chk_party_user/org/external`. Conversión explícita en `pgx/stakeholder_repository.go`.
- Semántica de acceso: `UserRef` confiere acceso directo; `OrganizationRef` confiere acceso si el `User` tiene `Membership` en esa `Organization`; `ExternalParty` **nunca** confiere acceso (contacto solo). Resuelto en `FindAccessibleByUser` / `HasAccess` vía join `memberships`.
- Validación en `RegisterProperty` para cada stakeholder: `Party.Validate()`, `Role` parseado, existencia de `User`/`Organization` referenciados (SELECT por id), y post-condición “al menos un stakeholder vincula al actor” (directo o vía membership). `ExternalParty` no vincula.
- H-2: no se cita UC-001/BR-05/06/08/ADR-001 contradichos; el lenguaje se apoya en `Membership`/`Organization` reales.

## Consecuencias

### Ventajas

- Expresa exactamente las cuatro opciones Stitch sin columnas ad-hoc: “I am both Owner & Manager” → dos filas `UserRef(actor)`; “I am Manager on behalf of Owner” → `UserRef(actor)` MANAGER + `ExternalParty` OWNER; “I am Owner delegating to another manager” → inverso; “Acting on behalf of an Organization” → `OrganizationRef(org)` OWNER/MANAGER +/− `UserRef(actor)`.
- `ExternalParty` queda claramente sin acceso, evitando privilegios implícitos.
- Un solo lugar (`party.go` + `stakeholder.go`) para evolucionar futuros `Party` sin tocar `identity`.

### Costos / riesgos

- Lógica de validación más rica en `RegisterProperty` (verificación de existencia y de vínculo al actor requiere consultas; external no requiere). Mitigado con índices y transacción única.
- El cliente debe enviar `party.type` explícito; la API rechaza combinaciones incoherentes con `party is not valid` (400).

### No objetivos

- No se modela aceptación de invitación como `Organization` (personal-only en v1 per A-8).
- No se introduce “Unit” ni se expande `OrganizationRole`.

## Alternativas descartadas

Columnas nullable en `properties` y tabla intermedia `parties` se descartan por rigidez y por introducir entidad sin comportamiento.

## Evolución

Si una futura slice permite que una invitación sea aceptada en nombre de una `Organization`, el `Party` ya lo soporta; solo requerirá ampliar la regla de aceptación (EP-001.13).

