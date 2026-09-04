CREATE TABLE IF NOT EXISTS properties (
    id UUID PRIMARY KEY,
    display_name TEXT NOT NULL CHECK (char_length(display_name) BETWEEN 2 AND 100),
    classification TEXT NOT NULL CHECK (classification IN ('house','apartment','condominium','multi_unit','commercial','other')),
    address_street TEXT NOT NULL,
    address_city TEXT NOT NULL,
    address_state TEXT,
    address_postal_code TEXT NOT NULL,
    address_country TEXT NOT NULL,
    rental_cadence TEXT NOT NULL CHECK (rental_cadence IN ('monthly','weekly','daily','annual')),
    standard_rent BIGINT NOT NULL CHECK (standard_rent > 0),
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_properties_created_by ON properties(created_by);
