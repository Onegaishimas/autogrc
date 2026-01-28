package push

import "errors"

var (
	// ErrJobNotFound is returned when a push job cannot be found.
	ErrJobNotFound = errors.New("push job not found")

	// ErrJobAlreadyRunning is returned when trying to start a push while one is running.
	ErrJobAlreadyRunning = errors.New("a push job is already running")

	// ErrNoStatementsSelected is returned when no statements are selected for push.
	ErrNoStatementsSelected = errors.New("no statements selected for push")

	// ErrStatementNotModified is returned when trying to push a statement that hasn't been modified.
	ErrStatementNotModified = errors.New("statement has not been modified")

	// ErrStatementHasConflict is returned when trying to push a statement with unresolved conflict.
	ErrStatementHasConflict = errors.New("statement has unresolved conflict")

	// ErrNoConnection is returned when no ServiceNow connection is configured.
	ErrNoConnection = errors.New("no ServiceNow connection configured")

	// ErrServiceNowError is returned when ServiceNow API returns an error.
	ErrServiceNowError = errors.New("ServiceNow API error")
)
