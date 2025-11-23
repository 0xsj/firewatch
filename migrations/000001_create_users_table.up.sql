-- Create users table in identity schema
CREATE TABLE IF NOT EXISTS identity.users (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Tenant (multi-tenancy support)
    tenant_id VARCHAR(255) NOT NULL,
    
    -- Authentication
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),  -- NULL for passwordless users
    
    -- Status & Role
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    
    -- Email Verification
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    email_verified_at TIMESTAMPTZ,
    
    -- Security
    failed_login_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    last_login_ip VARCHAR(45),
    last_login_user_agent TEXT,
    
    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT users_email_not_empty CHECK (email <> ''),
    CONSTRAINT users_status_valid CHECK (status IN ('pending', 'active', 'suspended', 'deleted')),
    CONSTRAINT users_role_valid CHECK (role IN ('user', 'admin', 'moderator')),
    CONSTRAINT users_failed_attempts_non_negative CHECK (failed_login_attempts >= 0)
);

-- Indexes for performance
CREATE UNIQUE INDEX idx_users_tenant_email ON identity.users(tenant_id, email);
CREATE INDEX idx_users_tenant_id ON identity.users(tenant_id);
CREATE INDEX idx_users_email ON identity.users(email);
CREATE INDEX idx_users_status ON identity.users(status);
CREATE INDEX idx_users_created_at ON identity.users(created_at DESC);
CREATE INDEX idx_users_email_verified ON identity.users(email_verified);

-- Trigger to update updated_at automatically
CREATE OR REPLACE FUNCTION identity.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON identity.users
    FOR EACH ROW
    EXECUTE FUNCTION identity.update_updated_at_column();

-- Audit trigger (logs all changes)
CREATE TRIGGER users_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON identity.users
    FOR EACH ROW
    EXECUTE FUNCTION audit.audit_trigger();

-- Comments for documentation
COMMENT ON TABLE identity.users IS 'User accounts with multi-tenant support';
COMMENT ON COLUMN identity.users.id IS 'Unique user identifier (UUID)';
COMMENT ON COLUMN identity.users.tenant_id IS 'Tenant identifier for multi-tenancy';
COMMENT ON COLUMN identity.users.email IS 'User email address (unique per tenant)';
COMMENT ON COLUMN identity.users.password_hash IS 'Bcrypt password hash (NULL for passwordless)';
COMMENT ON COLUMN identity.users.status IS 'Account status: pending, active, suspended, deleted';
COMMENT ON COLUMN identity.users.role IS 'User role: user, admin, moderator';
COMMENT ON COLUMN identity.users.email_verified IS 'Whether email has been verified';
COMMENT ON COLUMN identity.users.failed_login_attempts IS 'Number of consecutive failed login attempts';
COMMENT ON COLUMN identity.users.locked_until IS 'Account locked until this timestamp (for failed attempts)';