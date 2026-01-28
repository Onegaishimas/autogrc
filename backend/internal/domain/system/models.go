package system

import (
	"time"

	"github.com/google/uuid"
)

// System represents a system/application that contains controls.
// In IRM, this maps to a scoped item or business entity.
// DEMO MODE: Maps from incident categories.
type System struct {
	ID          uuid.UUID  `json:"id"`
	SNSysID     string     `json:"sn_sys_id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Acronym     string     `json:"acronym,omitempty"`
	Owner       string     `json:"owner,omitempty"`
	Status      string     `json:"status"`

	// Sync metadata
	SNUpdatedOn *time.Time `json:"sn_updated_on,omitempty"`
	LastPullAt  *time.Time `json:"last_pull_at,omitempty"`
	LastPushAt  *time.Time `json:"last_push_at,omitempty"`

	// Audit
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SystemWithStats includes additional statistics about the system.
type SystemWithStats struct {
	System
	ControlCount   int `json:"control_count"`
	StatementCount int `json:"statement_count"`
	ModifiedCount  int `json:"modified_count"` // Locally modified statements
}

// DiscoveredSystem represents a system found in ServiceNow that may not be imported yet.
type DiscoveredSystem struct {
	SNSysID     string `json:"sn_sys_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Owner       string `json:"owner,omitempty"`
	IsImported  bool   `json:"is_imported"` // True if already in local database
}

// ListParams holds parameters for listing systems.
type ListParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Search   string `json:"search,omitempty"`
	Status   string `json:"status,omitempty"`
}

// ListResult holds the result of listing systems.
type ListResult struct {
	Systems    []SystemWithStats `json:"systems"`
	TotalCount int               `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// UpsertInput holds data for creating or updating a system.
type UpsertInput struct {
	SNSysID     string
	Name        string
	Description string
	Acronym     string
	Owner       string
	Status      string
	SNUpdatedOn *time.Time
}
