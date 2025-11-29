-- Drop indexes
DROP INDEX IF EXISTS idx_tenants_name_search;
DROP INDEX IF EXISTS idx_tenants_created_at;
DROP INDEX IF EXISTS idx_tenants_plan;
DROP INDEX IF EXISTS idx_tenants_status;
DROP INDEX IF EXISTS idx_tenants_owner_id;
DROP INDEX IF EXISTS idx_tenants_slug;

-- Drop table
DROP TABLE IF EXISTS tenants;