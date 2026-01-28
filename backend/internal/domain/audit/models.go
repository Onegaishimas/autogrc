package audit

import (
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of audit event.
type EventType string

const (
	EventTypePull             EventType = "pull"
	EventTypePush             EventType = "push"
	EventTypeEdit             EventType = "edit"
	EventTypeConflictDetected EventType = "conflict_detected"
	EventTypeConflictResolved EventType = "conflict_resolved"
	EventTypeConnectionTest   EventType = "connection_test"
	EventTypeConnectionConfig EventType = "connection_config"
	EventTypeSystemImport     EventType = "system_import"
	EventTypeSystemDelete     EventType = "system_delete"
)

// Event represents an audit log entry.
type Event struct {
	ID         uuid.UUID              `json:"id"`
	EventType  EventType              `json:"event_type"`
	EntityType string                 `json:"entity_type"` // system, control, statement, connection
	EntityID   string                 `json:"entity_id"`
	Action     string                 `json:"action"`
	Status     string                 `json:"status"` // success, failure
	Details    map[string]interface{} `json:"details,omitempty"`
	UserEmail  *string                `json:"user_email,omitempty"`
	IPAddress  *string                `json:"ip_address,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// QueryFilters holds parameters for filtering audit events.
type QueryFilters struct {
	EventTypes  []EventType `json:"event_types,omitempty"`
	EntityTypes []string    `json:"entity_types,omitempty"`
	EntityID    *string     `json:"entity_id,omitempty"`
	Status      *string     `json:"status,omitempty"`
	StartDate   *time.Time  `json:"start_date,omitempty"`
	EndDate     *time.Time  `json:"end_date,omitempty"`
	Search      *string     `json:"search,omitempty"`
	Page        int         `json:"page"`
	PageSize    int         `json:"page_size"`
}

// QueryResult holds the result of querying audit events.
type QueryResult struct {
	Events     []Event `json:"events"`
	TotalCount int     `json:"total_count"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalPages int     `json:"total_pages"`
}

// Stats holds audit statistics.
type Stats struct {
	TotalEvents     int            `json:"total_events"`
	EventsByType    map[string]int `json:"events_by_type"`
	EventsByStatus  map[string]int `json:"events_by_status"`
	EventsToday     int            `json:"events_today"`
	EventsThisWeek  int            `json:"events_this_week"`
	EventsThisMonth int            `json:"events_this_month"`
}
