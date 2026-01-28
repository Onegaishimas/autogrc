package connection

import (
	"time"

	"github.com/controlcrud/backend/internal/domain/connection"
)

// ConfigRequest represents the request body for saving connection configuration.
type ConfigRequest struct {
	InstanceURL       string `json:"instance_url" validate:"required,url"`
	AuthMethod        string `json:"auth_method" validate:"required,oneof=basic oauth"`
	Username          string `json:"username,omitempty" validate:"required_if=AuthMethod basic"`
	Password          string `json:"password,omitempty" validate:"required_if=AuthMethod basic"`
	OAuthClientID     string `json:"oauth_client_id,omitempty" validate:"required_if=AuthMethod oauth"`
	OAuthClientSecret string `json:"oauth_client_secret,omitempty" validate:"required_if=AuthMethod oauth"`
	OAuthTokenURL     string `json:"oauth_token_url,omitempty" validate:"required_if=AuthMethod oauth,omitempty,url"`
}

// ToConfigInput converts the request to domain ConfigInput.
func (r *ConfigRequest) ToConfigInput() *connection.ConfigInput {
	return &connection.ConfigInput{
		InstanceURL:       r.InstanceURL,
		AuthMethod:        connection.AuthMethod(r.AuthMethod),
		Username:          r.Username,
		Password:          r.Password,
		OAuthClientID:     r.OAuthClientID,
		OAuthClientSecret: r.OAuthClientSecret,
		OAuthTokenURL:     r.OAuthTokenURL,
	}
}

// StatusResponse represents the response for connection status.
type StatusResponse struct {
	IsConfigured    bool       `json:"is_configured"`
	InstanceURL     string     `json:"instance_url,omitempty"`
	AuthMethod      string     `json:"auth_method,omitempty"`
	LastTestAt      *time.Time `json:"last_test_at,omitempty"`
	LastTestStatus  string     `json:"last_test_status"`
	InstanceVersion string     `json:"instance_version,omitempty"`
}

// NewStatusResponse creates a StatusResponse from domain Status.
func NewStatusResponse(status *connection.Status) *StatusResponse {
	return &StatusResponse{
		IsConfigured:    status.IsConfigured,
		InstanceURL:     status.InstanceURL,
		AuthMethod:      string(status.AuthMethod),
		LastTestAt:      status.LastTestAt,
		LastTestStatus:  string(status.LastTestStatus),
		InstanceVersion: status.LastTestInstanceVersion,
	}
}

// TestResponse represents the response for connection test.
type TestResponse struct {
	Success         bool   `json:"success"`
	Message         string `json:"message,omitempty"`
	InstanceVersion string `json:"instance_version,omitempty"`
	BuildTag        string `json:"build_tag,omitempty"`
	ResponseTimeMs  int64  `json:"response_time_ms,omitempty"`
}

// NewTestResponse creates a TestResponse from domain TestResult.
func NewTestResponse(result *connection.TestResult) *TestResponse {
	return &TestResponse{
		Success:         result.Success,
		Message:         result.ErrorMessage,
		InstanceVersion: result.InstanceVersion,
		BuildTag:        result.BuildTag,
		ResponseTimeMs:  result.ResponseTimeMs,
	}
}

// ConfigResponse represents the response after saving configuration.
type ConfigResponse struct {
	ID          string `json:"id"`
	InstanceURL string `json:"instance_url"`
	AuthMethod  string `json:"auth_method"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// ValidationErrorResponse represents a validation error response.
type ValidationErrorResponse struct {
	Error   string              `json:"error"`
	Message string              `json:"message"`
	Fields  []ValidationError   `json:"fields,omitempty"`
}

// ValidationError represents a single field validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
