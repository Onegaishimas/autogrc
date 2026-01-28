package system

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for system persistence operations.
type Repository interface {
	// GetByID retrieves a system by its internal ID.
	GetByID(ctx context.Context, id uuid.UUID) (*System, error)

	// GetBySNSysID retrieves a system by its ServiceNow sys_id.
	GetBySNSysID(ctx context.Context, snSysID string) (*System, error)

	// List retrieves systems with pagination and optional filters.
	List(ctx context.Context, params ListParams) (*ListResult, error)

	// ListAll retrieves all systems without pagination.
	ListAll(ctx context.Context) ([]System, error)

	// Upsert creates or updates a system based on sn_sys_id.
	Upsert(ctx context.Context, input UpsertInput) (*System, error)

	// UpsertBatch creates or updates multiple systems.
	UpsertBatch(ctx context.Context, inputs []UpsertInput) ([]System, error)

	// Delete removes a system and all its related controls/statements.
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateLastPullAt updates the last pull timestamp.
	UpdateLastPullAt(ctx context.Context, id uuid.UUID) error

	// GetAllSNSysIDs returns all ServiceNow sys_ids for existing systems.
	GetAllSNSysIDs(ctx context.Context) ([]string, error)
}
