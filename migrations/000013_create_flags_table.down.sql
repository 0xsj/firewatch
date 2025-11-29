-- Drop triggers
DROP TRIGGER IF EXISTS flags_audit_trigger ON flags.flags;
DROP TRIGGER IF EXISTS flags_updated_at ON flags.flags;

-- Drop function
DROP FUNCTION IF EXISTS flags.update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS flags.idx_flags_overrides;
DROP INDEX IF EXISTS flags.idx_flags_rules;
DROP INDEX IF EXISTS flags.idx_flags_created_at;
DROP INDEX IF EXISTS flags.idx_flags_enabled;
DROP INDEX IF EXISTS flags.idx_flags_tenant_id;
DROP INDEX IF EXISTS flags.idx_flags_tenant_key;

-- Drop table
DROP TABLE IF EXISTS flags.flags;

-- Drop schema (only if empty)
DROP SCHEMA IF EXISTS flags;