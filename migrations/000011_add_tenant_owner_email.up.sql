-- Add owner_email column to tenants table
ALTER TABLE tenants ADD COLUMN owner_email VARCHAR(255);

-- Comment
COMMENT ON COLUMN tenants.owner_email IS 'Email of the tenant owner (denormalized for notifications)';