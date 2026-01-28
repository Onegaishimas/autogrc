package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/controlcrud/backend/internal/domain/pull"
)

// PullRepository implements pull.Repository using PostgreSQL.
type PullRepository struct {
	db *sql.DB
}

// NewPullRepository creates a new pull repository.
func NewPullRepository(db *sql.DB) *PullRepository {
	return &PullRepository{db: db}
}

// Create creates a new pull job.
func (r *PullRepository) Create(ctx context.Context, input pull.CreateInput) (*pull.Job, error) {
	progress := pull.Progress{
		TotalSystems:    len(input.SystemIDs),
		Errors:          make([]string, 0),
	}
	progressJSON, err := json.Marshal(progress)
	if err != nil {
		return nil, err
	}

	job := &pull.Job{
		ID:        uuid.New(),
		SystemIDs: input.SystemIDs,
		Status:    pull.JobStatusPending,
		Progress:  progress,
		CreatedAt: time.Now(),
		CreatedBy: input.CreatedBy,
	}

	query := `
		INSERT INTO pull_jobs (id, system_ids, status, progress, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = r.db.ExecContext(ctx, query,
		job.ID,
		pq.Array(job.SystemIDs),
		job.Status,
		progressJSON,
		job.CreatedAt,
		job.CreatedBy,
	)
	if err != nil {
		return nil, err
	}

	return job, nil
}

// GetByID retrieves a pull job by ID.
func (r *PullRepository) GetByID(ctx context.Context, id uuid.UUID) (*pull.Job, error) {
	query := `
		SELECT id, system_ids, status, progress, error_message,
		       started_at, completed_at, created_at, created_by
		FROM pull_jobs
		WHERE id = $1
	`

	var job pull.Job
	var systemIDs pq.StringArray
	var progressJSON []byte
	var errorMessage sql.NullString
	var startedAt, completedAt sql.NullTime
	var createdBy *uuid.UUID

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&systemIDs,
		&job.Status,
		&progressJSON,
		&errorMessage,
		&startedAt,
		&completedAt,
		&job.CreatedAt,
		&createdBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Convert string array to UUID array
	job.SystemIDs = make([]uuid.UUID, 0, len(systemIDs))
	for _, s := range systemIDs {
		if id, err := uuid.Parse(s); err == nil {
			job.SystemIDs = append(job.SystemIDs, id)
		}
	}

	// Parse progress
	if err := json.Unmarshal(progressJSON, &job.Progress); err != nil {
		return nil, err
	}

	// Handle nullable fields
	if errorMessage.Valid {
		job.Error = errorMessage.String
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}
	job.CreatedBy = createdBy

	return &job, nil
}

// Update updates a pull job's status and progress.
func (r *PullRepository) Update(ctx context.Context, input pull.UpdateInput) (*pull.Job, error) {
	progressJSON, err := json.Marshal(input.Progress)
	if err != nil {
		return nil, err
	}

	query := `
		UPDATE pull_jobs
		SET status = $2, progress = $3, error_message = NULLIF($4, '')
		WHERE id = $1
	`

	_, err = r.db.ExecContext(ctx, query, input.ID, input.Status, progressJSON, input.Error)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, input.ID)
}

// UpdateProgress updates just the progress of a running job.
func (r *PullRepository) UpdateProgress(ctx context.Context, id uuid.UUID, progress pull.Progress) error {
	progressJSON, err := json.Marshal(progress)
	if err != nil {
		return err
	}

	query := `
		UPDATE pull_jobs
		SET progress = $2
		WHERE id = $1
	`

	_, err = r.db.ExecContext(ctx, query, id, progressJSON)
	return err
}

// SetStatus sets the job status with optional error message.
func (r *PullRepository) SetStatus(ctx context.Context, id uuid.UUID, status pull.JobStatus, errorMsg string) error {
	var query string
	var args []interface{}

	switch status {
	case pull.JobStatusRunning:
		query = `
			UPDATE pull_jobs
			SET status = $2, started_at = NOW()
			WHERE id = $1
		`
		args = []interface{}{id, status}
	case pull.JobStatusCompleted, pull.JobStatusFailed, pull.JobStatusCancelled:
		query = `
			UPDATE pull_jobs
			SET status = $2, completed_at = NOW(), error_message = NULLIF($3, '')
			WHERE id = $1
		`
		args = []interface{}{id, status, errorMsg}
	default:
		query = `
			UPDATE pull_jobs
			SET status = $2
			WHERE id = $1
		`
		args = []interface{}{id, status}
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// HasActiveJob returns true if there's an active (pending/running) job.
func (r *PullRepository) HasActiveJob(ctx context.Context) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM pull_jobs
			WHERE status IN ('pending', 'running')
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query).Scan(&exists)
	return exists, err
}

// List retrieves pull jobs with optional status filter.
func (r *PullRepository) List(ctx context.Context, status *pull.JobStatus, limit int) ([]pull.Job, error) {
	var query string
	var args []interface{}

	if status != nil {
		query = `
			SELECT id, system_ids, status, progress, error_message,
			       started_at, completed_at, created_at, created_by
			FROM pull_jobs
			WHERE status = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		args = []interface{}{*status, limit}
	} else {
		query = `
			SELECT id, system_ids, status, progress, error_message,
			       started_at, completed_at, created_at, created_by
			FROM pull_jobs
			ORDER BY created_at DESC
			LIMIT $1
		`
		args = []interface{}{limit}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []pull.Job
	for rows.Next() {
		var job pull.Job
		var systemIDs pq.StringArray
		var progressJSON []byte
		var errorMessage sql.NullString
		var startedAt, completedAt sql.NullTime
		var createdBy *uuid.UUID

		err := rows.Scan(
			&job.ID,
			&systemIDs,
			&job.Status,
			&progressJSON,
			&errorMessage,
			&startedAt,
			&completedAt,
			&job.CreatedAt,
			&createdBy,
		)
		if err != nil {
			return nil, err
		}

		// Convert string array to UUID array
		job.SystemIDs = make([]uuid.UUID, 0, len(systemIDs))
		for _, s := range systemIDs {
			if id, err := uuid.Parse(s); err == nil {
				job.SystemIDs = append(job.SystemIDs, id)
			}
		}

		// Parse progress
		if err := json.Unmarshal(progressJSON, &job.Progress); err != nil {
			return nil, err
		}

		// Handle nullable fields
		if errorMessage.Valid {
			job.Error = errorMessage.String
		}
		if startedAt.Valid {
			job.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			job.CompletedAt = &completedAt.Time
		}
		job.CreatedBy = createdBy

		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}
