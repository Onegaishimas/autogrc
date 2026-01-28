-- Migration: Create ServiceNow Connections Table
-- Feature: F1 - ServiceNow GRC Connection
-- Date: 2026-01-27

-- Create auth_method enum type
DO $$ BEGIN
    CREATE TYPE auth_method AS ENUM ('basic', 'oauth');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create connection_status enum type
DO $$ BEGIN
    CREATE TYPE connection_status AS ENUM ('success', 'failure', 'pending', 'unknown');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create servicenow_connections table
CREATE TABLE IF NOT EXISTS servicenow_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_url VARCHAR(255) NOT NULL,
    auth_method auth_method NOT NULL DEFAULT 'basic',

    -- Basic Auth (encrypted)
    username VARCHAR(255),
    password_encrypted BYTEA,
    password_nonce BYTEA,

    -- OAuth (encrypted)
    oauth_client_id VARCHAR(255),
    oauth_client_secret_encrypted BYTEA,
    oauth_client_secret_nonce BYTEA,
    oauth_token_url VARCHAR(255),

    -- Status tracking
    is_active BOOLEAN DEFAULT true,
    last_test_at TIMESTAMPTZ,
    last_test_status connection_status DEFAULT 'unknown',
    last_test_message TEXT,
    last_test_instance_version VARCHAR(50),

    -- Audit
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID,
    updated_by UUID
);

-- Create partial unique index alternative (for databases that don't support partial unique constraints)
-- This ensures only one row can have is_active = true
CREATE UNIQUE INDEX IF NOT EXISTS idx_connections_single_active
    ON servicenow_connections (is_active)
    WHERE is_active = true;

-- Create index on is_active for quick lookups
CREATE INDEX IF NOT EXISTS idx_connections_active
    ON servicenow_connections (is_active)
    WHERE is_active = true;

-- Add comment for documentation
COMMENT ON TABLE servicenow_connections IS 'Stores ServiceNow GRC connection configuration with encrypted credentials';
COMMENT ON COLUMN servicenow_connections.password_encrypted IS 'AES-256-GCM encrypted password';
COMMENT ON COLUMN servicenow_connections.password_nonce IS 'Unique nonce used for password encryption';
COMMENT ON COLUMN servicenow_connections.oauth_client_secret_encrypted IS 'AES-256-GCM encrypted OAuth client secret';
COMMENT ON COLUMN servicenow_connections.oauth_client_secret_nonce IS 'Unique nonce used for OAuth secret encryption';
