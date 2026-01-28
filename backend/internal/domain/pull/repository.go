package pull

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for pull job persistence operations.
type Repository interface {
	// Create creates a new pull job.
	Create(ctx context.Context, input CreateInput) (*Job, error)

	// GetByID retrieves a pull job by ID.
	GetByID(ctx context.Context, id uuid.UUID) (*Job, error)

	// Update updates a pull job's status and progress.
	Update(ctx context.Context, input UpdateInput) (*Job, error)

	// UpdateProgress updates just the progress of a running job.
	UpdateProgress(ctx context.Context, id uuid.UUID, progress Progress) error

	// SetStatus sets the job status with optional error message.
	SetStatus(ctx context.Context, id uuid.UUID, status JobStatus, errorMsg string) error

	// HasActiveJob returns true if there's an active (pending/running) job.
	HasActiveJob(ctx context.Context) (bool, error)

	// List retrieves pull jobs with optional status filter.
	List(ctx context.Context, status *JobStatus, limit int) ([]Job, error)
}
