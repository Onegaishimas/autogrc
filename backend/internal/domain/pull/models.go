package pull

import (
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the current state of a pull job.
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// IsActive returns true if the job is still in progress.
func (s JobStatus) IsActive() bool {
	return s == JobStatusPending || s == JobStatusRunning
}

// Progress tracks the progress of a pull operation.
type Progress struct {
	TotalSystems       int      `json:"total_systems"`
	CompletedSystems   int      `json:"completed_systems"`
	TotalControls      int      `json:"total_controls"`
	CompletedControls  int      `json:"completed_controls"`
	TotalStatements    int      `json:"total_statements"`
	CompletedStatements int     `json:"completed_statements"`
	CurrentSystem      string   `json:"current_system,omitempty"`
	Errors             []string `json:"errors,omitempty"`
}

// Job represents a background pull operation.
type Job struct {
	ID          uuid.UUID  `json:"id"`
	SystemIDs   []uuid.UUID `json:"system_ids"`
	Status      JobStatus  `json:"status"`
	Progress    Progress   `json:"progress"`
	Error       string     `json:"error,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
}

// CreateInput holds data for creating a new pull job.
type CreateInput struct {
	SystemIDs []uuid.UUID
	CreatedBy *uuid.UUID
}

// UpdateInput holds data for updating job status and progress.
type UpdateInput struct {
	ID       uuid.UUID
	Status   JobStatus
	Progress *Progress
	Error    string
}

// CalculateOverallProgress returns the completion percentage.
func (p *Progress) CalculateOverallProgress() int {
	total := p.TotalSystems + p.TotalControls + p.TotalStatements
	completed := p.CompletedSystems + p.CompletedControls + p.CompletedStatements

	if total == 0 {
		return 0
	}
	return int(float64(completed) / float64(total) * 100)
}
