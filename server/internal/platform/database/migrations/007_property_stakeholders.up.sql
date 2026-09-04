CREATE TABLE IF NOT EXISTS property_stakeholders (
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    party_type TEXT NOT NULL CHECK (party_type IN ('user','organization','external')),
    party_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    party_org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    party_external_name TEXT,
    party_external_email TEXT,
    role TEXT NOT NULL CHECK (role IN ('owner','manager')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_party_user CHECK (
        (party_type = 'user' AND party_user_id IS NOT NULL AND party_org_id IS NULL AND party_external_name IS NULL AND party_external_email IS NULL)
        OR (party_type = 'organization' AND party_org_id IS NOT NULL AND party_user_id IS NULL AND party_external_name IS NULL AND party_external_email IS NULL)
        OR (party_type = 'external' AND party_external_name IS NOT NULL AND party_external_email IS NOT NULL AND party_user_id IS NULL AND party_org_id IS NULL)
    ),
    CONSTRAINT chk_external_email CHECK (
        party_type != 'external' OR party_external_email ~ '^[^@]+@[^@]+\.[^@]+$'
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_property_stakeholder ON property_stakeholders (
    property_id, party_type, COALESCE(party_user_id::text, ''), COALESCE(party_org_id::text, ''), COALESCE(lower(party_external_email), ''), role
);

CREATE INDEX IF NOT EXISTS idx_property_stakeholders_property ON property_stakeholders(property_id);
CREATE INDEX IF NOT EXISTS idx_property_stakeholders_user ON property_stakeholders(party_user_id) WHERE party_type = 'user';
CREATE INDEX IF NOT EXISTS idx_property_stakeholders_org ON property_stakeholders(party_org_id) WHERE party_type = 'organization';
