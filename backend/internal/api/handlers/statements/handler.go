package statements

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/controlcrud/backend/internal/domain/statement"
)

// Handler handles statement-related HTTP requests.
type Handler struct {
	stmtService *statement.Service
	logger      *slog.Logger
}

// NewHandler creates a new statement handler.
func NewHandler(stmtService *statement.Service, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		stmtService: stmtService,
		logger:      logger,
	}
}

// RegisterRoutes registers the statement routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Statement CRUD
	mux.HandleFunc("GET /api/v1/statements", h.ListStatements)
	mux.HandleFunc("GET /api/v1/statements/modified", h.ListModified)
	mux.HandleFunc("GET /api/v1/statements/conflicts", h.ListConflicts)
	mux.HandleFunc("GET /api/v1/statements/{id}", h.GetStatement)
	mux.HandleFunc("PUT /api/v1/statements/{id}", h.UpdateStatement)
	mux.HandleFunc("POST /api/v1/statements/{id}/resolve", h.ResolveConflict)
	mux.HandleFunc("POST /api/v1/statements/{id}/revert", h.RevertToRemote)
}

// ListStatements returns statements with pagination. Accepts control_id OR system_id filter.
func (h *Handler) ListStatements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := statement.ListParams{
		Page:     1,
		PageSize: 20,
		Search:   r.URL.Query().Get("search"),
	}

	// Accept either control_id or system_id
	controlIDStr := r.URL.Query().Get("control_id")
	systemIDStr := r.URL.Query().Get("system_id")

	if controlIDStr == "" && systemIDStr == "" {
		h.writeError(w, http.StatusBadRequest, "Either control_id or system_id is required")
		return
	}

	if controlIDStr != "" {
		controlID, err := uuid.Parse(controlIDStr)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid control_id format")
			return
		}
		params.ControlID = controlID
	}

	if systemIDStr != "" {
		systemID, err := uuid.Parse(systemIDStr)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "Invalid system_id format")
			return
		}
		params.SystemID = systemID
	}

	if syncStatus := r.URL.Query().Get("sync_status"); syncStatus != "" {
		params.SyncStatus = statement.SyncStatus(syncStatus)
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

	result, err := h.stmtService.ListByControl(ctx, params)
	if err != nil {
		h.logger.Error("failed to list statements", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to list statements")
		return
	}

	// Transform to response
	response := ListStatementsResponse{
		Statements: make([]StatementResponse, 0, len(result.Statements)),
		TotalCount: result.TotalCount,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}

	for _, s := range result.Statements {
		response.Statements = append(response.Statements, h.transformStatement(&s))
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetStatement returns a single statement by ID.
func (h *Handler) GetStatement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	if idStr == "" {
		h.writeError(w, http.StatusBadRequest, "Statement ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid statement ID format")
		return
	}

	stmt, err := h.stmtService.GetByID(ctx, id)
	if err != nil {
		h.logger.Error("failed to get statement", "error", err, "id", idStr)
		if err == statement.ErrNotFound {
			h.writeError(w, http.StatusNotFound, "Statement not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Failed to get statement")
		return
	}

	h.writeJSON(w, http.StatusOK, h.transformStatement(stmt))
}

// UpdateStatement updates a statement's local content.
func (h *Handler) UpdateStatement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	if idStr == "" {
		h.writeError(w, http.StatusBadRequest, "Statement ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid statement ID format")
		return
	}

	var req UpdateStatementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	stmt, err := h.stmtService.UpdateLocal(ctx, statement.UpdateInput{
		ID:           id,
		LocalContent: req.LocalContent,
	})
	if err != nil {
		h.logger.Error("failed to update statement", "error", err, "id", idStr)
		if err == statement.ErrNotFound {
			h.writeError(w, http.StatusNotFound, "Statement not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Failed to update statement")
		return
	}

	h.writeJSON(w, http.StatusOK, h.transformStatement(stmt))
}

// ListModified returns all statements with local modifications.
func (h *Handler) ListModified(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stmts, err := h.stmtService.ListModified(ctx)
	if err != nil {
		h.logger.Error("failed to list modified statements", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to list modified statements")
		return
	}

	response := ModifiedStatementsResponse{
		Statements: make([]StatementResponse, 0, len(stmts)),
		Count:      len(stmts),
	}

	for _, s := range stmts {
		response.Statements = append(response.Statements, h.transformStatement(&s))
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ListConflicts returns all statements with sync conflicts.
func (h *Handler) ListConflicts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stmts, err := h.stmtService.ListConflicts(ctx)
	if err != nil {
		h.logger.Error("failed to list conflict statements", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to list conflict statements")
		return
	}

	response := ConflictStatementsResponse{
		Statements: make([]StatementResponse, 0, len(stmts)),
		Count:      len(stmts),
	}

	for _, s := range stmts {
		response.Statements = append(response.Statements, h.transformStatement(&s))
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ResolveConflict resolves a sync conflict on a statement.
func (h *Handler) ResolveConflict(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	if idStr == "" {
		h.writeError(w, http.StatusBadRequest, "Statement ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid statement ID format")
		return
	}

	var req ResolveConflictRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Map resolution string to enum
	var resolution statement.ConflictResolution
	switch req.Resolution {
	case "keep_local":
		resolution = statement.ConflictResolutionKeepLocal
	case "keep_remote":
		resolution = statement.ConflictResolutionKeepRemote
	case "merge":
		resolution = statement.ConflictResolutionMerge
	default:
		h.writeError(w, http.StatusBadRequest, "Invalid resolution type. Use: keep_local, keep_remote, or merge")
		return
	}

	stmt, err := h.stmtService.ResolveConflict(ctx, statement.ResolveConflictInput{
		ID:            id,
		Resolution:    resolution,
		MergedContent: req.MergedContent,
	})
	if err != nil {
		h.logger.Error("failed to resolve conflict", "error", err, "id", idStr)
		if err == statement.ErrNotFound {
			h.writeError(w, http.StatusNotFound, "Statement not found")
			return
		}
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, h.transformStatement(stmt))
}

// RevertToRemote discards local changes and reverts to remote content.
func (h *Handler) RevertToRemote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	if idStr == "" {
		h.writeError(w, http.StatusBadRequest, "Statement ID is required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid statement ID format")
		return
	}

	stmt, err := h.stmtService.RevertToRemote(ctx, id)
	if err != nil {
		h.logger.Error("failed to revert statement", "error", err, "id", idStr)
		if err == statement.ErrNotFound {
			h.writeError(w, http.StatusNotFound, "Statement not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Failed to revert statement")
		return
	}

	h.writeJSON(w, http.StatusOK, h.transformStatement(stmt))
}

// Helper methods

func (h *Handler) transformStatement(s *statement.Statement) StatementResponse {
	return StatementResponse{
		ID:                 s.ID,
		ControlID:          s.ControlID,
		SNSysID:            s.SNSysID,
		StatementType:      s.StatementType,
		RemoteContent:      s.RemoteContent,
		RemoteUpdatedAt:    s.RemoteUpdatedAt,
		LocalContent:       s.LocalContent,
		IsModified:         s.IsModified,
		ModifiedAt:         s.ModifiedAt,
		SyncStatus:         string(s.SyncStatus),
		ConflictResolvedAt: s.ConflictResolvedAt,
		EffectiveContent:   s.GetContent(),
		LastPullAt:         s.LastPullAt,
		LastPushAt:         s.LastPushAt,
		CreatedAt:          s.CreatedAt,
		UpdatedAt:          s.UpdatedAt,
	}
}

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
