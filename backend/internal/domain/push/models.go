package push

import (
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the state of a push job.
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// Job represents a push job that syncs statements to ServiceNow.
type Job struct {
	ID           uuid.UUID        `json:"id"`
	Status       JobStatus        `json:"status"`
	StatementIDs []uuid.UUID      `json:"statement_ids"`
	Results      []StatementResult `json:"results"`
	TotalCount   int              `json:"total_count"`
	Completed    int              `json:"completed"`
	Succeeded    int              `json:"succeeded"`
	Failed       int              `json:"failed"`
	StartedAt    *time.Time       `json:"started_at,omitempty"`
	CompletedAt  *time.Time       `json:"completed_at,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
}

// StatementResult represents the result of pushing a single statement.
type StatementResult struct {
	StatementID uuid.UUID  `json:"statement_id"`
	Success     bool       `json:"success"`
	Error       *string    `json:"error,omitempty"`
	PushedAt    *time.Time `json:"pushed_at,omitempty"`
}

// StartRequest contains the parameters for starting a push job.
type StartRequest struct {
	StatementIDs []uuid.UUID `json:"statement_ids"`
}

// IsPushJobActive returns true if the job is still running.
func IsPushJobActive(status JobStatus) bool {
	return status == JobStatusPending || status == JobStatusRunning
}
