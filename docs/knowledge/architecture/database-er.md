# Database entity relationships

Current PostgreSQL schema as of migrations `001`–`007`
(`server/internal/platform/database/migrations/`). All tables are UUID-keyed
except `memberships` and `platform_grants` (composite).

```mermaid
erDiagram
    users ||--o| credentials : "has (1:1)"
    users ||--o{ organizations : "creates (created_by)"
    users ||--o{ memberships : "user_id"
    organizations ||--o{ memberships : "organization_id"
    users ||--o{ platform_grants : "user_id"
    users ||--o{ properties : "creates (created_by)"
    properties ||--o{ property_stakeholders : "property_id"
    users ||--o{ property_stakeholders : "party_user_id"
    organizations ||--o{ property_stakeholders : "party_org_id"

    users {
        uuid id PK
        text email UNIQUE
        text display_name
        timestamptz created_at
        timestamptz updated_at
    }
    credentials {
        uuid id PK
        uuid user_id UNIQUE FK "users(id) CASCADE"
        text password_hash
    }
    organizations {
        uuid id PK
        text name
        text slug UNIQUE
        uuid created_by FK "users(id) RESTRICT"
    }
    memberships {
        uuid user_id FK "users(id) CASCADE"
        uuid organization_id FK "organizations(id) CASCADE"
        text role "OWNER"
    }
    platform_grants {
        uuid user_id FK "users(id) CASCADE"
        text role "SUPER_ADMIN"
    }
    properties {
        uuid id PK
        text display_name
        text classification "house|apartment|condominium|multi_unit|commercial|other"
        text address_street
        text address_city
        text address_state
        text address_postal_code
        text address_country
        text rental_cadence "monthly|weekly|daily|annual"
        bigint standard_rent
        uuid created_by FK "users(id) RESTRICT"
    }
    property_stakeholders {
        uuid property_id FK "properties(id) CASCADE"
        text party_type "user|organization|external"
        uuid party_user_id FK "users(id) CASCADE"
        uuid party_org_id FK "organizations(id) CASCADE"
        text party_external_name
        text party_external_email
        text role "owner|manager"
        timestamptz created_at
    }
```

## Notes

- **Identity/bootstrap**: `users`, `credentials`, `organizations`,
  `memberships` (`role = OWNER`), `platform_grants` (`role = SUPER_ADMIN`).
- **Property**: `properties` holds identity + address + cadence + rent;
  `created_by` is provenance (`ON DELETE RESTRICT`), not ownership.
- **Property access** is resolved via `property_stakeholders`: a User has access
  if a stakeholder names them directly (`party_type = 'user'`) or names an
  `Organization` on which they hold a `membership` (`party_type = 'organization'`
  joined to `memberships`). `party_type = 'external'` never confers access.
- **Uniqueness** (A-12): `uq_property_stakeholder` on
  `(property_id, party_type, COALESCE(party_user_id), COALESCE(party_org_id),
  COALESCE(lower(party_external_email)), role)` — the same Party may hold both
  `owner` and `manager` on a property.

> This page documents the schema as evidence of current behavior. Schema history
> and rationale: ADR-004/ADR-005 under `../adr/`; migration drafts reviewed at
> EP-001.10.
