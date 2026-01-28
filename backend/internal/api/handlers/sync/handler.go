package sync

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/controlcrud/backend/internal/domain/pull"
	"github.com/controlcrud/backend/internal/domain/system"
)

// Handler handles sync-related HTTP requests.
type Handler struct {
	systemService *system.Service
	pullService   *pull.Service
	logger        *slog.Logger
}

// NewHandler creates a new sync handler.
func NewHandler(systemService *system.Service, pullService *pull.Service, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		systemService: systemService,
		pullService:   pullService,
		logger:        logger,
	}
}

// RegisterRoutes registers the sync routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// System discovery and management
	mux.HandleFunc("GET /api/v1/sync/systems/discover", h.DiscoverSystems)
	mux.HandleFunc("GET /api/v1/sync/systems", h.ListSystems)
	mux.HandleFunc("POST /api/v1/sync/systems/import", h.ImportSystems)
	mux.HandleFunc("DELETE /api/v1/sync/systems/{id}", h.DeleteSystem)

	// Pull operations
	mux.HandleFunc("POST /api/v1/sync/pull", h.StartPull)
	mux.HandleFunc("GET /api/v1/sync/pull/{id}", h.GetPullStatus)
	mux.HandleFunc("DELETE /api/v1/sync/pull/{id}", h.CancelPull)
}

// DiscoverSystems fetches systems from ServiceNow and marks imported ones.
func (h *Handler) DiscoverSystems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	discovered, err := h.systemService.DiscoverSystems(ctx)
	if err != nil {
		h.logger.Error("failed to discover systems", "error", err)
		if err == system.ErrNoConnection {
			h.writeError(w, http.StatusBadRequest, "ServiceNow connection not configured")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Failed to discover systems")
		return
	}

	// Transform to response
	response := DiscoverSystemsResponse{
		Systems: make([]DiscoveredSystemResponse, 0, len(discovered)),
		Count:   len(discovered),
	}

	for _, d := range discovered {
		response.Systems = append(response.Systems, DiscoveredSystemResponse{
			SNSysID:     d.SNSysID,
			Name:        d.Name,
			Description: d.Description,
			Owner:       d.Owner,
			IsImported:  d.IsImported,
		})
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ListSystems returns local (imported) systems with pagination.
func (h *Handler) ListSystems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query params
	params := system.ListParams{
		Page:     1,
		PageSize: 20,
		Search:   r.URL.Query().Get("search"),
		Status:   r.URL.Query().Get("status"),
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			params.Page = p
		}
	}

	if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			params.PageSize = ps
		}
	}

	result, err := h.systemService.ListSystems(ctx, params)
	if err != nil {
		h.logger.Error("failed to list systems", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to list systems")
		return
	}

	// Transform to response
	response := ListSystemsResponse{
		Systems:    make([]LocalSystemResponse, 0, len(result.Systems)),
		TotalCount: result.TotalCount,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}

	for _, s := range result.Systems {
		response.Systems = append(response.Systems, LocalSystemResponse{
			ID:             s.ID,
			SNSysID:        s.SNSysID,
			Name:           s.Name,
			Description:    s.Description,
			Acronym:        s.Acronym,
			Owner:          s.Owner,
			Status:         s.Status,
			ControlCount:   s.ControlCount,
			StatementCount: s.StatementCount,
			ModifiedCount:  s.ModifiedCount,
			LastPullAt:     s.LastPullAt,
			LastPushAt:     s.LastPushAt,
			CreatedAt:      s.CreatedAt,
			UpdatedAt:      s.UpdatedAt,
		})
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ImportSystems imports selected systems from ServiceNow.
func (h *Handler) ImportSystems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req ImportSystemsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.SNSysIDs) == 0 {
		h.writeError(w, http.StatusBadRequest, "At least one system ID is required")
		return
	}

	if len(req.SNSysIDs) > 10 {
		h.writeError(w, http.StatusBadRequest, "Maximum 10 systems can be imported at once")
		return
	}

	imported, err := h.systemService.ImportSystems(ctx, req.SNSysIDs)
	if err != nil {
		h.logger.Error("failed to import systems", "error", err)
		if err == system.ErrNoConnection {
			h.writeError(w, http.StatusBadRequest, "ServiceNow connection not configured")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Failed to import systems")
		return
	}

	// Transform to response
	response := ImportSystemsResponse{
		Imported: make([]LocalSystemResponse, 0, len(imported)),
		Count:    len(imported),
	}

	for _, s := range imported {
		response.Imported = append(response.Imported, LocalSystemResponse{
			ID:          s.ID,
			SNSysID:     s.SNSysID,
			Name:        s.Name,
			Description: s.Description,
			Acronym:     s.Acronym,
			Owner:       s.Owner,
			Status:      s.Status,
			CreatedAt:   s.CreatedAt,
			UpdatedAt:   s.UpdatedAt,
		})
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// DeleteSystem removes an imported system.
func (h *Handler) DeleteSystem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	if idStr == "" {
		h.writeError(w, http.StatusBadRequest, "System ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid system ID format")
		return
	}

	if err := h.systemService.DeleteSystem(ctx, id); err != nil {
		h.logger.Error("failed to delete system", "error", err, "id", idStr)
		if err == system.ErrNotFound {
			h.writeError(w, http.StatusNotFound, "System not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Failed to delete system")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"message": "System deleted successfully",
	})
}

// =============================================================================
// PULL OPERATIONS
// =============================================================================

// StartPull starts a new pull operation for the specified systems.
func (h *Handler) StartPull(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req StartPullRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.SystemIDs) == 0 {
		h.writeError(w, http.StatusBadRequest, "At least one system ID is required")
		return
	}

	if len(req.SystemIDs) > 10 {
		h.writeError(w, http.StatusBadRequest, "Maximum 10 systems can be pulled at once")
		return
	}

	job, err := h.pullService.StartPull(ctx, req.SystemIDs)
	if err != nil {
		h.logger.Error("failed to start pull", "error", err)
		switch err {
		case pull.ErrNoConnection:
			h.writeError(w, http.StatusBadRequest, "ServiceNow connection not configured")
		case pull.ErrConcurrentJob:
			h.writeError(w, http.StatusConflict, "Another pull operation is already in progress")
		case pull.ErrInvalidInput:
			h.writeError(w, http.StatusBadRequest, "Invalid system IDs")
		default:
			h.writeError(w, http.StatusInternalServerError, "Failed to start pull operation")
		}
		return
	}

	h.writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"job": h.transformJob(job),
	})
}

