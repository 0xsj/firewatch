-- Create flags schema
CREATE SCHEMA IF NOT EXISTS flags;

-- Create flags table
CREATE TABLE IF NOT EXISTS flags.flags (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Tenant (multi-tenancy support)
    tenant_id VARCHAR(255) NOT NULL,

    -- Flag Identity
    key VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',

    -- State
    enabled BOOLEAN NOT NULL DEFAULT FALSE,

    -- Variants (JSON array)
    variants JSONB NOT NULL DEFAULT '[]'::JSONB,
    default_variant VARCHAR(255) NOT NULL DEFAULT '',

    -- Targeting Rules (JSON array)
    rules JSONB NOT NULL DEFAULT '[]'::JSONB,

    -- Overrides (JSON array)
    overrides JSONB NOT NULL DEFAULT '[]'::JSONB,

    -- Optimistic Locking
    version INTEGER NOT NULL DEFAULT 1,

    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT flags_key_not_empty CHECK (key <> ''),
    CONSTRAINT flags_key_format CHECK (key ~ '^[a-z][a-z0-9_-]*$'),
    CONSTRAINT flags_version_positive CHECK (version > 0)
);

-- Indexes for performance
CREATE UNIQUE INDEX idx_flags_tenant_key ON flags.flags(tenant_id, key);
CREATE INDEX idx_flags_tenant_id ON flags.flags(tenant_id);
CREATE INDEX idx_flags_enabled ON flags.flags(enabled);
CREATE INDEX idx_flags_created_at ON flags.flags(created_at DESC);

-- GIN index for JSONB searches (rules and overrides)
CREATE INDEX idx_flags_rules ON flags.flags USING GIN (rules);
CREATE INDEX idx_flags_overrides ON flags.flags USING GIN (overrides);

-- Trigger to update updated_at automatically
CREATE OR REPLACE FUNCTION flags.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER flags_updated_at
    BEFORE UPDATE ON flags.flags
    FOR EACH ROW
    EXECUTE FUNCTION flags.update_updated_at_column();

-- Audit trigger (logs all changes)
CREATE TRIGGER flags_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON flags.flags
    FOR EACH ROW
    EXECUTE FUNCTION audit.audit_trigger();

-- Comments for documentation
COMMENT ON TABLE flags.flags IS 'Feature flags with multi-tenant support';
COMMENT ON COLUMN flags.flags.id IS 'Unique flag identifier (UUID)';
COMMENT ON COLUMN flags.flags.tenant_id IS 'Tenant identifier for multi-tenancy';
COMMENT ON COLUMN flags.flags.key IS 'Unique flag key within tenant (snake_case)';
COMMENT ON COLUMN flags.flags.name IS 'Human-readable flag name';
COMMENT ON COLUMN flags.flags.description IS 'Flag description';
COMMENT ON COLUMN flags.flags.enabled IS 'Whether the flag is enabled';
COMMENT ON COLUMN flags.flags.variants IS 'JSON array of variant definitions';
COMMENT ON COLUMN flags.flags.default_variant IS 'Default variant key when no rules match';
COMMENT ON COLUMN flags.flags.rules IS 'JSON array of targeting rules';
COMMENT ON COLUMN flags.flags.overrides IS 'JSON array of target-specific overrides';
COMMENT ON COLUMN flags.flags.version IS 'Version for optimistic locking';