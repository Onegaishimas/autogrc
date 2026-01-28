package servicenow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"
)

// PaginationConfig holds configuration for paginated requests.
type PaginationConfig struct {
	PageSize       int           // Items per page (default 100)
	MaxPages       int           // Maximum pages to fetch (0 = unlimited)
	RetryDelay     time.Duration // Initial delay between retries
	MaxRetryDelay  time.Duration // Maximum delay for exponential backoff
	RateLimitDelay time.Duration // Delay when rate limited
}

// DefaultPaginationConfig returns sensible defaults.
func DefaultPaginationConfig() *PaginationConfig {
	return &PaginationConfig{
		PageSize:       100,
		MaxPages:       0, // unlimited
		RetryDelay:     500 * time.Millisecond,
		MaxRetryDelay:  30 * time.Second,
		RateLimitDelay: 60 * time.Second,
	}
}

// PaginatedResult holds the results of a paginated fetch operation.
type PaginatedResult[T any] struct {
	Records    []T
	TotalCount int
	PagesFetched int
	Errors     []error
}

// ProgressCallback is called after each page is fetched.
// Returns false to cancel the operation.
type ProgressCallback func(fetched, total int) bool

// FetchAllPages fetches all pages from a ServiceNow table endpoint.
// It handles pagination, rate limiting, and retries automatically.
func FetchAllPages[T any](
	ctx context.Context,
	client *SNClient,
	endpoint string,
	query map[string]string,
	config *PaginationConfig,
	onProgress ProgressCallback,
) (*PaginatedResult[T], error) {
	if config == nil {
		config = DefaultPaginationConfig()
	}

	result := &PaginatedResult[T]{
		Records: make([]T, 0),
	}

	offset := 0
	pageNum := 0

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		// Check max pages limit
		if config.MaxPages > 0 && pageNum >= config.MaxPages {
			break
		}

		// Build request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return result, fmt.Errorf("%w: failed to create request: %v", ErrConnectionFailed, err)
		}

		// Set headers
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		// Build query parameters
		q := req.URL.Query()
		for k, v := range query {
			q.Set(k, v)
		}
		q.Set("sysparm_offset", strconv.Itoa(offset))
		q.Set("sysparm_limit", strconv.Itoa(config.PageSize))
		req.URL.RawQuery = q.Encode()

		// Apply authentication
		if client.auth != nil {
			if err := client.auth.ApplyAuth(req); err != nil {
				return result, fmt.Errorf("failed to apply auth: %w", err)
			}
		}

		// Execute request with retries and rate limit handling
		resp, err := executeWithRetry(ctx, client, req, config)
		if err != nil {
			result.Errors = append(result.Errors, err)
			return result, err
		}
		defer resp.Body.Close()

		// Check for errors
		if err := checkResponseError(resp); err != nil {
			result.Errors = append(result.Errors, err)
			return result, err
		}

		// Parse total count from header (ServiceNow returns this)
		if totalHeader := resp.Header.Get("X-Total-Count"); totalHeader != "" {
			if total, err := strconv.Atoi(totalHeader); err == nil {
				result.TotalCount = total
			}
		}

		// Parse response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return result, fmt.Errorf("%w: failed to read response: %v", ErrInvalidResponse, err)
		}

		var tableResponse TableAPIResponse[T]
		if err := json.Unmarshal(body, &tableResponse); err != nil {
			return result, fmt.Errorf("%w: failed to parse response: %v", ErrInvalidResponse, err)
		}

		// Append records
		result.Records = append(result.Records, tableResponse.Result...)
		result.PagesFetched++
		pageNum++

		// Call progress callback
		if onProgress != nil {
			if !onProgress(len(result.Records), result.TotalCount) {
				// Cancelled by callback
				return result, nil
			}
		}

		// Check if we've fetched all records
		if len(tableResponse.Result) < config.PageSize {
			// Last page (fewer records than page size)
			break
		}
		if result.TotalCount > 0 && len(result.Records) >= result.TotalCount {
			// Fetched all records based on total count
			break
		}

		offset += config.PageSize
	}

	return result, nil
}

