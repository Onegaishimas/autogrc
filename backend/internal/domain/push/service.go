package push

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/controlcrud/backend/internal/domain/connection"
	"github.com/controlcrud/backend/internal/domain/statement"
)

// Service provides business logic for push operations.
type Service struct {
	stmtRepo    statement.Repository
	connService *connection.Service
	logger      *slog.Logger

	// In-memory job storage (could be replaced with database)
	jobs   map[uuid.UUID]*Job
	jobsMu sync.RWMutex
}

// NewService creates a new push service.
func NewService(
	stmtRepo statement.Repository,
	connService *connection.Service,
	logger *slog.Logger,
) *Service {
	return &Service{
		stmtRepo:    stmtRepo,
		connService: connService,
		logger:      logger,
		jobs:        make(map[uuid.UUID]*Job),
	}
}

// StartPush starts a new push job for the specified statements.
func (s *Service) StartPush(ctx context.Context, req StartRequest) (*Job, error) {
	if len(req.StatementIDs) == 0 {
		return nil, ErrNoStatementsSelected
	}

	// Verify we have a ServiceNow connection
	_, err := s.connService.GetSNClient(ctx)
	if err != nil {
		if err == connection.ErrConnectionNotFound {
			return nil, ErrNoConnection
		}
		return nil, fmt.Errorf("failed to get ServiceNow client: %w", err)
	}

	// Verify all statements exist and are modified
	for _, stmtID := range req.StatementIDs {
		stmt, err := s.stmtRepo.GetByID(ctx, stmtID)
		if err != nil {
			return nil, fmt.Errorf("statement %s not found: %w", stmtID, err)
		}
		if !stmt.IsModified {
			return nil, fmt.Errorf("statement %s: %w", stmtID, ErrStatementNotModified)
		}
		if stmt.SyncStatus == statement.SyncStatusConflict {
			return nil, fmt.Errorf("statement %s: %w", stmtID, ErrStatementHasConflict)
		}
	}

	// Create the job
	now := time.Now()
	job := &Job{
		ID:           uuid.New(),
		Status:       JobStatusPending,
		StatementIDs: req.StatementIDs,
		Results:      []StatementResult{},
		TotalCount:   len(req.StatementIDs),
		Completed:    0,
		Succeeded:    0,
		Failed:       0,
		StartedAt:    &now,
		CreatedAt:    now,
	}

	// Store job
	s.jobsMu.Lock()
	s.jobs[job.ID] = job
	s.jobsMu.Unlock()

	// Execute push in background
	go s.executePush(job)

	return job, nil
}

// GetJob retrieves a push job by ID.
func (s *Service) GetJob(ctx context.Context, jobID uuid.UUID) (*Job, error) {
	s.jobsMu.RLock()
	job, exists := s.jobs[jobID]
	s.jobsMu.RUnlock()

	if !exists {
		return nil, ErrJobNotFound
	}

	return job, nil
}

// CancelJob cancels a running push job.
func (s *Service) CancelJob(ctx context.Context, jobID uuid.UUID) error {
	s.jobsMu.Lock()
	job, exists := s.jobs[jobID]
	if exists && IsPushJobActive(job.Status) {
		job.Status = JobStatusCancelled
		now := time.Now()
		job.CompletedAt = &now
	}
	s.jobsMu.Unlock()

	if !exists {
		return ErrJobNotFound
	}

	return nil
}

// executePush runs the push job asynchronously.
func (s *Service) executePush(job *Job) {
	ctx := context.Background()

	// Update job status to running
	s.jobsMu.Lock()
	job.Status = JobStatusRunning
	s.jobsMu.Unlock()

	// Get ServiceNow client
	snClient, err := s.connService.GetSNClient(ctx)
	if err != nil {
		s.jobsMu.Lock()
		job.Status = JobStatusFailed
		now := time.Now()
		job.CompletedAt = &now
		s.jobsMu.Unlock()
		s.logger.Error("failed to get ServiceNow client for push job",
			"job_id", job.ID,
			"error", err)
		return
	}

	// Process each statement
	for _, stmtID := range job.StatementIDs {
		// Check if job was cancelled
		s.jobsMu.RLock()
		cancelled := job.Status == JobStatusCancelled
		s.jobsMu.RUnlock()
		if cancelled {
			s.logger.Info("push job cancelled", "job_id", job.ID)
			return
		}

		result := s.pushStatement(ctx, snClient, stmtID)

		// Update job with result
		s.jobsMu.Lock()
		job.Results = append(job.Results, result)
		job.Completed++
		if result.Success {
			job.Succeeded++
		} else {
			job.Failed++
		}
		s.jobsMu.Unlock()
	}

	// Mark job as completed
	s.jobsMu.Lock()
	if job.Status != JobStatusCancelled {
		if job.Failed > 0 && job.Succeeded == 0 {
			job.Status = JobStatusFailed
		} else {
			job.Status = JobStatusCompleted
		}
	}
	now := time.Now()
	job.CompletedAt = &now
	s.jobsMu.Unlock()

	s.logger.Info("push job completed",
		"job_id", job.ID,
		"total", job.TotalCount,
		"succeeded", job.Succeeded,
		"failed", job.Failed)
}

// pushStatement pushes a single statement to ServiceNow.
func (s *Service) pushStatement(ctx context.Context, snClient interface {
	UpdateStatement(ctx context.Context, sysID string, content string) error
}, stmtID uuid.UUID) StatementResult {
	// Get the statement
	stmt, err := s.stmtRepo.GetByID(ctx, stmtID)
	if err != nil {
		errMsg := fmt.Sprintf("failed to get statement: %v", err)
		return StatementResult{
			StatementID: stmtID,
			Success:     false,
			Error:       &errMsg,
		}
	}

	// Get content to push
	content := stmt.GetContent()
	if content == "" {
		errMsg := "statement has no content to push"
		return StatementResult{
			StatementID: stmtID,
			Success:     false,
			Error:       &errMsg,
		}
	}

	// Push to ServiceNow
	err = snClient.UpdateStatement(ctx, stmt.SNSysID, content)
	if err != nil {
		errMsg := fmt.Sprintf("failed to push to ServiceNow: %v", err)
		s.logger.Error("push statement failed",
			"statement_id", stmtID,
			"sn_sys_id", stmt.SNSysID,
			"error", err)
		return StatementResult{
			StatementID: stmtID,
			Success:     false,
			Error:       &errMsg,
		}
	}

	// Mark statement as synced
	err = s.stmtRepo.MarkAsSynced(ctx, stmtID)
	if err != nil {
		s.logger.Error("failed to mark statement as synced",
			"statement_id", stmtID,
			"error", err)
		// Don't fail the push result - the push succeeded
	}

	now := time.Now()
	return StatementResult{
		StatementID: stmtID,
		Success:     true,
		PushedAt:    &now,
	}
}
