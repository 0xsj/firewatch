-- Remove owner_email column from tenants table
ALTER TABLE tenants DROP COLUMN IF EXISTS owner_email;