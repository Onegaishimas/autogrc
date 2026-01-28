package pull

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/controlcrud/backend/internal/domain/control"
	"github.com/controlcrud/backend/internal/domain/statement"
	"github.com/controlcrud/backend/internal/domain/system"
	"github.com/controlcrud/backend/internal/infrastructure/servicenow"
)

// SNClientProvider provides a ServiceNow client dynamically.
type SNClientProvider interface {
	GetSNClient(ctx context.Context) (servicenow.Client, error)
}

// Service provides business logic for pull operations.
type Service struct {
	pullRepo     Repository
	systemRepo   system.Repository
	controlRepo  control.Repository
	stmtRepo     statement.Repository
	snClientGetter SNClientProvider
	logger       *slog.Logger

	// Active job tracking for cancellation
	mu           sync.RWMutex
	cancelFuncs  map[uuid.UUID]context.CancelFunc
}

// NewService creates a new pull service.
func NewService(
	pullRepo Repository,
	systemRepo system.Repository,
	controlRepo control.Repository,
	stmtRepo statement.Repository,
	snClientGetter SNClientProvider,
	logger *slog.Logger,
) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		pullRepo:       pullRepo,
		systemRepo:     systemRepo,
		controlRepo:    controlRepo,
		stmtRepo:       stmtRepo,
		snClientGetter: snClientGetter,
		logger:         logger,
		cancelFuncs:    make(map[uuid.UUID]context.CancelFunc),
	}
}

// StartPull creates a new pull job and starts execution asynchronously.
func (s *Service) StartPull(ctx context.Context, systemIDs []uuid.UUID) (*Job, error) {
	if len(systemIDs) == 0 {
		return nil, ErrInvalidInput
	}

	// Check for existing active jobs
	hasActive, err := s.pullRepo.HasActiveJob(ctx)
	if err != nil {
		return nil, err
	}
	if hasActive {
		return nil, ErrConcurrentJob
	}

	// Verify all systems exist
	for _, id := range systemIDs {
		sys, err := s.systemRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if sys == nil {
			return nil, fmt.Errorf("%w: system %s not found", ErrInvalidInput, id)
		}
	}

	// Create the job
	job, err := s.pullRepo.Create(ctx, CreateInput{
		SystemIDs: systemIDs,
	})
	if err != nil {
		return nil, err
	}

	s.logger.Info("created pull job", "job_id", job.ID, "system_count", len(systemIDs))

	// Start execution asynchronously
	go s.executePull(job.ID, systemIDs)

	return job, nil
}

// GetJob retrieves a pull job by ID.
func (s *Service) GetJob(ctx context.Context, id uuid.UUID) (*Job, error) {
	job, err := s.pullRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, ErrNotFound
	}
	return job, nil
}

// CancelJob cancels an active pull job.
func (s *Service) CancelJob(ctx context.Context, id uuid.UUID) error {
	job, err := s.pullRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if job == nil {
		return ErrNotFound
	}

	if !job.Status.IsActive() {
		return ErrJobAlreadyComplete
	}

	// Cancel the context
	s.mu.RLock()
	cancelFunc, exists := s.cancelFuncs[id]
	s.mu.RUnlock()

	if exists {
		cancelFunc()
	}

	// Update status
	return s.pullRepo.SetStatus(ctx, id, JobStatusCancelled, "cancelled by user")
}

