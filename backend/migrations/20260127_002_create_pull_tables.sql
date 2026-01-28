-- Migration: Create Pull Tables for Control Package Storage
-- Feature: F2 - Control Package Pull
-- Date: 2026-01-27

-- =============================================================================
-- ENUM TYPES
-- =============================================================================

-- Sync status for tracking local modifications
DO $$ BEGIN
    CREATE TYPE sync_status AS ENUM ('synced', 'modified', 'conflict', 'new');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Job status for pull operations
DO $$ BEGIN
    CREATE TYPE job_status AS ENUM ('pending', 'running', 'completed', 'failed', 'cancelled');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- =============================================================================
-- SYSTEMS TABLE
-- =============================================================================
-- Stores ServiceNow systems (cmdb_ci_service or similar)
-- DEMO MODE: Currently maps to incident caller reference, will map to IRM systems

CREATE TABLE IF NOT EXISTS systems (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- ServiceNow identifiers
    sn_sys_id VARCHAR(32) NOT NULL,

    -- System information
    name VARCHAR(255) NOT NULL,
    description TEXT,
    acronym VARCHAR(50),
    owner VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active',

    -- Sync metadata
    sn_updated_on TIMESTAMPTZ,
    last_pull_at TIMESTAMPTZ,
    last_push_at TIMESTAMPTZ,

    -- Audit
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Unique constraint on ServiceNow sys_id
CREATE UNIQUE INDEX IF NOT EXISTS idx_systems_sn_sys_id
    ON systems (sn_sys_id);

-- Index for listing active systems
CREATE INDEX IF NOT EXISTS idx_systems_status
    ON systems (status);

COMMENT ON TABLE systems IS 'Stores systems/applications from ServiceNow for control mapping';
COMMENT ON COLUMN systems.sn_sys_id IS 'ServiceNow sys_id - unique identifier from ServiceNow';

-- =============================================================================
-- CONTROLS TABLE
-- =============================================================================
-- Stores NIST 800-53 controls associated with systems
-- DEMO MODE: Maps from incident category, will map to sn_compliance_control

CREATE TABLE IF NOT EXISTS controls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Relationship
    system_id UUID NOT NULL REFERENCES systems(id) ON DELETE CASCADE,

    -- ServiceNow identifiers
    sn_sys_id VARCHAR(32) NOT NULL,

    -- Control information
    control_id VARCHAR(50) NOT NULL,      -- e.g., "AC-1", "SC-7"
    control_name VARCHAR(255) NOT NULL,
    control_family VARCHAR(50),            -- e.g., "AC", "SC"
    description TEXT,

    -- Implementation details
    implementation_status VARCHAR(50) DEFAULT 'not_implemented',
    responsible_role VARCHAR(255),

    -- Sync metadata
    sn_updated_on TIMESTAMPTZ,
    last_pull_at TIMESTAMPTZ,
    last_push_at TIMESTAMPTZ,

    -- Audit
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Unique constraint on system + ServiceNow sys_id
CREATE UNIQUE INDEX IF NOT EXISTS idx_controls_system_sn_sys_id
    ON controls (system_id, sn_sys_id);

-- Index for filtering by system
CREATE INDEX IF NOT EXISTS idx_controls_system_id
    ON controls (system_id);

-- Index for filtering by control family
CREATE INDEX IF NOT EXISTS idx_controls_family
    ON controls (control_family);

COMMENT ON TABLE controls IS 'Stores NIST 800-53 controls mapped to systems';
COMMENT ON COLUMN controls.control_id IS 'NIST 800-53 control identifier (e.g., AC-1, SC-7)';
COMMENT ON COLUMN controls.control_family IS 'Control family prefix (e.g., AC, SC, AU)';

-- =============================================================================
-- STATEMENTS TABLE
-- =============================================================================
-- Stores control implementation statements (the actual content users edit)
-- DEMO MODE: Maps from incident short_description, will map to sn_compliance_policy_statement

CREATE TABLE IF NOT EXISTS statements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Relationship
    control_id UUID NOT NULL REFERENCES controls(id) ON DELETE CASCADE,

    -- ServiceNow identifiers
    sn_sys_id VARCHAR(32) NOT NULL,

    -- Statement information
    statement_type VARCHAR(50) DEFAULT 'implementation',  -- implementation, assessment, etc.

    -- Content - remote (from ServiceNow)
    remote_content TEXT,
    remote_updated_at TIMESTAMPTZ,

    -- Content - local (user edits)
    local_content TEXT,
    is_modified BOOLEAN DEFAULT false,
    modified_at TIMESTAMPTZ,
    modified_by UUID,

    -- Sync status
    sync_status sync_status DEFAULT 'synced',
    conflict_resolved_at TIMESTAMPTZ,
    conflict_resolved_by UUID,

    -- Sync metadata
    sn_updated_on TIMESTAMPTZ,
    last_pull_at TIMESTAMPTZ,
    last_push_at TIMESTAMPTZ,

    -- Audit
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Unique constraint on control + ServiceNow sys_id
CREATE UNIQUE INDEX IF NOT EXISTS idx_statements_control_sn_sys_id
    ON statements (control_id, sn_sys_id);

-- Index for filtering by control
CREATE INDEX IF NOT EXISTS idx_statements_control_id
    ON statements (control_id);

-- Index for finding modified statements
CREATE INDEX IF NOT EXISTS idx_statements_modified
    ON statements (is_modified)
    WHERE is_modified = true;

-- Index for finding conflicts
CREATE INDEX IF NOT EXISTS idx_statements_conflicts
    ON statements (sync_status)
    WHERE sync_status = 'conflict';

COMMENT ON TABLE statements IS 'Stores control implementation statements - the actual content users author';
COMMENT ON COLUMN statements.remote_content IS 'Content as fetched from ServiceNow';
COMMENT ON COLUMN statements.local_content IS 'Locally edited content (takes precedence when is_modified=true)';
COMMENT ON COLUMN statements.sync_status IS 'Sync state: synced=matches remote, modified=local changes, conflict=both changed';

-- =============================================================================
-- PULL JOBS TABLE
-- =============================================================================
-- Tracks background pull operations

CREATE TABLE IF NOT EXISTS pull_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Job configuration
    system_ids UUID[] NOT NULL,

    -- Job status
    status job_status DEFAULT 'pending',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,

    -- Progress tracking (JSONB for flexibility)
    progress JSONB DEFAULT '{
        "total_systems": 0,
        "completed_systems": 0,
        "total_controls": 0,
        "completed_controls": 0,
        "total_statements": 0,
        "completed_statements": 0,
        "current_system": null,
        "errors": []
    }'::jsonb,

    -- Error details
    error_message TEXT,

    -- Audit
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID
);

-- Index for finding active jobs
CREATE INDEX IF NOT EXISTS idx_pull_jobs_status
    ON pull_jobs (status)
    WHERE status IN ('pending', 'running');

-- Index for job history
CREATE INDEX IF NOT EXISTS idx_pull_jobs_created_at
    ON pull_jobs (created_at DESC);

COMMENT ON TABLE pull_jobs IS 'Tracks background pull job operations';
COMMENT ON COLUMN pull_jobs.progress IS 'JSON object tracking pull progress for UI display';