// executeWithRetry executes a request with exponential backoff and rate limit handling.
func executeWithRetry(
	ctx context.Context,
	client *SNClient,
	req *http.Request,
	config *PaginationConfig,
) (*http.Response, error) {
	var lastErr error
	delay := config.RetryDelay

	for attempt := 0; attempt <= client.config.MaxRetries; attempt++ {
		// Check context before each attempt
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		resp, err := client.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("%w: %v", ErrConnectionFailed, err)
			time.Sleep(delay)
			delay = minDuration(delay*2, config.MaxRetryDelay)
			continue
		}

		// Handle rate limiting
		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()

			// Check for Retry-After header
			retryAfter := config.RateLimitDelay
			if retryHeader := resp.Header.Get("Retry-After"); retryHeader != "" {
				if seconds, err := strconv.Atoi(retryHeader); err == nil {
					retryAfter = time.Duration(seconds) * time.Second
				}
			}

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryAfter):
			}
			continue
		}

		// Success or non-retryable error
		if resp.StatusCode < 500 {
			return resp, nil
		}

		// Server error - retry
		resp.Body.Close()
		lastErr = fmt.Errorf("%w: status %d", ErrServerError, resp.StatusCode)
		time.Sleep(delay)
		delay = minDuration(delay*2, config.MaxRetryDelay)
	}

	if lastErr == nil {
		lastErr = ErrConnectionFailed
	}
	return nil, lastErr
}

// checkResponseError checks the response status and returns an appropriate error.
func checkResponseError(resp *http.Response) error {
	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return ErrAuthFailed
	case resp.StatusCode == http.StatusForbidden:
		return fmt.Errorf("%w: access forbidden", ErrAuthFailed)
	case resp.StatusCode == http.StatusNotFound:
		return ErrNotFound
	case resp.StatusCode == http.StatusTooManyRequests:
		return ErrRateLimited
	case resp.StatusCode >= 500:
		return fmt.Errorf("%w: status %d", ErrServerError, resp.StatusCode)
	case resp.StatusCode != http.StatusOK:
		return fmt.Errorf("%w: status %d", ErrInvalidResponse, resp.StatusCode)
	}
	return nil
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// =============================================================================
// SYSTEM FETCH METHODS
// =============================================================================
// DEMO MODE: Currently using 'incident' table. When IRM is available,
// change to appropriate IRM tables (e.g., sn_grc_m2m_scoped_item_policy_statement)

// SystemRecord represents a system/application from ServiceNow.
// DEMO: Maps from incident caller_id reference. IRM: Maps from cmdb_ci_service or similar.
type SystemRecord struct {
	SysID       string `json:"sys_id"`
	Name        string `json:"name"`
	Description string `json:"short_description,omitempty"`
	Status      string `json:"operational_status,omitempty"`
	Owner       string `json:"owned_by,omitempty"`
	SysUpdatedOn string `json:"sys_updated_on,omitempty"`
}

// FetchSystems fetches systems/applications from ServiceNow.
// DEMO MODE: Returns distinct categories from incidents as mock systems.
func (c *SNClient) FetchSystems(ctx context.Context, config *PaginationConfig, onProgress ProgressCallback) (*PaginatedResult[SystemRecord], error) {
	// DEMO: Using incident categories as mock systems
	// IRM: Would use cmdb_ci_service or sn_grc_business_entity table
	endpoint := fmt.Sprintf("%s/api/now/table/sys_choice", c.config.InstanceURL)

	query := map[string]string{
		"sysparm_query":  "name=incident^element=category^inactive=false",
		"sysparm_fields": "sys_id,label,value,sys_updated_on",
	}

	// Fetch choices which represent our "systems" in demo mode
	choiceResult, err := FetchAllPages[map[string]interface{}](ctx, c, endpoint, query, config, onProgress)
	if err != nil {
		return nil, err
	}

	// Transform to SystemRecord
	result := &PaginatedResult[SystemRecord]{
		Records:      make([]SystemRecord, 0, len(choiceResult.Records)),
		TotalCount:   choiceResult.TotalCount,
		PagesFetched: choiceResult.PagesFetched,
		Errors:       choiceResult.Errors,
	}

	for _, choice := range choiceResult.Records {
		sysID, _ := choice["sys_id"].(string)
		label, _ := choice["label"].(string)
		value, _ := choice["value"].(string)
		updatedOn, _ := choice["sys_updated_on"].(string)

		result.Records = append(result.Records, SystemRecord{
			SysID:        sysID,
			Name:         label,
			Description:  fmt.Sprintf("Category: %s", value),
			Status:       "active",
			SysUpdatedOn: updatedOn,
		})
	}

	return result, nil
}

