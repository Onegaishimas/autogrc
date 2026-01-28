package audit

import (
	"time"

	"github.com/google/uuid"
)

// EventResponse represents an audit event in API responses.
type EventResponse struct {
	ID         uuid.UUID              `json:"id"`
	EventType  string                 `json:"event_type"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Action     string                 `json:"action"`
	Status     string                 `json:"status"`
	Details    map[string]interface{} `json:"details,omitempty"`
	UserEmail  *string                `json:"user_email,omitempty"`
	IPAddress  *string                `json:"ip_address,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// QueryEventsResponse is the response for listing audit events.
type QueryEventsResponse struct {
	Events     []EventResponse `json:"events"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// StatsResponse is the response for audit statistics.
type StatsResponse struct {
	TotalEvents     int            `json:"total_events"`
	EventsByType    map[string]int `json:"events_by_type"`
	EventsByStatus  map[string]int `json:"events_by_status"`
	EventsToday     int            `json:"events_today"`
	EventsThisWeek  int            `json:"events_this_week"`
	EventsThisMonth int            `json:"events_this_month"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
