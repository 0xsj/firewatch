-- Create audit schema
CREATE SCHEMA IF NOT EXISTS audit;

-- Create audit entries table
CREATE TABLE audit.entries (
    id              UUID PRIMARY KEY,
    event_id        VARCHAR(255) NOT NULL,
    event_type      VARCHAR(255) NOT NULL,
    source          VARCHAR(255) NOT NULL,
    timestamp       TIMESTAMPTZ NOT NULL,
    
    -- Context
    tenant_id       VARCHAR(255),
    user_id         VARCHAR(255),
    correlation_id  VARCHAR(255),
    
    -- Event data (stored as JSONB for flexibility)
    payload         JSONB NOT NULL DEFAULT '{}',
    metadata        JSONB NOT NULL DEFAULT '{}',
    
    -- Record keeping
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_audit_entries_event_type ON audit.entries(event_type);
CREATE INDEX idx_audit_entries_source ON audit.entries(source);
CREATE INDEX idx_audit_entries_tenant_id ON audit.entries(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX idx_audit_entries_user_id ON audit.entries(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_audit_entries_correlation_id ON audit.entries(correlation_id) WHERE correlation_id IS NOT NULL;
CREATE INDEX idx_audit_entries_timestamp ON audit.entries(timestamp DESC);

-- Composite index for common filter combinations
CREATE INDEX idx_audit_entries_tenant_timestamp ON audit.entries(tenant_id, timestamp DESC) WHERE tenant_id IS NOT NULL;

-- Comment
COMMENT ON TABLE audit.entries IS 'Audit trail of all domain events in the system';