-- Create sessions table in identity schema
CREATE TABLE identity.sessions (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    tenant_id       VARCHAR(255) NOT NULL,
    provider        VARCHAR(50) NOT NULL,
    status          VARCHAR(50) NOT NULL DEFAULT 'active',
    refresh_token   VARCHAR(255) NOT NULL,
    
    -- Client info
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    device_id       VARCHAR(255),
    
    -- Timestamps
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at      TIMESTAMPTZ,
    
    -- Constraints
    CONSTRAINT sessions_status_check CHECK (status IN ('active', 'expired', 'revoked', 'logged_out'))
);

-- Indexes
CREATE INDEX idx_sessions_user_id ON identity.sessions(user_id);
CREATE INDEX idx_sessions_tenant_id ON identity.sessions(tenant_id);
CREATE INDEX idx_sessions_refresh_token ON identity.sessions(refresh_token);
CREATE INDEX idx_sessions_status ON identity.sessions(status);
CREATE INDEX idx_sessions_expires_at ON identity.sessions(expires_at);

-- Composite index for finding active sessions by user
CREATE INDEX idx_sessions_user_active ON identity.sessions(user_id, status) WHERE status = 'active';

-- Composite index for cleanup of expired sessions
CREATE INDEX idx_sessions_expired ON identity.sessions(status, expires_at) WHERE status = 'active';

-- Comments
COMMENT ON TABLE identity.sessions IS 'User authentication sessions';
COMMENT ON COLUMN identity.sessions.provider IS 'Auth provider: password, google, github, magic_link';
COMMENT ON COLUMN identity.sessions.status IS 'Session status: active, expired, revoked, logged_out';
COMMENT ON COLUMN identity.sessions.refresh_token IS 'Opaque refresh token for token rotation';