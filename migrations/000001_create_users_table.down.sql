-- Drop triggers
DROP TRIGGER IF EXISTS users_audit_trigger ON identity.users;
DROP TRIGGER IF EXISTS users_updated_at ON identity.users;

-- Drop function
DROP FUNCTION IF EXISTS identity.update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS identity.idx_users_email_verified;
DROP INDEX IF EXISTS identity.idx_users_created_at;
DROP INDEX IF EXISTS identity.idx_users_status;
DROP INDEX IF EXISTS identity.idx_users_email;
DROP INDEX IF EXISTS identity.idx_users_tenant_id;
DROP INDEX IF EXISTS identity.idx_users_tenant_email;

-- Drop table
DROP TABLE IF EXISTS identity.users;