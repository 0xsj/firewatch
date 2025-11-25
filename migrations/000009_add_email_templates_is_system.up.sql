-- Add missing columns to email templates
ALTER TABLE email.templates ADD COLUMN IF NOT EXISTS is_system BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE email.templates ADD COLUMN IF NOT EXISTS activated_at TIMESTAMPTZ;
ALTER TABLE email.templates ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ;

-- Add index for system templates lookup
CREATE INDEX IF NOT EXISTS idx_templates_is_system ON email.templates (is_system) WHERE is_system = true;