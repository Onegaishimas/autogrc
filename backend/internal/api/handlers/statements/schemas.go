package statements

import (
	"time"

	"github.com/google/uuid"
)

// StatementResponse represents a statement in API responses.
type StatementResponse struct {
	ID            uuid.UUID  `json:"id"`
	ControlID     uuid.UUID  `json:"control_id"`
	SNSysID       string     `json:"sn_sys_id"`
	StatementType string     `json:"statement_type"`

	// Content
	RemoteContent   string     `json:"remote_content,omitempty"`
	RemoteUpdatedAt *time.Time `json:"remote_updated_at,omitempty"`
	LocalContent    string     `json:"local_content,omitempty"`
	IsModified      bool       `json:"is_modified"`
	ModifiedAt      *time.Time `json:"modified_at,omitempty"`

	// Sync status
	SyncStatus         string     `json:"sync_status"`
	ConflictResolvedAt *time.Time `json:"conflict_resolved_at,omitempty"`

	// Computed field for display
	EffectiveContent string `json:"effective_content"`

	// Timestamps
	LastPullAt *time.Time `json:"last_pull_at,omitempty"`
	LastPushAt *time.Time `json:"last_push_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// ListStatementsResponse is the response for listing statements.
type ListStatementsResponse struct {
	Statements []StatementResponse `json:"statements"`
	TotalCount int                 `json:"total_count"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

// UpdateStatementRequest is the request to update a statement's local content.
type UpdateStatementRequest struct {
	LocalContent string `json:"local_content"`
}

// ResolveConflictRequest is the request to resolve a sync conflict.
type ResolveConflictRequest struct {
	Resolution    string `json:"resolution"` // "keep_local", "keep_remote", "merge"
	MergedContent string `json:"merged_content,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// ModifiedStatementsResponse is the response for listing modified statements.
type ModifiedStatementsResponse struct {
	Statements []StatementResponse `json:"statements"`
	Count      int                 `json:"count"`
}

// ConflictStatementsResponse is the response for listing conflict statements.
type ConflictStatementsResponse struct {
	Statements []StatementResponse `json:"statements"`
	Count      int                 `json:"count"`
}
