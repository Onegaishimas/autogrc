package audit

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for audit persistence operations.
type Repository interface {
	// Insert creates a new audit event.
	Insert(ctx context.Context, event *Event) error

	// GetByID retrieves an audit event by ID.
	GetByID(ctx context.Context, id uuid.UUID) (*Event, error)

	// Query retrieves audit events based on filters.
	Query(ctx context.Context, filters QueryFilters) (*QueryResult, error)

	// GetStats retrieves audit statistics.
	GetStats(ctx context.Context) (*Stats, error)
}
