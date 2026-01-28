package system

import "errors"

// Domain errors for system operations.
var (
	ErrNotFound         = errors.New("system not found")
	ErrNoConnection     = errors.New("ServiceNow connection not configured")
	ErrServiceNowError  = errors.New("ServiceNow API error")
	ErrInvalidInput     = errors.New("invalid input")
)
