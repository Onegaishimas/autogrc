package servicenow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// =============================================================================
// DEMO MODE CONFIGURATION
// =============================================================================
// The following constants control which ServiceNow table is queried.
// Currently using 'incident' table because IRM (Integrated Risk Management)
// is not installed on the dev instance.
//
// TO SWITCH TO IRM:
// 1. Change policyStatementTable to "sn_compliance_policy_statement"
// 2. Update policyStatementFields to include IRM-specific fields (name, u_control_family)
// 3. Remove "priority" from fields (incident-specific)
// 4. Update transformPolicyStatement in domain/controls/service.go to remove fallbacks
// 5. Update frontend ControlCard.tsx state mappings for IRM states
//
// See: 0xcc/docs/INCIDENT_TO_IRM_MIGRATION.md for complete migration guide
// =============================================================================

const (
	// policyStatementTable is the ServiceNow table to query for policy statements.
	// DEMO: "incident" - Change to "sn_compliance_policy_statement" for IRM
	policyStatementTable = "incident"

	// policyStatementFieldsDemo are fields available on the incident table (demo mode)
	// DEMO: These are incident fields - IRM would use: sys_id,number,name,short_description,description,state,category,u_control_family,active,sys_created_on,sys_updated_on
)

// GetPolicyStatements fetches policy statements from ServiceNow.
// DEMO MODE: Currently using 'incident' table. See constants above to switch to IRM.
func (c *SNClient) GetPolicyStatements(ctx context.Context, params *PolicyStatementParams) (*PolicyStatementResponse, error) {
	// Build the endpoint URL using the configured table
	endpoint := fmt.Sprintf("%s/api/now/table/%s", c.config.InstanceURL, policyStatementTable)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request: %v", ErrConnectionFailed, err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Build query parameters
	q := req.URL.Query()

	// Pagination
	limit := 20
	if params != nil && params.Limit > 0 {
		limit = params.Limit
		if limit > 100 {
			limit = 100
		}
	}
	q.Set("sysparm_limit", strconv.Itoa(limit))

	offset := 0
	if params != nil && params.Offset > 0 {
		offset = params.Offset
	}
	q.Set("sysparm_offset", strconv.Itoa(offset))

	// Fields to return (using incident fields for demo)
	fields := []string{
		"sys_id", "number", "short_description", "description",
		"state", "category", "priority", "active",
		"sys_created_on", "sys_updated_on",
	}
	if params != nil && len(params.Fields) > 0 {
		fields = params.Fields
	}
	q.Set("sysparm_fields", strings.Join(fields, ","))

	// Build query string for search/filter
	var queryParts []string
	if params != nil && params.Query != "" {
		// Search by number or short_description (case-insensitive contains)
		searchQuery := fmt.Sprintf("numberLIKE%s^ORshort_descriptionLIKE%s",
			params.Query, params.Query)
		queryParts = append(queryParts, searchQuery)
	}

	// Only active records by default
	queryParts = append(queryParts, "active=true")

	if len(queryParts) > 0 {
		q.Set("sysparm_query", strings.Join(queryParts, "^"))
	}

	// Ordering
	orderBy := "number"
	if params != nil && params.OrderBy != "" {
		orderBy = params.OrderBy
	}
	orderDir := ""
	if params != nil && params.OrderDir == "desc" {
		orderDir = "DESC"
	}
	q.Set("sysparm_order_by", orderBy)
	if orderDir != "" {
		q.Set("sysparm_order_direction", orderDir)
	}

	// Request total count in response headers
	q.Set("sysparm_suppress_pagination_header", "false")

	req.URL.RawQuery = q.Encode()

	// Apply authentication
	if c.auth != nil {
		if err := c.auth.ApplyAuth(req); err != nil {
			return nil, fmt.Errorf("failed to apply auth: %w", err)
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
			time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, lastErr)
	}
	defer resp.Body.Close()

	// Handle response status codes
	if err := checkResponseStatus(resp); err != nil {
		return nil, err
	}

	// Parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read response: %v", ErrInvalidResponse, err)
	}

	var tableResponse TableAPIResponse[PolicyStatementRecord]
	if err := json.Unmarshal(body, &tableResponse); err != nil {
		return nil, fmt.Errorf("%w: failed to parse response: %v", ErrInvalidResponse, err)
	}

	// Get total count from header
	totalCount := len(tableResponse.Result)
	if totalHeader := resp.Header.Get("X-Total-Count"); totalHeader != "" {
		if count, err := strconv.Atoi(totalHeader); err == nil {
			totalCount = count
		}
	}

	return &PolicyStatementResponse{
		Records:    tableResponse.Result,
		TotalCount: totalCount,
	}, nil
}

// GetPolicyStatement fetches a single policy statement by sys_id.
// DEMO MODE: Currently using 'incident' table. See constants above to switch to IRM.
func (c *SNClient) GetPolicyStatement(ctx context.Context, sysID string) (*PolicyStatementRecord, error) {
	endpoint := fmt.Sprintf("%s/api/now/table/%s/%s", c.config.InstanceURL, policyStatementTable, sysID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request: %v", ErrConnectionFailed, err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Apply authentication
	if c.auth != nil {
		if err := c.auth.ApplyAuth(req); err != nil {
			return nil, fmt.Errorf("failed to apply auth: %w", err)
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
			time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, lastErr)
	}
	defer resp.Body.Close()

	// Handle response status codes
	if err := checkResponseStatus(resp); err != nil {
		return nil, err
	}

	// Parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read response: %v", ErrInvalidResponse, err)
	}

	// Single record response has different structure
	var singleResponse struct {
		Result PolicyStatementRecord `json:"result"`
	}
	if err := json.Unmarshal(body, &singleResponse); err != nil {
		return nil, fmt.Errorf("%w: failed to parse response: %v", ErrInvalidResponse, err)
	}

	return &singleResponse.Result, nil
}

// checkResponseStatus checks HTTP response status and returns appropriate error.
func checkResponseStatus(resp *http.Response) error {
	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return ErrAuthFailed
	case resp.StatusCode == http.StatusForbidden:
		return ErrAuthFailed
	case resp.StatusCode == http.StatusNotFound:
		return ErrNotFound
	case resp.StatusCode == http.StatusTooManyRequests:
		return ErrRateLimited
	case resp.StatusCode >= 500:
		return fmt.Errorf("%w: status %d", ErrServerError, resp.StatusCode)
	case resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated:
		return fmt.Errorf("%w: status %d", ErrInvalidResponse, resp.StatusCode)
	}
	return nil
}

// UpdateStatement updates a statement in ServiceNow.
// DEMO MODE: Updates the incident's short_description field.
// FOR IRM: Would update sn_compliance_policy_statement.u_implementation_statement
func (c *SNClient) UpdateStatement(ctx context.Context, sysID string, content string) error {
	endpoint := fmt.Sprintf("%s/api/now/table/%s/%s", c.config.InstanceURL, policyStatementTable, sysID)

	// Build the payload - using short_description for demo (incident table)
	// FOR IRM: Change to "u_implementation_statement" or appropriate field
	payload := map[string]string{
		"short_description": content,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return fmt.Errorf("%w: failed to create request: %v", ErrConnectionFailed, err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Apply authentication
	if c.auth != nil {
		if err := c.auth.ApplyAuth(req); err != nil {
			return fmt.Errorf("failed to apply auth: %w", err)
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
			time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
		}
	}

	if lastErr != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, lastErr)
	}
	defer resp.Body.Close()

	// Handle response status codes
	if err := checkResponseStatus(resp); err != nil {
		return err
	}

	return nil
}
