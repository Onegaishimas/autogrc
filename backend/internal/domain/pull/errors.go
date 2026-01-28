package pull

import "errors"

var (
	// ErrNotFound is returned when a pull job is not found.
	ErrNotFound = errors.New("pull job not found")

	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("invalid input")

	// ErrNoConnection is returned when ServiceNow connection is not configured.
	ErrNoConnection = errors.New("servicenow connection not configured")

	// ErrServiceNowError wraps ServiceNow API errors.
	ErrServiceNowError = errors.New("servicenow error")

	// ErrJobAlreadyComplete is returned when trying to update a completed job.
	ErrJobAlreadyComplete = errors.New("job already completed")

	// ErrJobCancelled is returned when a job is cancelled during execution.
	ErrJobCancelled = errors.New("job cancelled")

	// ErrConcurrentJob is returned when another pull job is already running.
	ErrConcurrentJob = errors.New("another pull job is already running")
)
