package sync

import (
	"time"

	"github.com/google/uuid"
)

// DiscoveredSystemResponse represents a system found in ServiceNow.
type DiscoveredSystemResponse struct {
	SNSysID     string `json:"sn_sys_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Owner       string `json:"owner,omitempty"`
	IsImported  bool   `json:"is_imported"`
}

// DiscoverSystemsResponse is the response for system discovery.
type DiscoverSystemsResponse struct {
	Systems []DiscoveredSystemResponse `json:"systems"`
	Count   int                        `json:"count"`
}

// LocalSystemResponse represents an imported system.
type LocalSystemResponse struct {
	ID             uuid.UUID  `json:"id"`
	SNSysID        string     `json:"sn_sys_id"`
	Name           string     `json:"name"`
	Description    string     `json:"description,omitempty"`
	Acronym        string     `json:"acronym,omitempty"`
	Owner          string     `json:"owner,omitempty"`
	Status         string     `json:"status"`
	ControlCount   int        `json:"control_count"`
	StatementCount int        `json:"statement_count"`
	ModifiedCount  int        `json:"modified_count"`
	LastPullAt     *time.Time `json:"last_pull_at,omitempty"`
	LastPushAt     *time.Time `json:"last_push_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ListSystemsResponse is the response for listing local systems.
type ListSystemsResponse struct {
	Systems    []LocalSystemResponse `json:"systems"`
	TotalCount int                   `json:"total_count"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// ImportSystemsRequest is the request to import systems.
type ImportSystemsRequest struct {
	SNSysIDs []string `json:"sn_sys_ids"`
}

// ImportSystemsResponse is the response after importing systems.
type ImportSystemsResponse struct {
	Imported []LocalSystemResponse `json:"imported"`
	Count    int                   `json:"count"`
}

// StartPullRequest is the request to start a pull operation.
type StartPullRequest struct {
	SystemIDs []uuid.UUID `json:"system_ids"`
}

// PullJobResponse represents a pull job.
type PullJobResponse struct {
	ID          uuid.UUID         `json:"id"`
	SystemIDs   []uuid.UUID       `json:"system_ids"`
	Status      string            `json:"status"`
	Progress    PullProgressResponse `json:"progress"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	Error       string            `json:"error,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

// PullProgressResponse represents pull operation progress.
type PullProgressResponse struct {
	TotalSystems      int      `json:"total_systems"`
	CompletedSystems  int      `json:"completed_systems"`
	TotalControls     int      `json:"total_controls"`
	CompletedControls int      `json:"completed_controls"`
	TotalStatements   int      `json:"total_statements"`
	CompletedStatements int    `json:"completed_statements"`
	CurrentSystem     string   `json:"current_system,omitempty"`
	Errors            []string `json:"errors,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
