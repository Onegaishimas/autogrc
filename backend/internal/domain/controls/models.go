// Package controls provides domain logic for managing compliance controls.
package controls

import "time"

// PolicyStatement represents a compliance policy statement (control template).
type PolicyStatement struct {
	ID               string    `json:"id"`
	Number           string    `json:"number"`
	Name             string    `json:"name"`
	ShortDescription string    `json:"short_description"`
	Description      string    `json:"description,omitempty"`
	State            string    `json:"state"`
	Category         string    `json:"category,omitempty"`
	ControlFamily    string    `json:"control_family,omitempty"`
	Active           bool      `json:"active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ListParams represents parameters for listing policy statements.
type ListParams struct {
	Page     int    // 1-indexed page number
	PageSize int    // Items per page (default: 20, max: 100)
	Search   string // Search by name/number
	SortBy   string // Field to sort by
	SortDir  string // "asc" or "desc"
}

// ListResult represents a paginated list of policy statements.
type ListResult struct {
	Items      []PolicyStatement `json:"items"`
	TotalCount int               `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// DefaultListParams returns default list parameters.
func DefaultListParams() *ListParams {
	return &ListParams{
		Page:     1,
		PageSize: 20,
		SortBy:   "number",
		SortDir:  "asc",
	}
}

// Normalize applies defaults and constraints to list parameters.
func (p *ListParams) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	if p.SortBy == "" {
		p.SortBy = "number"
	}
	if p.SortDir != "desc" {
		p.SortDir = "asc"
	}
}

// Offset returns the offset for pagination.
func (p *ListParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}
