-- Remove index
DROP INDEX IF EXISTS email.idx_templates_is_system;

-- Remove columns
ALTER TABLE email.templates DROP COLUMN IF EXISTS is_system;
ALTER TABLE email.templates DROP COLUMN IF EXISTS activated_at;
ALTER TABLE email.templates DROP COLUMN IF EXISTS archived_at;