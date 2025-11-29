-- Initialize database schema
-- This runs automatically when PostgreSQL container starts for the first time

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create schemas
CREATE SCHEMA IF NOT EXISTS identity;
CREATE SCHEMA IF NOT EXISTS audit;

-- Set search path
ALTER DATABASE hexagonal_go SET search_path TO identity, public;

-- Create audit trigger function (for tracking changes)
CREATE OR REPLACE FUNCTION audit.audit_trigger()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit.audit_log (
            schema_name,
            table_name,
            operation,
            new_data,
            changed_at
        ) VALUES (
            TG_TABLE_SCHEMA,
            TG_TABLE_NAME,
            TG_OP,
            row_to_json(NEW),
            NOW()
        );
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit.audit_log (
            schema_name,
            table_name,
            operation,
            old_data,
            new_data,
            changed_at
        ) VALUES (
            TG_TABLE_SCHEMA,
            TG_TABLE_NAME,
            TG_OP,
            row_to_json(OLD),
            row_to_json(NEW),
            NOW()
        );
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit.audit_log (
            schema_name,
            table_name,
            operation,
            old_data,
            changed_at
        ) VALUES (
            TG_TABLE_SCHEMA,
            TG_TABLE_NAME,
            TG_OP,
            row_to_json(OLD),
            NOW()
        );
        RETURN OLD;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Create audit log table
CREATE TABLE IF NOT EXISTS audit.audit_log (
    id BIGSERIAL PRIMARY KEY,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    operation TEXT NOT NULL,
    old_data JSONB,
    new_data JSONB,
    changed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    changed_by TEXT
);

-- Create index on audit log
CREATE INDEX IF NOT EXISTS idx_audit_log_table ON audit.audit_log(table_name);
CREATE INDEX IF NOT EXISTS idx_audit_log_changed_at ON audit.audit_log(changed_at);

-- Grant permissions
GRANT USAGE ON SCHEMA identity TO hexagonal;
GRANT USAGE ON SCHEMA audit TO hexagonal;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA identity TO hexagonal;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA audit TO hexagonal;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA identity TO hexagonal;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA audit TO hexagonal;

-- Set default privileges for future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA identity GRANT ALL ON TABLES TO hexagonal;
ALTER DEFAULT PRIVILEGES IN SCHEMA audit GRANT ALL ON TABLES TO hexagonal;
ALTER DEFAULT PRIVILEGES IN SCHEMA identity GRANT ALL ON SEQUENCES TO hexagonal;
ALTER DEFAULT PRIVILEGES IN SCHEMA audit GRANT ALL ON SEQUENCES TO hexagonal;