// executePull runs the pull operation for the given systems.
func (s *Service) executePull(jobID uuid.UUID, systemIDs []uuid.UUID) {
	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register cancel function
	s.mu.Lock()
	s.cancelFuncs[jobID] = cancel
	s.mu.Unlock()

	// Cleanup on exit
	defer func() {
		s.mu.Lock()
		delete(s.cancelFuncs, jobID)
		s.mu.Unlock()
	}()

	// Get ServiceNow client
	snClient, err := s.snClientGetter.GetSNClient(ctx)
	if err != nil {
		s.logger.Error("failed to get ServiceNow client", "job_id", jobID, "error", err)
		s.pullRepo.SetStatus(ctx, jobID, JobStatusFailed, "ServiceNow connection not available")
		return
	}

	// Initialize progress
	progress := Progress{
		TotalSystems: len(systemIDs),
		Errors:       make([]string, 0),
	}

	// Update status to running
	if err := s.pullRepo.SetStatus(ctx, jobID, JobStatusRunning, ""); err != nil {
		s.logger.Error("failed to set job status", "job_id", jobID, "error", err)
		return
	}

	// Process each system
	for _, systemID := range systemIDs {
		// Check for cancellation
		select {
		case <-ctx.Done():
			s.logger.Info("pull job cancelled", "job_id", jobID)
			return
		default:
		}

		// Get system details
		sys, err := s.systemRepo.GetByID(ctx, systemID)
		if err != nil || sys == nil {
			progress.Errors = append(progress.Errors, fmt.Sprintf("system %s not found", systemID))
			continue
		}

		progress.CurrentSystem = sys.Name
		s.updateProgress(ctx, jobID, progress)

		// Pull controls and statements for this system
		if err := s.pullSystemData(ctx, snClient, sys, &progress); err != nil {
			s.logger.Error("failed to pull system data", "system", sys.Name, "error", err)
			progress.Errors = append(progress.Errors, fmt.Sprintf("%s: %v", sys.Name, err))
		}

		// Update system's last pull timestamp
		if err := s.systemRepo.UpdateLastPullAt(ctx, systemID); err != nil {
			s.logger.Warn("failed to update last_pull_at", "system_id", systemID, "error", err)
		}

		progress.CompletedSystems++
		progress.CurrentSystem = ""
		s.updateProgress(ctx, jobID, progress)
	}

	// Final status
	if ctx.Err() != nil {
		// Was cancelled
		return
	}

	status := JobStatusCompleted
	errorMsg := ""
	if len(progress.Errors) > 0 && progress.CompletedSystems == 0 {
		status = JobStatusFailed
		errorMsg = "all systems failed"
	}

	s.pullRepo.SetStatus(ctx, jobID, status, errorMsg)
	s.logger.Info("pull job completed",
		"job_id", jobID,
		"systems", progress.CompletedSystems,
		"controls", progress.CompletedControls,
		"statements", progress.CompletedStatements,
		"errors", len(progress.Errors),
	)
}

// pullSystemData fetches controls and statements for a single system.
func (s *Service) pullSystemData(
	ctx context.Context,
	snClient servicenow.Client,
	sys *system.System,
	progress *Progress,
) error {
	// Fetch controls from ServiceNow
	controlResult, err := snClient.FetchControls(ctx, sys.SNSysID, nil, nil)
	if err != nil {
		return fmt.Errorf("fetch controls: %w", err)
	}

	progress.TotalControls += len(controlResult.Records)

	// Process each control
	for _, snControl := range controlResult.Records {
		// Check cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Parse timestamps
		var snUpdatedOn *time.Time
		if snControl.SysUpdatedOn != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", snControl.SysUpdatedOn); err == nil {
				snUpdatedOn = &t
			}
		}

		// Upsert control
		ctrl, err := s.controlRepo.Upsert(ctx, control.UpsertInput{
			SystemID:             sys.ID,
			SNSysID:              snControl.SysID,
			ControlID:            snControl.ControlID,
			ControlName:          snControl.Name,
			ControlFamily:        snControl.ControlFamily,
			Description:          snControl.Description,
			ImplementationStatus: snControl.ImplementationStatus,
			SNUpdatedOn:          snUpdatedOn,
		})
		if err != nil {
			progress.Errors = append(progress.Errors, fmt.Sprintf("control %s: %v", snControl.ControlID, err))
			continue
		}

		progress.CompletedControls++

		// Fetch statements for this control
		stmtResult, err := snClient.FetchStatements(ctx, snControl.SysID, nil, nil)
		if err != nil {
			progress.Errors = append(progress.Errors, fmt.Sprintf("statements for %s: %v", snControl.ControlID, err))
			continue
		}

		progress.TotalStatements += len(stmtResult.Records)

		// Process each statement
		for _, snStmt := range stmtResult.Records {
			var stmtUpdatedOn *time.Time
			if snStmt.SysUpdatedOn != "" {
				if t, err := time.Parse("2006-01-02 15:04:05", snStmt.SysUpdatedOn); err == nil {
					stmtUpdatedOn = &t
				}
			}

			_, err := s.stmtRepo.Upsert(ctx, statement.UpsertInput{
				ControlID:     ctrl.ID,
				SNSysID:       snStmt.SysID,
				StatementType: snStmt.StatementType,
				RemoteContent: snStmt.Content,
				SNUpdatedOn:   stmtUpdatedOn,
			})
			if err != nil {
				progress.Errors = append(progress.Errors, fmt.Sprintf("statement %s: %v", snStmt.Number, err))
				continue
			}

			progress.CompletedStatements++
		}
	}

	return nil
}

// updateProgress updates the job progress in the database.
func (s *Service) updateProgress(ctx context.Context, jobID uuid.UUID, progress Progress) {
	if err := s.pullRepo.UpdateProgress(ctx, jobID, progress); err != nil {
		s.logger.Warn("failed to update progress", "job_id", jobID, "error", err)
	}
}
