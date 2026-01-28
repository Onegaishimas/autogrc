package controls

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/controlcrud/backend/internal/domain/connection"
	"github.com/controlcrud/backend/internal/infrastructure/servicenow"
)

// Service provides business logic for controls management.
type Service struct {
	connService *connection.Service
}

// NewService creates a new controls service.
func NewService(connService *connection.Service) *Service {
	return &Service{
		connService: connService,
	}
}

// ListPolicyStatements fetches policy statements from ServiceNow.
func (s *Service) ListPolicyStatements(ctx context.Context, params *ListParams) (*ListResult, error) {
	// Normalize parameters
	if params == nil {
		params = DefaultListParams()
	}
	params.Normalize()

	// Get ServiceNow client
	snClient, err := s.connService.GetSNClient(ctx)
	if err != nil {
		if errors.Is(err, connection.ErrConnectionNotFound) {
			return nil, ErrNoConnection
		}
		return nil, fmt.Errorf("%w: %v", ErrServiceNowError, err)
	}

	// Build ServiceNow query parameters
	snParams := &servicenow.PolicyStatementParams{
		Limit:    params.PageSize,
		Offset:   params.Offset(),
		Query:    params.Search,
		OrderBy:  params.SortBy,
		OrderDir: params.SortDir,
	}

	// Fetch from ServiceNow
	response, err := snClient.GetPolicyStatements(ctx, snParams)
	if err != nil {
		if errors.Is(err, servicenow.ErrAuthFailed) {
			return nil, ErrAuthFailed
		}
		if errors.Is(err, servicenow.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrServiceNowError, err)
	}

	// Transform to domain models
	items := make([]PolicyStatement, len(response.Records))
	for i, record := range response.Records {
		items[i] = transformPolicyStatement(record)
	}

	// Calculate pagination
	totalPages := response.TotalCount / params.PageSize
	if response.TotalCount%params.PageSize > 0 {
		totalPages++
	}

	return &ListResult{
		Items:      items,
		TotalCount: response.TotalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetPolicyStatement fetches a single policy statement by ID.
func (s *Service) GetPolicyStatement(ctx context.Context, id string) (*PolicyStatement, error) {
	// Get ServiceNow client
	snClient, err := s.connService.GetSNClient(ctx)
	if err != nil {
		if errors.Is(err, connection.ErrConnectionNotFound) {
			return nil, ErrNoConnection
		}
		return nil, fmt.Errorf("%w: %v", ErrServiceNowError, err)
	}

	// Fetch from ServiceNow
	record, err := snClient.GetPolicyStatement(ctx, id)
	if err != nil {
		if errors.Is(err, servicenow.ErrAuthFailed) {
			return nil, ErrAuthFailed
		}
		if errors.Is(err, servicenow.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrServiceNowError, err)
	}

	result := transformPolicyStatement(*record)
	return &result, nil
}

// transformPolicyStatement converts a ServiceNow record to domain model.
//
// =============================================================================
// DEMO MODE FALLBACKS
// =============================================================================
// The following fallbacks are needed because we're using the incident table
// instead of the IRM sn_compliance_policy_statement table:
//
// 1. Name Fallback: Incidents don't have a "name" field, so we use short_description
// 2. ControlFamily Fallback: Incidents use "priority" instead of "u_control_family"
//
// TO SWITCH TO IRM: Remove the fallback logic below - IRM records have proper values
// See: 0xcc/docs/INCIDENT_TO_IRM_MIGRATION.md for complete migration guide
// =============================================================================
func transformPolicyStatement(record servicenow.PolicyStatementRecord) PolicyStatement {
	// ==========================================================================
	// DEMO FALLBACK #1: Name
	// Incidents don't have "name" field - use short_description instead
	// IRM: Remove this fallback - Name will be populated
	// ==========================================================================
	name := record.Name
	if name == "" {
		name = record.ShortDescription // DEMO ONLY: Remove for IRM
	}

	// ==========================================================================
	// DEMO FALLBACK #2: ControlFamily
	// Incidents use "priority" (1-5) instead of control family (AC, AU, etc.)
	// IRM: Remove this fallback - ControlFamily will be populated
	// ==========================================================================
	controlFamily := record.ControlFamily
	if controlFamily == "" && record.Priority != "" {
		controlFamily = "Priority " + record.Priority // DEMO ONLY: Remove for IRM
	}

	ps := PolicyStatement{
		ID:               record.SysID,
		Number:           record.Number,
		Name:             name,
		ShortDescription: record.ShortDescription,
		Description:      record.Description,
		State:            record.State,
		Category:         record.Category,
		ControlFamily:    controlFamily,
		Active:           record.Active == "true" || record.Active == "1",
	}

	// Parse timestamps
	if record.SysCreatedOn != "" {
		if t, err := parseServiceNowTime(record.SysCreatedOn); err == nil {
			ps.CreatedAt = t
		}
	}
	if record.SysUpdatedOn != "" {
		if t, err := parseServiceNowTime(record.SysUpdatedOn); err == nil {
			ps.UpdatedAt = t
		}
	}

	return ps
}

// parseServiceNowTime parses a ServiceNow timestamp string.
func parseServiceNowTime(s string) (time.Time, error) {
	// ServiceNow uses format: 2024-01-15 10:30:00
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}
