-- Create oauth_accounts table in identity schema
CREATE TABLE identity.oauth_accounts (
    id                UUID PRIMARY KEY,
    user_id           UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    tenant_id         VARCHAR(255) NOT NULL,
    
    -- Provider info
    provider          VARCHAR(50) NOT NULL,  -- 'google', 'github', etc.
    provider_user_id  VARCHAR(255) NOT NULL, -- User ID from OAuth provider
    email             VARCHAR(255) NOT NULL,
    
    -- Tokens (should be encrypted in production)
    access_token      TEXT,
    refresh_token     TEXT,
    token_expires_at  TIMESTAMPTZ,
    
    -- Provider profile data (stored as JSON)
    profile_data      JSONB,
    
    -- Timestamps
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at      TIMESTAMPTZ,
    
    -- Constraints
    UNIQUE(provider, provider_user_id),
    UNIQUE(user_id, provider)
);

-- Indexes
CREATE INDEX idx_oauth_accounts_user_id ON identity.oauth_accounts(user_id);
CREATE INDEX idx_oauth_accounts_provider ON identity.oauth_accounts(provider);
CREATE INDEX idx_oauth_accounts_provider_user_id ON identity.oauth_accounts(provider, provider_user_id);
CREATE INDEX idx_oauth_accounts_email ON identity.oauth_accounts(email);

-- Comments
COMMENT ON TABLE identity.oauth_accounts IS 'OAuth provider accounts linked to users';
COMMENT ON COLUMN identity.oauth_accounts.provider IS 'OAuth provider name (google, github, etc.)';
COMMENT ON COLUMN identity.oauth_accounts.provider_user_id IS 'User ID from the OAuth provider';
COMMENT ON COLUMN identity.oauth_accounts.profile_data IS 'JSON data from OAuth provider profile';