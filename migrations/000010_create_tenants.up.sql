-- Tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY,
    slug VARCHAR(63) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    plan VARCHAR(50) NOT NULL DEFAULT 'free',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    settings JSONB NOT NULL DEFAULT '{}',
    owner_id UUID NOT NULL,
    billing_id VARCHAR(255),
    trial_ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1,
    
    -- Constraints
    CONSTRAINT tenants_slug_format CHECK (slug ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$' OR slug ~ '^[a-z0-9]$'),
    CONSTRAINT tenants_slug_length CHECK (LENGTH(slug) >= 3 AND LENGTH(slug) <= 63),
    CONSTRAINT tenants_plan_valid CHECK (plan IN ('free', 'starter', 'pro', 'enterprise')),
    CONSTRAINT tenants_status_valid CHECK (status IN ('trialing', 'active', 'suspended', 'cancelled', 'deleted'))
);

-- Indexes
CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_owner_id ON tenants(owner_id);
CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_plan ON tenants(plan);
CREATE INDEX idx_tenants_created_at ON tenants(created_at DESC);

-- Index for search on name and slug
CREATE INDEX idx_tenants_name_search ON tenants USING gin(to_tsvector('english', name));

-- Comments
COMMENT ON TABLE tenants IS 'Multi-tenant organizations/workspaces';
COMMENT ON COLUMN tenants.id IS 'Unique tenant identifier (UUIDv7)';
COMMENT ON COLUMN tenants.slug IS 'URL-friendly unique identifier';
COMMENT ON COLUMN tenants.name IS 'Display name of the tenant';
COMMENT ON COLUMN tenants.plan IS 'Subscription plan: free, starter, pro, enterprise';
COMMENT ON COLUMN tenants.status IS 'Tenant status: trialing, active, suspended, cancelled, deleted';
COMMENT ON COLUMN tenants.settings IS 'Tenant-specific configuration (JSONB)';
COMMENT ON COLUMN tenants.owner_id IS 'User ID of the tenant owner';
COMMENT ON COLUMN tenants.billing_id IS 'External billing system identifier (e.g., Stripe customer ID)';
COMMENT ON COLUMN tenants.trial_ends_at IS 'Trial period expiration timestamp';
COMMENT ON COLUMN tenants.version IS 'Optimistic locking version';