// GetPullStatus returns the current status of a pull job.
func (h *Handler) GetPullStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	if idStr == "" {
		h.writeError(w, http.StatusBadRequest, "Job ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid job ID format")
		return
	}

	job, err := h.pullService.GetJob(ctx, id)
	if err != nil {
		h.logger.Error("failed to get pull job", "error", err, "id", idStr)
		if err == pull.ErrNotFound {
			h.writeError(w, http.StatusNotFound, "Pull job not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Failed to get pull job")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"job": h.transformJob(job),
	})
}

// CancelPull cancels an active pull job.
func (h *Handler) CancelPull(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	if idStr == "" {
		h.writeError(w, http.StatusBadRequest, "Job ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid job ID format")
		return
	}

	if err := h.pullService.CancelJob(ctx, id); err != nil {
		h.logger.Error("failed to cancel pull job", "error", err, "id", idStr)
		switch err {
		case pull.ErrNotFound:
			h.writeError(w, http.StatusNotFound, "Pull job not found")
		case pull.ErrJobAlreadyComplete:
			h.writeError(w, http.StatusConflict, "Job has already completed")
		default:
			h.writeError(w, http.StatusInternalServerError, "Failed to cancel pull job")
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"message": "Pull job cancelled",
	})
}

// transformJob converts a pull.Job to PullJobResponse.
func (h *Handler) transformJob(job *pull.Job) PullJobResponse {
	return PullJobResponse{
		ID:        job.ID,
		SystemIDs: job.SystemIDs,
		Status:    string(job.Status),
		Progress: PullProgressResponse{
			TotalSystems:        job.Progress.TotalSystems,
			CompletedSystems:    job.Progress.CompletedSystems,
			TotalControls:       job.Progress.TotalControls,
			CompletedControls:   job.Progress.CompletedControls,
			TotalStatements:     job.Progress.TotalStatements,
			CompletedStatements: job.Progress.CompletedStatements,
			CurrentSystem:       job.Progress.CurrentSystem,
			Errors:              job.Progress.Errors,
		},
		StartedAt:   job.StartedAt,
		CompletedAt: job.CompletedAt,
		Error:       job.Error,
		CreatedAt:   job.CreatedAt,
	}
}

// Helper methods

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}
