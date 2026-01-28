package servicenow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Common errors for ServiceNow operations.
var (
	ErrAuthFailed       = errors.New("authentication failed")
	ErrNotFound         = errors.New("resource not found")
	ErrTimeout          = errors.New("request timed out")
	ErrRateLimited      = errors.New("rate limit exceeded")
	ErrServerError      = errors.New("server error")
	ErrInvalidResponse  = errors.New("invalid response from ServiceNow")
	ErrConnectionFailed = errors.New("connection failed")
)

// Client defines the interface for ServiceNow API operations.
type Client interface {
	// TestConnection tests the connection to ServiceNow and returns instance info.
	TestConnection(ctx context.Context) (*TestConnectionResult, error)

	// SetAuth sets the authentication provider for requests.
	SetAuth(auth AuthProvider)

	// GetPolicyStatements fetches policy statements from ServiceNow GRC.
	GetPolicyStatements(ctx context.Context, params *PolicyStatementParams) (*PolicyStatementResponse, error)

	// GetPolicyStatement fetches a single policy statement by sys_id.
	GetPolicyStatement(ctx context.Context, sysID string) (*PolicyStatementRecord, error)

	// FetchSystems fetches systems/applications from ServiceNow.
	FetchSystems(ctx context.Context, config *PaginationConfig, onProgress ProgressCallback) (*PaginatedResult[SystemRecord], error)

	// FetchControls fetches controls for a system from ServiceNow.
	FetchControls(ctx context.Context, systemSysID string, config *PaginationConfig, onProgress ProgressCallback) (*PaginatedResult[ControlRecord], error)

	// FetchStatements fetches implementation statements for a control from ServiceNow.
	FetchStatements(ctx context.Context, controlSysID string, config *PaginationConfig, onProgress ProgressCallback) (*PaginatedResult[StatementRecord], error)

	// UpdateStatement updates a statement in ServiceNow.
	// In DEMO mode, updates the incident's short_description field.
	UpdateStatement(ctx context.Context, sysID string, content string) error
}

// AuthProvider provides authentication for ServiceNow requests.
type AuthProvider interface {
	// ApplyAuth applies authentication to an HTTP request.
	ApplyAuth(req *http.Request) error

	// Type returns the authentication type (basic, oauth).
	Type() string
}

// BasicAuthProvider provides HTTP Basic authentication.
type BasicAuthProvider struct {
	Username string
	Password string
}

// ApplyAuth applies Basic authentication to the request.
func (p *BasicAuthProvider) ApplyAuth(req *http.Request) error {
	req.SetBasicAuth(p.Username, p.Password)
	return nil
}

// Type returns "basic".
func (p *BasicAuthProvider) Type() string {
	return "basic"
}

// OAuthProvider provides OAuth 2.0 authentication.
type OAuthProvider struct {
	ClientID     string
	ClientSecret string
	TokenURL     string
	accessToken  string
	expiresAt    time.Time
}

// ApplyAuth applies OAuth Bearer token to the request.
func (p *OAuthProvider) ApplyAuth(req *http.Request) error {
	// TODO: Implement OAuth token refresh logic
	if p.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+p.accessToken)
	}
	return nil
}

// Type returns "oauth".
func (p *OAuthProvider) Type() string {
	return "oauth"
}

// ClientConfig holds configuration for the ServiceNow client.
type ClientConfig struct {
	InstanceURL string
	Timeout     time.Duration
	MaxRetries  int
}

// DefaultConfig returns default client configuration.
func DefaultConfig(instanceURL string) *ClientConfig {
	return &ClientConfig{
		InstanceURL: instanceURL,
		Timeout:     10 * time.Second,
		MaxRetries:  3,
	}
}

// SNClient implements the Client interface for ServiceNow API.
type SNClient struct {
	config     *ClientConfig
	httpClient *http.Client
	auth       AuthProvider
}

