# ADR-004 — PropertyStakeholder as per-property relation

- **Estado:** Aceptado
- **Fecha:** 2026-09-02
- **Relacionado con:** [Plan EP-001 — First Property Experience](../engineering/plans/DONARIUM-PA-EXP-001-first-property-experience-engineering-plan-proposal.md) (EP-001.03), ADR-005, `server/internal/properties/stakeholder.go`, `007_property_stakeholders.{up,down}.sql`
- **Decide:** `software-engineer` (EP-001.03)
- **Revisa:** `docs/engineering/ENGINEERING_LOG.md` (EP-001.03)

## Contexto

Donarium ya modela `Membership` (per-organization, `OrganizationRole`) y `PlatformGrant` (platform-wide). EP-001.03 necesita representar relaciones Owner/Manager **por propiedad**, donde el mismo Party puede ser OWNER en una propiedad y MANAGER en otra, e incluso ambos roles en la misma propiedad (A-12).

Alternativas consideradas:

1. **Extender `OrganizationRole` / `Membership`** para cubrir propiedades (p. ej., añadir `PROPERTY_OWNER`, `PROPERTY_MANAGER` a `OrganizationRole` o una columna `property_id` en `memberships`). Esto reutilizaría la tabla `memberships` pero contaminaría el límite organizativo: `memberships` dejaría de significar “pertenencia a una organización” y pasaría a ser un contenedor polimórfico (org vs property), con checks y unicidades entrecruzadas.
2. **Nueva relación `PropertyStakeholder`** — tabla dedicada `property_stakeholders` con `(property_id, party_type, party_reference, role)` y unicidad `(property_id, party, role)` per A-12, sin tocar `identity` schema. La verificación de acceso se resuelve por join `property_stakeholders` ↔ `memberships` para el caso `OrganizationRef`.

## Decisión

Se adopta **(2) `PropertyStakeholder` como relación dedicada por propiedad**.

- Entidad `PropertyStakeholder(PropertyID, Party, Role, CreatedAt)` en `server/internal/properties/stakeholder.go`.
- Persistencia en `property_stakeholders` (`007_property_stakeholders.up.sql`) con `party_type IN ('user','organization','external')`, `role IN ('owner','manager')`, checks que garantizan columnas coherentes por tipo, índice único `uq_property_stakeholder` sobre `(property_id, party_type, COALESCE(party_user_id), COALESCE(party_org_id), COALESCE(lower(party_external_email)), role)` per A-12, e índices por `property_id`, `party_user_id` y `party_org_id`.
- `Membership` y `OrganizationRole` **no se extienden**; el esquema `identity` permanece intacto (requisito explícito de EP-001.03).
- `RegisterProperty` persiste la `Property` y todos los `PropertyStakeholder` declarados **atómicamente** en la misma transacción (`TransactionManager.WithinTransaction`); fallo en cualquier stakeholder revierte la propiedad.
- Acceso a propiedad se finaliza aquí: `FindAccessibleByUser` y `HasAccess` resuelven `UserRef` directo y `OrganizationRef` vía `memberships`. `ExternalParty` nunca confiere acceso.
- H-2: esta decisión no cita las secciones contradichas de UC-001, BR-05/06/08 ni ADR-001 como autoridad; la semántica se sostiene en la forma actual del código (`Membership` per-organización, `PlatformGrant` platform-wide).

## Consecuencias

### Ventajas

- Límite de dominio preservado: `identity` no se convierte en contenedor de propiedades.
- Unicidad por propiedad clara y testeable; mismo Party puede ostentar ambos roles sin duplicados por rol.
- Migración aditiva y reversible; sin impacto en datos existentes de `memberships`/`platform_grants`.
- Evolución futura (p. ej., permisos más finos que OWNER/MANAGER) no exige alterar `OrganizationRole`.

### Costos / riesgos

- Una tabla y repositorio más (`StakeholderRepository`); el servicio `RegisterProperty` debe validar existencia de `User`/`Organization` referenciados y que al menos un stakeholder vincule al actor (directo o vía membership) — añade consultas por stakeholder.
- El join de acceso introduce una lectura adicional (`memberships`) en `FindAccessibleByUser`/`HasAccess`, mitigado con índices.

### No objetivos

- No se implementa edición de stakeholders tras el registro; no se implementa pago ni invitación (EP-001.12/.13).
- No se extiende el modelo de invitaciones; `ExternalParty` sigue siendo solo contacto.

## Alternativas descartadas

Extender `OrganizationRole`/`Membership` se descarta por contaminación de límites y pérdida de claridad de `memberships` como relación per-organización.

## Evolución

Reevaluar si Donarium introduce permisos por propiedad más granulares (p. ej., `VIEWER`, `ACCOUNTANT`). La tabla y el enum están preparados para registrar nuevos roles sin reescribir `Membership`.

