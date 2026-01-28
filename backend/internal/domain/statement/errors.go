package statement

import "errors"

// Domain errors for statement operations.
var (
	ErrNotFound       = errors.New("statement not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrControlNotFound = errors.New("control not found")
	ErrConflict       = errors.New("sync conflict detected")
)
