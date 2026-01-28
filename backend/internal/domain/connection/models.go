// Package connection provides domain logic for ServiceNow connection management.
package connection

import (
	"time"

	"github.com/google/uuid"
)

// AuthMethod represents the authentication method for ServiceNow.
type AuthMethod string

const (
	AuthMethodBasic AuthMethod = "basic"
	AuthMethodOAuth AuthMethod = "oauth"
)

// ConnectionStatus represents the status of the last connection test.
type ConnectionStatus string

const (
	StatusSuccess ConnectionStatus = "success"
	StatusFailure ConnectionStatus = "failure"
	StatusPending ConnectionStatus = "pending"
	StatusUnknown ConnectionStatus = "unknown"
)

// Connection represents a ServiceNow connection configuration.
type Connection struct {
	ID          uuid.UUID        `json:"id"`
	InstanceURL string           `json:"instance_url"`
	AuthMethod  AuthMethod       `json:"auth_method"`

	// Basic Auth credentials (encrypted in storage)
	Username          string `json:"username,omitempty"`
	PasswordEncrypted []byte `json:"-"`
	PasswordNonce     []byte `json:"-"`

	// OAuth credentials (encrypted in storage)
	OAuthClientID              string `json:"oauth_client_id,omitempty"`
	OAuthClientSecretEncrypted []byte `json:"-"`
	OAuthClientSecretNonce     []byte `json:"-"`
	OAuthTokenURL              string `json:"oauth_token_url,omitempty"`

	// Status tracking
	IsActive               bool             `json:"is_active"`
	LastTestAt             *time.Time       `json:"last_test_at,omitempty"`
	LastTestStatus         ConnectionStatus `json:"last_test_status"`
	LastTestMessage        string           `json:"last_test_message,omitempty"`
	LastTestInstanceVersion string           `json:"last_test_instance_version,omitempty"`

	// Audit fields
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty"`
}

// ConfigInput represents input for creating or updating a connection.
type ConfigInput struct {
	InstanceURL string     `json:"instance_url" validate:"required,url"`
	AuthMethod  AuthMethod `json:"auth_method" validate:"required,oneof=basic oauth"`

	// Basic Auth
	Username string `json:"username,omitempty" validate:"required_if=AuthMethod basic"`
	Password string `json:"password,omitempty" validate:"required_if=AuthMethod basic"`

	// OAuth
	OAuthClientID     string `json:"oauth_client_id,omitempty" validate:"required_if=AuthMethod oauth"`
	OAuthClientSecret string `json:"oauth_client_secret,omitempty" validate:"required_if=AuthMethod oauth"`
	OAuthTokenURL     string `json:"oauth_token_url,omitempty" validate:"required_if=AuthMethod oauth,omitempty,url"`
}

// ConnectionStatus represents the current connection status for display.
type Status struct {
	IsConfigured           bool             `json:"is_configured"`
	InstanceURL            string           `json:"instance_url,omitempty"`
	AuthMethod             AuthMethod       `json:"auth_method,omitempty"`
	LastTestAt             *time.Time       `json:"last_test_at,omitempty"`
	LastTestStatus         ConnectionStatus `json:"last_test_status"`
	LastTestMessage        string           `json:"last_test_message,omitempty"`
	LastTestInstanceVersion string           `json:"last_test_instance_version,omitempty"`
}

// TestResult represents the result of a connection test.
type TestResult struct {
	Success         bool      `json:"success"`
	InstanceVersion string    `json:"instance_version,omitempty"`
	BuildTag        string    `json:"build_tag,omitempty"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	ResponseTimeMs  int64     `json:"response_time_ms"`
	TestedAt        time.Time `json:"tested_at"`
}

// Validate validates the ConfigInput.
func (c *ConfigInput) Validate() error {
	if c.InstanceURL == "" {
		return ErrInstanceURLRequired
	}
	if c.AuthMethod == "" {
		return ErrAuthMethodRequired
	}
	if c.AuthMethod != AuthMethodBasic && c.AuthMethod != AuthMethodOAuth {
		return ErrInvalidAuthMethod
	}

	switch c.AuthMethod {
	case AuthMethodBasic:
		if c.Username == "" {
			return ErrUsernameRequired
		}
		if c.Password == "" {
			return ErrPasswordRequired
		}
	case AuthMethodOAuth:
		if c.OAuthClientID == "" {
			return ErrClientIDRequired
		}
		if c.OAuthClientSecret == "" {
			return ErrClientSecretRequired
		}
		if c.OAuthTokenURL == "" {
			return ErrTokenURLRequired
		}
	}

	return nil
}
