-- Create password_reset_tokens table in identity schema
CREATE TABLE identity.password_reset_tokens (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    tenant_id       VARCHAR(255) NOT NULL,
    token           VARCHAR(255) NOT NULL UNIQUE,
    
    -- Metadata
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    
    -- State
    expires_at      TIMESTAMPTZ NOT NULL,
    used_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_password_reset_tokens_user_id ON identity.password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_tokens_token ON identity.password_reset_tokens(token);
CREATE INDEX idx_password_reset_tokens_expires_at ON identity.password_reset_tokens(expires_at);
CREATE INDEX idx_password_reset_tokens_used_at ON identity.password_reset_tokens(used_at);

-- Composite index for finding valid tokens
CREATE INDEX idx_password_reset_tokens_valid ON identity.password_reset_tokens(token, used_at, expires_at);

-- Comments
COMMENT ON TABLE identity.password_reset_tokens IS 'Password reset tokens for secure password recovery';
COMMENT ON COLUMN identity.password_reset_tokens.token IS 'Secure random token sent via email';
COMMENT ON COLUMN identity.password_reset_tokens.used_at IS 'Timestamp when token was used (tokens are single-use)';