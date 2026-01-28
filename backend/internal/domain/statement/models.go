package statement

import (
	"time"

	"github.com/google/uuid"
)

// SyncStatus represents the synchronization state of a statement.
type SyncStatus string

const (
	SyncStatusSynced   SyncStatus = "synced"   // Local matches remote
	SyncStatusModified SyncStatus = "modified" // Local has changes
	SyncStatusConflict SyncStatus = "conflict" // Both local and remote changed
	SyncStatusNew      SyncStatus = "new"      // New local statement
)

// Statement represents a control implementation statement.
// In IRM, this maps to sn_compliance_policy_statement.
// DEMO MODE: Maps from incidents.
type Statement struct {
	ID        uuid.UUID `json:"id"`
	ControlID uuid.UUID `json:"control_id"`
	SNSysID   string    `json:"sn_sys_id"`

	StatementType string `json:"statement_type"` // implementation, assessment, etc.

	// Remote content (from ServiceNow)
	RemoteContent   string     `json:"remote_content,omitempty"`
	RemoteUpdatedAt *time.Time `json:"remote_updated_at,omitempty"`

	// Local content (user edits)
	LocalContent string     `json:"local_content,omitempty"`
	IsModified   bool       `json:"is_modified"`
	ModifiedAt   *time.Time `json:"modified_at,omitempty"`
	ModifiedBy   *uuid.UUID `json:"modified_by,omitempty"`

	// Sync status
	SyncStatus         SyncStatus `json:"sync_status"`
	ConflictResolvedAt *time.Time `json:"conflict_resolved_at,omitempty"`
	ConflictResolvedBy *uuid.UUID `json:"conflict_resolved_by,omitempty"`

	// Sync metadata
	SNUpdatedOn *time.Time `json:"sn_updated_on,omitempty"`
	LastPullAt  *time.Time `json:"last_pull_at,omitempty"`
	LastPushAt  *time.Time `json:"last_push_at,omitempty"`

	// Audit
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetContent returns the effective content (local if modified, otherwise remote).
func (s *Statement) GetContent() string {
	if s.IsModified && s.LocalContent != "" {
		return s.LocalContent
	}
	return s.RemoteContent
}

// ListParams holds parameters for listing statements.
type ListParams struct {
	ControlID  uuid.UUID  `json:"control_id"`
	SystemID   uuid.UUID  `json:"system_id"`   // Filter by system (joins through controls)
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	SyncStatus SyncStatus `json:"sync_status,omitempty"`
	Search     string     `json:"search,omitempty"`
}

// ListResult holds the result of listing statements.
type ListResult struct {
	Statements []Statement `json:"statements"`
	TotalCount int         `json:"total_count"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// UpsertInput holds data for creating or updating a statement from ServiceNow.
type UpsertInput struct {
	ControlID     uuid.UUID
	SNSysID       string
	StatementType string
	RemoteContent string
	SNUpdatedOn   *time.Time
}

// UpdateInput holds data for updating local content.
type UpdateInput struct {
	ID           uuid.UUID
	LocalContent string
	ModifiedBy   *uuid.UUID
}

// ConflictResolution represents how a conflict was resolved.
type ConflictResolution string

const (
	ConflictResolutionKeepLocal  ConflictResolution = "keep_local"
	ConflictResolutionKeepRemote ConflictResolution = "keep_remote"
	ConflictResolutionMerge      ConflictResolution = "merge"
)

// ResolveConflictInput holds data for resolving a sync conflict.
type ResolveConflictInput struct {
	ID           uuid.UUID
	Resolution   ConflictResolution
	MergedContent string // Used when Resolution is ConflictResolutionMerge
	ResolvedBy   *uuid.UUID
}
