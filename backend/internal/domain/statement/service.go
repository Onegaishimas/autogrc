package statement

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

// Service provides business logic for statement operations.
type Service struct {
	repo   Repository
	logger *slog.Logger
}

// NewService creates a new statement service.
func NewService(repo Repository, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// GetByID retrieves a statement by its ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Statement, error) {
	stmt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if stmt == nil {
		return nil, ErrNotFound
	}
	return stmt, nil
}

// ListByControl retrieves statements for a control with pagination.
func (s *Service) ListByControl(ctx context.Context, params ListParams) (*ListResult, error) {
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

// ListModified retrieves all statements with local modifications.
func (s *Service) ListModified(ctx context.Context) ([]Statement, error) {
	return s.repo.ListModified(ctx)
}

// ListConflicts retrieves all statements with sync conflicts.
func (s *Service) ListConflicts(ctx context.Context) ([]Statement, error) {
	return s.repo.ListConflicts(ctx)
}

// UpdateLocal updates the local content of a statement.
func (s *Service) UpdateLocal(ctx context.Context, input UpdateInput) (*Statement, error) {
	// Verify statement exists
	existing, err := s.repo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}

	s.logger.Info("updating statement", "id", input.ID, "has_content", input.LocalContent != "")
	return s.repo.UpdateLocal(ctx, input)
}

// ResolveConflict resolves a sync conflict.
func (s *Service) ResolveConflict(ctx context.Context, input ResolveConflictInput) (*Statement, error) {
	// Verify statement exists and has a conflict
	existing, err := s.repo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	if existing.SyncStatus != SyncStatusConflict {
		return nil, fmt.Errorf("%w: statement does not have a conflict", ErrInvalidInput)
	}

	// Validate merge content for merge resolution
	if input.Resolution == ConflictResolutionMerge && input.MergedContent == "" {
		return nil, fmt.Errorf("%w: merged content is required for merge resolution", ErrInvalidInput)
	}

	s.logger.Info("resolving conflict", "id", input.ID, "resolution", input.Resolution)
	return s.repo.ResolveConflict(ctx, input)
}

// MarkAsSynced marks a statement as synced after push.
func (s *Service) MarkAsSynced(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNotFound
	}

	return s.repo.MarkAsSynced(ctx, id)
}

// RevertToRemote discards local changes and reverts to remote content.
func (s *Service) RevertToRemote(ctx context.Context, id uuid.UUID) (*Statement, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}

	if !existing.IsModified {
		return existing, nil // Already synced
	}

	// Use the resolve conflict mechanism to keep remote
	return s.repo.ResolveConflict(ctx, ResolveConflictInput{
		ID:         id,
		Resolution: ConflictResolutionKeepRemote,
	})
}
