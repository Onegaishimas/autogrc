package audit

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// Service provides business logic for audit operations.
type Service struct {
	repo   Repository
	logger *slog.Logger
}

// NewService creates a new audit service.
func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// Record creates a new audit event.
func (s *Service) Record(ctx context.Context, event Event) error {
	event.ID = uuid.New()
	event.CreatedAt = time.Now()

	if err := s.repo.Insert(ctx, &event); err != nil {
		s.logger.Error("failed to record audit event",
			"event_type", event.EventType,
			"entity_type", event.EntityType,
			"error", err)
		return fmt.Errorf("failed to record audit event: %w", err)
	}

	s.logger.Debug("audit event recorded",
		"event_id", event.ID,
		"event_type", event.EventType,
		"entity_type", event.EntityType,
		"action", event.Action)

	return nil
}

// RecordAsync records an audit event without blocking.
// Errors are logged but not returned.
func (s *Service) RecordAsync(event Event) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Record(ctx, event); err != nil {
			s.logger.Error("async audit record failed", "error", err)
		}
	}()
}

// GetByID retrieves an audit event by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Event, error) {
	return s.repo.GetByID(ctx, id)
}

// Query retrieves audit events based on filters.
func (s *Service) Query(ctx context.Context, filters QueryFilters) (*QueryResult, error) {
	// Set defaults
	if filters.PageSize <= 0 {
		filters.PageSize = 50
	}
	if filters.PageSize > 100 {
		filters.PageSize = 100
	}
	if filters.Page < 1 {
		filters.Page = 1
	}

	return s.repo.Query(ctx, filters)
}

// GetStats retrieves audit statistics.
func (s *Service) GetStats(ctx context.Context) (*Stats, error) {
	return s.repo.GetStats(ctx)
}

// ExportCSV exports audit events as CSV.
func (s *Service) ExportCSV(ctx context.Context, filters QueryFilters) ([]byte, error) {
	// Remove pagination for export
	filters.Page = 1
	filters.PageSize = 10000 // Max export limit

	result, err := s.repo.Query(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Header
	header := []string{
		"Event ID", "Timestamp", "Event Type", "Entity Type", "Entity ID",
		"Action", "Status", "User Email", "Details",
	}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write header: %w", err)
	}

	// Data rows
	for _, event := range result.Events {
		detailsJSON := ""
		if len(event.Details) > 0 {
			b, _ := json.Marshal(event.Details)
			detailsJSON = string(b)
		}

		row := []string{
			event.ID.String(),
			event.CreatedAt.Format(time.RFC3339),
			string(event.EventType),
			event.EntityType,
			event.EntityID,
			event.Action,
			event.Status,
			safeString(event.UserEmail),
			detailsJSON,
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("csv write error: %w", err)
	}

	return buf.Bytes(), nil
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
