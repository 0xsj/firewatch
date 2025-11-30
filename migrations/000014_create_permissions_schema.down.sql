-- Drop assignments table first (foreign key dependency)
DROP TABLE IF EXISTS permissions.assignments;

-- Drop roles table
DROP TABLE IF EXISTS permissions.roles;

-- Drop schema
DROP SCHEMA IF EXISTS permissions;