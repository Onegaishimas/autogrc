-- Audit events table for tracking all sync operations
CREATE TABLE IF NOT EXISTS audit_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'success',
    details JSONB DEFAULT '{}',
    user_email VARCHAR(255),
    ip_address VARCHAR(45),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_audit_event_type ON audit_events(event_type);
CREATE INDEX IF NOT EXISTS idx_audit_entity ON audit_events(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_events(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_status ON audit_events(status);
CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_events(user_email) WHERE user_email IS NOT NULL;

-- Full-text search on details
CREATE INDEX IF NOT EXISTS idx_audit_details_gin ON audit_events USING GIN (details);

-- Composite index for common filter combination
CREATE INDEX IF NOT EXISTS idx_audit_common_filters ON audit_events(event_type, entity_type, created_at DESC);