// ControlRecord represents a control from ServiceNow.
// DEMO: Derived from incident data. IRM: Maps from sn_compliance_control.
type ControlRecord struct {
	SysID              string `json:"sys_id"`
	ControlID          string `json:"control_id"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	ControlFamily      string `json:"control_family,omitempty"`
	ImplementationStatus string `json:"implementation_status,omitempty"`
	SysUpdatedOn       string `json:"sys_updated_on,omitempty"`
}

// FetchControls fetches controls for a system from ServiceNow.
// DEMO MODE: Returns mock controls based on incident priorities.
func (c *SNClient) FetchControls(ctx context.Context, systemSysID string, config *PaginationConfig, onProgress ProgressCallback) (*PaginatedResult[ControlRecord], error) {
	// DEMO: Using priorities as mock controls
	// IRM: Would use sn_compliance_control table with system filter
	endpoint := fmt.Sprintf("%s/api/now/table/sys_choice", c.config.InstanceURL)

	query := map[string]string{
		"sysparm_query":  "name=incident^element=priority^inactive=false",
		"sysparm_fields": "sys_id,label,value,sys_updated_on",
	}

	choiceResult, err := FetchAllPages[map[string]interface{}](ctx, c, endpoint, query, config, onProgress)
	if err != nil {
		return nil, err
	}

	// Transform to ControlRecord
	result := &PaginatedResult[ControlRecord]{
		Records:      make([]ControlRecord, 0, len(choiceResult.Records)),
		TotalCount:   choiceResult.TotalCount,
		PagesFetched: choiceResult.PagesFetched,
		Errors:       choiceResult.Errors,
	}

	// Map priority to NIST control families for demo
	familyMap := map[string]string{
		"1": "AC",  // Access Control
		"2": "AU",  // Audit
		"3": "CM",  // Configuration Management
		"4": "IA",  // Identification and Authentication
		"5": "SC",  // System and Communications Protection
	}

	for _, choice := range choiceResult.Records {
		sysID, _ := choice["sys_id"].(string)
		label, _ := choice["label"].(string)
		value, _ := choice["value"].(string)
		updatedOn, _ := choice["sys_updated_on"].(string)

		family := familyMap[value]
		if family == "" {
			family = "XX" // Unknown
		}

		result.Records = append(result.Records, ControlRecord{
			SysID:         sysID,
			ControlID:     fmt.Sprintf("%s-%s", family, value),
			Name:          label,
			Description:   fmt.Sprintf("Control derived from priority %s", value),
			ControlFamily: family,
			ImplementationStatus: "not_assessed",
			SysUpdatedOn:  updatedOn,
		})
	}

	return result, nil
}

// StatementRecord represents an implementation statement from ServiceNow.
// DEMO: Uses incident short_description. IRM: Maps from sn_compliance_policy_statement.
type StatementRecord struct {
	SysID        string `json:"sys_id"`
	Number       string `json:"number,omitempty"`
	Name         string `json:"name,omitempty"`
	Content      string `json:"content"`
	StatementType string `json:"statement_type,omitempty"`
	SysUpdatedOn string `json:"sys_updated_on,omitempty"`
}

// FetchStatements fetches implementation statements for a control from ServiceNow.
// DEMO MODE: Returns incidents as mock statements.
func (c *SNClient) FetchStatements(ctx context.Context, controlSysID string, config *PaginationConfig, onProgress ProgressCallback) (*PaginatedResult[StatementRecord], error) {
	// DEMO: Using incidents as mock statements
	// IRM: Would use sn_compliance_policy_statement table
	endpoint := fmt.Sprintf("%s/api/now/table/incident", c.config.InstanceURL)

	query := map[string]string{
		"sysparm_query":  "active=true",
		"sysparm_fields": "sys_id,number,short_description,description,sys_updated_on",
		"sysparm_limit":  strconv.Itoa(int(math.Min(float64(DefaultPaginationConfig().PageSize), 20))), // Limit for demo
	}

	incidentResult, err := FetchAllPages[map[string]interface{}](ctx, c, endpoint, query, config, onProgress)
	if err != nil {
		return nil, err
	}

	// Transform to StatementRecord
	result := &PaginatedResult[StatementRecord]{
		Records:      make([]StatementRecord, 0, len(incidentResult.Records)),
		TotalCount:   incidentResult.TotalCount,
		PagesFetched: incidentResult.PagesFetched,
		Errors:       incidentResult.Errors,
	}

	for _, incident := range incidentResult.Records {
		sysID, _ := incident["sys_id"].(string)
		number, _ := incident["number"].(string)
		shortDesc, _ := incident["short_description"].(string)
		desc, _ := incident["description"].(string)
		updatedOn, _ := incident["sys_updated_on"].(string)

		content := shortDesc
		if desc != "" {
			content = fmt.Sprintf("%s\n\n%s", shortDesc, desc)
		}

		result.Records = append(result.Records, StatementRecord{
			SysID:         sysID,
			Number:        number,
			Name:          shortDesc,
			Content:       content,
			StatementType: "implementation",
			SysUpdatedOn:  updatedOn,
		})
	}

	return result, nil
}
