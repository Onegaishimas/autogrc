package control

import "errors"

// Domain errors for control operations.
var (
	ErrNotFound        = errors.New("control not found")
	ErrInvalidInput    = errors.New("invalid input")
	ErrSystemNotFound  = errors.New("system not found")
)
