-- Create permissions schema
CREATE SCHEMA IF NOT EXISTS permissions;

-- Roles table
CREATE TABLE permissions.roles (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '[]',
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1,

    CONSTRAINT unique_role_name_per_tenant UNIQUE(tenant_id, name)
);

-- Role assignments table
CREATE TABLE permissions.assignments (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES permissions.roles(id) ON DELETE CASCADE,
    tenant_id VARCHAR(255) NOT NULL,
    resource_id VARCHAR(255),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL
);

-- Unique index for assignments (using expression for nullable resource_id)
CREATE UNIQUE INDEX idx_unique_user_role_scope 
    ON permissions.assignments(user_id, role_id, tenant_id, COALESCE(resource_id, ''));

-- Indexes for roles
CREATE INDEX idx_roles_tenant_id ON permissions.roles(tenant_id);
CREATE INDEX idx_roles_is_system ON permissions.roles(is_system);
CREATE INDEX idx_roles_name ON permissions.roles(name);

-- Indexes for assignments
CREATE INDEX idx_assignments_user_id ON permissions.assignments(user_id);
CREATE INDEX idx_assignments_role_id ON permissions.assignments(role_id);
CREATE INDEX idx_assignments_tenant_id ON permissions.assignments(tenant_id);
CREATE INDEX idx_assignments_expires_at ON permissions.assignments(expires_at) WHERE expires_at IS NOT NULL;

-- Insert system roles
INSERT INTO permissions.roles (id, tenant_id, name, description, permissions, is_system) VALUES
    ('00000000-0000-0000-0000-000000000001', '*', 'super_admin', 'Full system access', '["*:*"]', true),
    ('00000000-0000-0000-0000-000000000002', '*', 'admin', 'Tenant administrator', '["manage:users", "manage:flags", "read:audit_logs", "manage:roles", "manage:api_keys"]', true),
    ('00000000-0000-0000-0000-000000000003', '*', 'moderator', 'Content moderator', '["read:self", "update:self", "read:users", "read:flags", "read:audit_logs"]', true),
    ('00000000-0000-0000-0000-000000000004', '*', 'user', 'Standard user', '["read:self", "update:self"]', true),
    ('00000000-0000-0000-0000-000000000005', '*', 'guest', 'Guest user', '[]', true);