// NewSNClient creates a new ServiceNow client.
func NewSNClient(config *ClientConfig) (*SNClient, error) {
	if config == nil {
		return nil, errors.New("config is required")
	}
	if config.InstanceURL == "" {
		return nil, errors.New("instance URL is required")
	}

	// Validate URL format
	_, err := url.Parse(config.InstanceURL)
	if err != nil {
		return nil, fmt.Errorf("invalid instance URL: %w", err)
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	return &SNClient{
		config:     config,
		httpClient: httpClient,
	}, nil
}

// SetAuth sets the authentication provider.
func (c *SNClient) SetAuth(auth AuthProvider) {
	c.auth = auth
}

// TestConnection tests the connection to ServiceNow and returns instance info.
func (c *SNClient) TestConnection(ctx context.Context) (*TestConnectionResult, error) {
	startTime := time.Now()
	result := &TestConnectionResult{
		TestedAt: startTime,
	}

	// Query sys_properties table to get instance version
	// This is a lightweight call that confirms connectivity and auth
	endpoint := fmt.Sprintf("%s/api/now/table/sys_properties", c.config.InstanceURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("failed to create request: %v", err)
		return result, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Add query parameters to filter for version info
	q := req.URL.Query()
	q.Add("sysparm_query", "name=glide.product.name^ORname=glide.product.version^ORname=glide.buildtag")
	q.Add("sysparm_limit", "10")
	q.Add("sysparm_fields", "name,value")
	req.URL.RawQuery = q.Encode()

	// Apply authentication
	if c.auth != nil {
		if err := c.auth.ApplyAuth(req); err != nil {
			result.Success = false
			result.ErrorMessage = fmt.Sprintf("failed to apply auth: %v", err)
			return result, fmt.Errorf("failed to apply auth: %w", err)
		}
	}

	// Execute request with retries
	var resp *http.Response
	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		resp, lastErr = c.httpClient.Do(req)
		if lastErr == nil && resp.StatusCode < 500 {
			break
		}
		if attempt < c.config.MaxRetries {
			// Exponential backoff
			time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
		}
	}

	result.ResponseTimeMs = time.Since(startTime).Milliseconds()

	if lastErr != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("request failed: %v", lastErr)
		return result, fmt.Errorf("%w: %v", ErrConnectionFailed, lastErr)
	}
	defer resp.Body.Close()

	// Handle response status codes
	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		result.Success = false
		result.ErrorMessage = "authentication failed - check credentials"
		return result, ErrAuthFailed
	case resp.StatusCode == http.StatusForbidden:
		result.Success = false
		result.ErrorMessage = "access forbidden - check user permissions"
		return result, ErrAuthFailed
	case resp.StatusCode == http.StatusNotFound:
		result.Success = false
		result.ErrorMessage = "API endpoint not found - check instance URL"
		return result, ErrNotFound
	case resp.StatusCode == http.StatusTooManyRequests:
		result.Success = false
		result.ErrorMessage = "rate limit exceeded"
		return result, ErrRateLimited
	case resp.StatusCode >= 500:
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("server error (status %d)", resp.StatusCode)
		return result, ErrServerError
	case resp.StatusCode != http.StatusOK:
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		return result, fmt.Errorf("%w: status %d", ErrInvalidResponse, resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("failed to read response: %v", err)
		return result, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	var propsResponse TableAPIResponse[SysProperty]
	if err := json.Unmarshal(body, &propsResponse); err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("failed to parse response: %v", err)
		return result, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	// Extract instance info from properties
	info := InstanceInfo{
		InstanceURL: c.config.InstanceURL,
	}
	for _, prop := range propsResponse.Result {
		switch prop.Name {
		case "glide.product.name":
			// Usually "ServiceNow"
		case "glide.product.version":
			info.Version = prop.Value
		case "glide.buildtag":
			info.BuildTag = prop.Value
		}
	}

	result.Success = true
	result.InstanceInfo = info

	return result, nil
}

// mapHTTPError maps HTTP status codes to appropriate errors.
func mapHTTPError(statusCode int) error {
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrAuthFailed
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		return ErrTimeout
	default:
		if statusCode >= 500 {
			return ErrServerError
		}
		return fmt.Errorf("unexpected status code: %d", statusCode)
	}
}
