// Package servicenow provides a client for interacting with ServiceNow GRC APIs.
package servicenow

import "time"

// InstanceInfo contains information about a ServiceNow instance.
type InstanceInfo struct {
	Version     string `json:"version"`
	BuildTag    string `json:"build_tag"`
	BuildDate   string `json:"build_date"`
	InstanceID  string `json:"instance_id"`
	InstanceURL string `json:"instance_url"`
}

// TableAPIResponse represents a generic ServiceNow Table API response.
type TableAPIResponse[T any] struct {
	Result []T `json:"result"`
}

// SysProperty represents a ServiceNow system property record.
type SysProperty struct {
	SysID       string `json:"sys_id"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

// ComplianceStatement represents a ServiceNow GRC compliance statement.
type ComplianceStatement struct {
	SysID           string    `json:"sys_id"`
	Number          string    `json:"number"`
	ShortDesc       string    `json:"short_description"`
	Description     string    `json:"description"`
	Statement       string    `json:"statement"`
	ControlID       string    `json:"control"`
	PolicyID        string    `json:"policy"`
	Status          string    `json:"status"`
	SysCreatedOn    time.Time `json:"-"`
	SysCreatedOnStr string    `json:"sys_created_on"`
	SysUpdatedOn    time.Time `json:"-"`
	SysUpdatedOnStr string    `json:"sys_updated_on"`
	SysUpdatedBy    string    `json:"sys_updated_by"`
}

// ComplianceControl represents a ServiceNow GRC compliance control.
type ComplianceControl struct {
	SysID           string    `json:"sys_id"`
	Number          string    `json:"number"`
	Name            string    `json:"name"`
	ShortDesc       string    `json:"short_description"`
	Description     string    `json:"description"`
	ControlFamily   string    `json:"control_family"`
	FrameworkID     string    `json:"framework"`
	Status          string    `json:"status"`
	SysCreatedOn    time.Time `json:"-"`
	SysCreatedOnStr string    `json:"sys_created_on"`
	SysUpdatedOn    time.Time `json:"-"`
	SysUpdatedOnStr string    `json:"sys_updated_on"`
}

// InformationSystem represents a ServiceNow GRC information system record.
type InformationSystem struct {
	SysID            string    `json:"sys_id"`
	Number           string    `json:"number"`
	Name             string    `json:"name"`
	ShortDesc        string    `json:"short_description"`
	Description      string    `json:"description"`
	SystemOwner      string    `json:"system_owner"`
	Status           string    `json:"status"`
	SecurityCategory string    `json:"security_category"`
	SysCreatedOn     time.Time `json:"-"`
	SysCreatedOnStr  string    `json:"sys_created_on"`
	SysUpdatedOn     time.Time `json:"-"`
	SysUpdatedOnStr  string    `json:"sys_updated_on"`
}

// APIError represents a ServiceNow API error response.
type APIError struct {
	Error struct {
		Message string `json:"message"`
		Detail  string `json:"detail"`
	} `json:"error"`
	Status string `json:"status"`
}

// TestConnectionResult contains the result of a connection test.
type TestConnectionResult struct {
	Success         bool         `json:"success"`
	InstanceInfo    InstanceInfo `json:"instance_info,omitempty"`
	ErrorMessage    string       `json:"error_message,omitempty"`
	ResponseTimeMs  int64        `json:"response_time_ms"`
	TestedAt        time.Time    `json:"tested_at"`
}

// PolicyStatementRecord represents a ServiceNow IRM policy statement record.
//
// =============================================================================
// DEMO MODE: Currently mapped to incident table fields
// =============================================================================
// When switching to IRM (sn_compliance_policy_statement table):
// - Name will be populated (remove fallback to ShortDescription in service.go)
// - ControlFamily will have real values (remove Priority fallback in service.go)
// - Priority field can be removed (incident-specific)
// - State values will be strings like "draft", "active" instead of numbers
//
// See: 0xcc/docs/INCIDENT_TO_IRM_MIGRATION.md for complete migration guide
// =============================================================================
type PolicyStatementRecord struct {
	SysID            string `json:"sys_id"`
	Number           string `json:"number"`
	Name             string `json:"name"`              // IRM: populated | DEMO: empty (use ShortDescription)
	ShortDescription string `json:"short_description"` // Both: populated
	Description      string `json:"description"`       // Both: populated
	State            string `json:"state"`             // IRM: "draft","active" | DEMO: "1","2","3"
	Category         string `json:"category"`          // Both: populated (different values)
	ControlFamily    string `json:"u_control_family"`  // IRM: real value | DEMO: empty
	Priority         string `json:"priority"`          // DEMO ONLY: used as ControlFamily fallback (remove for IRM)
	Active           string `json:"active"`            // Both: "true"/"false" or "1"/"0"
	SysCreatedOn     string `json:"sys_created_on"`    // Both: timestamp
	SysUpdatedOn     string `json:"sys_updated_on"`    // Both: timestamp
}

// PolicyStatementParams contains parameters for fetching policy statements.
type PolicyStatementParams struct {
	Limit    int
	Offset   int
	Query    string
	Fields   []string
	OrderBy  string
	OrderDir string
}

// PolicyStatementResponse contains the response from fetching policy statements.
type PolicyStatementResponse struct {
	Records    []PolicyStatementRecord
	TotalCount int
}
