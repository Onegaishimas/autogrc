package control

import (
	"time"

	"github.com/google/uuid"
)

// Control represents a NIST 800-53 control mapped to a system.
// In IRM, this maps to sn_compliance_control.
// DEMO MODE: Derived from incident priorities.
type Control struct {
	ID         uuid.UUID `json:"id"`
	SystemID   uuid.UUID `json:"system_id"`
	SNSysID    string    `json:"sn_sys_id"`
	ControlID  string    `json:"control_id"`   // e.g., "AC-1", "SC-7"
	ControlName string   `json:"control_name"`
	ControlFamily string `json:"control_family,omitempty"` // e.g., "AC", "SC"
	Description string   `json:"description,omitempty"`

	ImplementationStatus string `json:"implementation_status"`
	ResponsibleRole      string `json:"responsible_role,omitempty"`

	// Sync metadata
	SNUpdatedOn *time.Time `json:"sn_updated_on,omitempty"`
	LastPullAt  *time.Time `json:"last_pull_at,omitempty"`
	LastPushAt  *time.Time `json:"last_push_at,omitempty"`

	// Audit
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ControlWithStats includes statement counts.
type ControlWithStats struct {
	Control
	StatementCount int `json:"statement_count"`
	ModifiedCount  int `json:"modified_count"`
}

// ListParams holds parameters for listing controls.
type ListParams struct {
	SystemID      uuid.UUID `json:"system_id"`
	Page          int       `json:"page"`
	PageSize      int       `json:"page_size"`
	Search        string    `json:"search,omitempty"`
	ControlFamily string    `json:"control_family,omitempty"`
}

// ListResult holds the result of listing controls.
type ListResult struct {
	Controls   []ControlWithStats `json:"controls"`
	TotalCount int                `json:"total_count"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// UpsertInput holds data for creating or updating a control.
type UpsertInput struct {
	SystemID             uuid.UUID
	SNSysID              string
	ControlID            string
	ControlName          string
	ControlFamily        string
	Description          string
	ImplementationStatus string
	ResponsibleRole      string
	SNUpdatedOn          *time.Time
}

// NIST800_53Families maps family codes to full names.
var NIST800_53Families = map[string]string{
	"AC": "Access Control",
	"AT": "Awareness and Training",
	"AU": "Audit and Accountability",
	"CA": "Assessment, Authorization, and Monitoring",
	"CM": "Configuration Management",
	"CP": "Contingency Planning",
	"IA": "Identification and Authentication",
	"IR": "Incident Response",
	"MA": "Maintenance",
	"MP": "Media Protection",
	"PE": "Physical and Environmental Protection",
	"PL": "Planning",
	"PM": "Program Management",
	"PS": "Personnel Security",
	"PT": "PII Processing and Transparency",
	"RA": "Risk Assessment",
	"SA": "System and Services Acquisition",
	"SC": "System and Communications Protection",
	"SI": "System and Information Integrity",
	"SR": "Supply Chain Risk Management",
}

// ExtractControlFamily extracts the family prefix from a control ID.
// e.g., "AC-1" -> "AC", "SC-7(1)" -> "SC"
func ExtractControlFamily(controlID string) string {
	for i, c := range controlID {
		if c == '-' || (c >= '0' && c <= '9') {
			return controlID[:i]
		}
	}
	return controlID
}
