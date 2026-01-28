package connection

import "errors"

// Domain errors for connection operations.
var (
	// Validation errors
	ErrInstanceURLRequired  = errors.New("instance URL is required")
	ErrAuthMethodRequired   = errors.New("authentication method is required")
	ErrInvalidAuthMethod    = errors.New("invalid authentication method: must be 'basic' or 'oauth'")
	ErrUsernameRequired     = errors.New("username is required for basic authentication")
	ErrPasswordRequired     = errors.New("password is required for basic authentication")
	ErrClientIDRequired     = errors.New("client ID is required for OAuth authentication")
	ErrClientSecretRequired = errors.New("client secret is required for OAuth authentication")
	ErrTokenURLRequired     = errors.New("token URL is required for OAuth authentication")

	// Repository errors
	ErrConnectionNotFound = errors.New("connection not found")
	ErrConnectionExists   = errors.New("active connection already exists")

	// Service errors
	ErrEncryptionFailed = errors.New("failed to encrypt credentials")
	ErrDecryptionFailed = errors.New("failed to decrypt credentials")
	ErrTestFailed       = errors.New("connection test failed")
)
