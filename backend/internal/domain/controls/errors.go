package controls

import "errors"

// Domain errors for controls operations.
var (
	// ErrNoConnection indicates no ServiceNow connection is configured.
	ErrNoConnection = errors.New("no ServiceNow connection configured")

	// ErrServiceNowError indicates a ServiceNow API error.
	ErrServiceNowError = errors.New("ServiceNow API error")

	// ErrNotFound indicates the requested resource was not found.
	ErrNotFound = errors.New("policy statement not found")

	// ErrAuthFailed indicates authentication with ServiceNow failed.
	ErrAuthFailed = errors.New("ServiceNow authentication failed")
)
