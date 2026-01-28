package push

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/controlcrud/backend/internal/domain/push"
)

// Handler handles HTTP requests for push operations.
type Handler struct {
	service *push.Service
	logger  *slog.Logger
}

// NewHandler creates a new push handler.
func NewHandler(service *push.Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers push routes with the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/push", h.StartPush)
	mux.HandleFunc("GET /api/v1/push/{id}", h.GetPushStatus)
	mux.HandleFunc("DELETE /api/v1/push/{id}", h.CancelPush)
}

// StartPush handles POST /api/v1/push
func (h *Handler) StartPush(w http.ResponseWriter, r *http.Request) {
	var req StartPushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if len(req.StatementIDs) == 0 {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "At least one statement ID is required")
		return
	}

	job, err := h.service.StartPush(r.Context(), push.StartRequest{
		StatementIDs: req.StatementIDs,
	})
	if err != nil {
		switch {
		case errors.Is(err, push.ErrNoConnection):
			h.writeError(w, http.StatusBadRequest, "no_connection", "No ServiceNow connection configured")
		case errors.Is(err, push.ErrStatementNotModified):
			h.writeError(w, http.StatusBadRequest, "not_modified", err.Error())
		case errors.Is(err, push.ErrStatementHasConflict):
			h.writeError(w, http.StatusBadRequest, "has_conflict", err.Error())
		default:
			h.logger.Error("failed to start push", "error", err)
			h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to start push")
		}
		return
	}

	h.writeJSON(w, http.StatusAccepted, StartPushResponse{
		Job: h.toJobResponse(job),
	})
}

// GetPushStatus handles GET /api/v1/push/{id}
func (h *Handler) GetPushStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	jobID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "Invalid job ID format")
		return
	}

	job, err := h.service.GetJob(r.Context(), jobID)
	if err != nil {
		if errors.Is(err, push.ErrJobNotFound) {
			h.writeError(w, http.StatusNotFound, "not_found", "Push job not found")
			return
		}
		h.logger.Error("failed to get push job", "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get push job")
		return
	}

	h.writeJSON(w, http.StatusOK, PushStatusResponse{
		Job: h.toJobResponse(job),
	})
}

// CancelPush handles DELETE /api/v1/push/{id}
func (h *Handler) CancelPush(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	jobID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "Invalid job ID format")
		return
	}

	err = h.service.CancelJob(r.Context(), jobID)
	if err != nil {
		if errors.Is(err, push.ErrJobNotFound) {
			h.writeError(w, http.StatusNotFound, "not_found", "Push job not found")
			return
		}
		h.logger.Error("failed to cancel push job", "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to cancel push job")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// toJobResponse converts a domain Job to API response.
func (h *Handler) toJobResponse(job *push.Job) JobResponse {
	results := make([]StatementResultResp, len(job.Results))
	for i, r := range job.Results {
		results[i] = StatementResultResp{
			StatementID: r.StatementID,
			Success:     r.Success,
			Error:       r.Error,
			PushedAt:    r.PushedAt,
		}
	}

	return JobResponse{
		ID:          job.ID,
		Status:      string(job.Status),
		TotalCount:  job.TotalCount,
		Completed:   job.Completed,
		Succeeded:   job.Succeeded,
		Failed:      job.Failed,
		Results:     results,
		StartedAt:   job.StartedAt,
		CompletedAt: job.CompletedAt,
		CreatedAt:   job.CreatedAt,
	}
}

// writeJSON writes a JSON response.
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response.
func (h *Handler) writeError(w http.ResponseWriter, status int, code, message string) {
	h.writeJSON(w, status, ErrorResponse{
		Error:   code,
		Message: message,
	})
}
