package statement

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for statement persistence operations.
type Repository interface {
	// GetByID retrieves a statement by its internal ID.
	GetByID(ctx context.Context, id uuid.UUID) (*Statement, error)

	// GetBySNSysID retrieves a statement by its ServiceNow sys_id and control_id.
	GetBySNSysID(ctx context.Context, controlID uuid.UUID, snSysID string) (*Statement, error)

	// List retrieves statements for a control with pagination.
	List(ctx context.Context, params ListParams) (*ListResult, error)

	// ListByControl retrieves all statements for a control.
	ListByControl(ctx context.Context, controlID uuid.UUID) ([]Statement, error)

	// ListModified retrieves all statements with local modifications.
	ListModified(ctx context.Context) ([]Statement, error)

	// ListConflicts retrieves all statements with sync conflicts.
	ListConflicts(ctx context.Context) ([]Statement, error)

	// Upsert creates or updates a statement from ServiceNow.
	// Preserves local modifications and detects conflicts.
	Upsert(ctx context.Context, input UpsertInput) (*Statement, error)

	// UpsertBatch creates or updates multiple statements.
	UpsertBatch(ctx context.Context, inputs []UpsertInput) ([]Statement, error)

	// UpdateLocal updates the local content of a statement.
	UpdateLocal(ctx context.Context, input UpdateInput) (*Statement, error)

	// ResolveConflict resolves a sync conflict.
	ResolveConflict(ctx context.Context, input ResolveConflictInput) (*Statement, error)

	// Delete removes a statement.
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByControl removes all statements for a control.
	DeleteByControl(ctx context.Context, controlID uuid.UUID) error

	// MarkAsSynced marks a statement as synced after push.
	MarkAsSynced(ctx context.Context, id uuid.UUID) error
}
