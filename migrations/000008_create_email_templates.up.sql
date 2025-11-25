-- Email templates schema
CREATE SCHEMA IF NOT EXISTS email;

-- Email templates table
CREATE TABLE email.templates (
    id UUID PRIMARY KEY,
    tenant_id VARCHAR(50),
    slug VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    subject VARCHAR(500) NOT NULL,
    body_html TEXT NOT NULL,
    body_text TEXT,
    variables JSONB NOT NULL DEFAULT '[]',
    locale VARCHAR(10) NOT NULL DEFAULT 'en',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    version INT NOT NULL DEFAULT 1,
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uq_email_templates_tenant_slug_locale UNIQUE(tenant_id, slug, locale)
);

-- Indexes
CREATE INDEX idx_email_templates_tenant_id ON email.templates(tenant_id);
CREATE INDEX idx_email_templates_slug ON email.templates(slug);
CREATE INDEX idx_email_templates_locale ON email.templates(locale);
CREATE INDEX idx_email_templates_status ON email.templates(status);
CREATE INDEX idx_email_templates_tenant_status ON email.templates(tenant_id, status);

-- Comments
COMMENT ON TABLE email.templates IS 'Email templates for transactional emails';
COMMENT ON COLUMN email.templates.tenant_id IS 'Tenant ID, NULL for system-wide templates';
COMMENT ON COLUMN email.templates.slug IS 'Template identifier (e.g., welcome, password_reset)';
COMMENT ON COLUMN email.templates.variables IS 'JSON array of variable definitions with name, type, required, default, description';
COMMENT ON COLUMN email.templates.locale IS 'ISO locale code (e.g., en, es, fr)';
COMMENT ON COLUMN email.templates.status IS 'Template status: draft, active, archived';