package system

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/controlcrud/backend/internal/infrastructure/servicenow"
)

// SNClientProvider provides a ServiceNow client dynamically.
type SNClientProvider interface {
	GetSNClient(ctx context.Context) (servicenow.Client, error)
}

// Service provides business logic for system operations.
type Service struct {
	repo           Repository
	snClientGetter SNClientProvider
	logger         *slog.Logger
}

// NewService creates a new system service.
func NewService(repo Repository, snClientGetter SNClientProvider, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		repo:           repo,
		snClientGetter: snClientGetter,
		logger:         logger,
	}
}

// getSNClient gets the ServiceNow client from the provider.
func (s *Service) getSNClient(ctx context.Context) (servicenow.Client, error) {
	if s.snClientGetter == nil {
		return nil, ErrNoConnection
	}
	return s.snClientGetter.GetSNClient(ctx)
}

// DiscoverSystems fetches systems from ServiceNow and marks which ones are already imported.
func (s *Service) DiscoverSystems(ctx context.Context) ([]DiscoveredSystem, error) {
	snClient, err := s.getSNClient(ctx)
	if err != nil {
		return nil, err
	}

	s.logger.Info("discovering systems from ServiceNow")

	// Fetch systems from ServiceNow
	result, err := snClient.FetchSystems(ctx, nil, nil)
	if err != nil {
		s.logger.Error("failed to fetch systems from ServiceNow", "error", err)
		return nil, fmt.Errorf("%w: %v", ErrServiceNowError, err)
	}

	// Get existing system IDs
	existingSysIDs, err := s.repo.GetAllSNSysIDs(ctx)
	if err != nil {
		s.logger.Error("failed to get existing system IDs", "error", err)
		return nil, err
	}

	// Create a map for quick lookup
	existingMap := make(map[string]bool, len(existingSysIDs))
	for _, id := range existingSysIDs {
		existingMap[id] = true
	}

	// Transform to DiscoveredSystem
	discovered := make([]DiscoveredSystem, 0, len(result.Records))
	for _, record := range result.Records {
		discovered = append(discovered, DiscoveredSystem{
			SNSysID:     record.SysID,
			Name:        record.Name,
			Description: record.Description,
			Owner:       record.Owner,
			IsImported:  existingMap[record.SysID],
		})
	}

	s.logger.Info("discovered systems", "count", len(discovered), "already_imported", len(existingSysIDs))
	return discovered, nil
}

// ImportSystems imports selected systems from ServiceNow into the local database.
func (s *Service) ImportSystems(ctx context.Context, snSysIDs []string) ([]System, error) {
	snClient, err := s.getSNClient(ctx)
	if err != nil {
		return nil, err
	}

	if len(snSysIDs) == 0 {
		return nil, ErrInvalidInput
	}

	s.logger.Info("importing systems", "count", len(snSysIDs))

	// Fetch all systems from ServiceNow (we'll filter locally)
	result, err := snClient.FetchSystems(ctx, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrServiceNowError, err)
	}

	// Create map of requested IDs
	requestedIDs := make(map[string]bool, len(snSysIDs))
	for _, id := range snSysIDs {
		requestedIDs[id] = true
	}

	// Filter and prepare upsert inputs
	inputs := make([]UpsertInput, 0, len(snSysIDs))
	for _, record := range result.Records {
		if !requestedIDs[record.SysID] {
			continue
		}

		var snUpdatedOn *time.Time
		if record.SysUpdatedOn != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", record.SysUpdatedOn); err == nil {
				snUpdatedOn = &t
			}
		}

		inputs = append(inputs, UpsertInput{
			SNSysID:     record.SysID,
			Name:        record.Name,
			Description: record.Description,
			Owner:       record.Owner,
			Status:      record.Status,
			SNUpdatedOn: snUpdatedOn,
		})
	}

	if len(inputs) == 0 {
		s.logger.Warn("no matching systems found to import")
		return []System{}, nil
	}

	// Upsert systems
	systems, err := s.repo.UpsertBatch(ctx, inputs)
	if err != nil {
		s.logger.Error("failed to upsert systems", "error", err)
		return nil, err
	}

	s.logger.Info("imported systems", "count", len(systems))
	return systems, nil
}

// ListSystems retrieves local systems with pagination.
func (s *Service) ListSystems(ctx context.Context, params ListParams) (*ListResult, error) {
	// Set defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	return s.repo.List(ctx, params)
}

// GetSystem retrieves a single system by ID.
func (s *Service) GetSystem(ctx context.Context, id uuid.UUID) (*System, error) {
	system, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if system == nil {
		return nil, ErrNotFound
	}
	return system, nil
}

// DeleteSystem removes a system and all its associated data.
func (s *Service) DeleteSystem(ctx context.Context, id uuid.UUID) error {
	// Verify system exists
	system, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if system == nil {
		return ErrNotFound
	}

	s.logger.Info("deleting system", "id", id, "name", system.Name)
	return s.repo.Delete(ctx, id)
}
