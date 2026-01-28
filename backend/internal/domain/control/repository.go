package control

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for control persistence operations.
type Repository interface {
	// GetByID retrieves a control by its internal ID.
	GetByID(ctx context.Context, id uuid.UUID) (*Control, error)

	// GetBySNSysID retrieves a control by its ServiceNow sys_id and system_id.
	GetBySNSysID(ctx context.Context, systemID uuid.UUID, snSysID string) (*Control, error)

	// List retrieves controls for a system with pagination.
	List(ctx context.Context, params ListParams) (*ListResult, error)

	// ListBySystem retrieves all controls for a system.
	ListBySystem(ctx context.Context, systemID uuid.UUID) ([]Control, error)

	// Upsert creates or updates a control.
	Upsert(ctx context.Context, input UpsertInput) (*Control, error)

	// UpsertBatch creates or updates multiple controls.
	UpsertBatch(ctx context.Context, inputs []UpsertInput) ([]Control, error)

	// Delete removes a control and its statements.
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteBySystem removes all controls for a system.
	DeleteBySystem(ctx context.Context, systemID uuid.UUID) error
}
