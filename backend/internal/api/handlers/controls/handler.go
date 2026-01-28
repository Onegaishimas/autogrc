package controls

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/controlcrud/backend/internal/domain/controls"
)

// Handler handles HTTP requests for controls management.
type Handler struct {
	service *controls.Service
}

// NewHandler creates a new controls handler.
func NewHandler(service *controls.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers the controls routes with the provided mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/controls/policy-statements", h.ListPolicyStatements)
	mux.HandleFunc("GET /api/v1/controls/policy-statements/{id}", h.GetPolicyStatement)
}

// ListPolicyStatements handles GET /api/v1/controls/policy-statements
// Returns a paginated list of policy statements from ServiceNow.
func (h *Handler) ListPolicyStatements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	params := &controls.ListParams{
		Page:     parseIntParam(r, "page", 1),
		PageSize: parseIntParam(r, "page_size", 20),
		Search:   r.URL.Query().Get("search"),
		SortBy:   r.URL.Query().Get("sort_by"),
		SortDir:  r.URL.Query().Get("sort_dir"),
	}

	// Fetch policy statements
	result, err := h.service.ListPolicyStatements(ctx, params)
	if err != nil {
		handleError(w, err)
		return
	}

	// Return response
	writeJSON(w, http.StatusOK, NewListPolicyStatementsResponse(result))
}

// GetPolicyStatement handles GET /api/v1/controls/policy-statements/{id}
// Returns a single policy statement by ID.
func (h *Handler) GetPolicyStatement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get ID from path
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "invalid_id", "Policy statement ID is required")
		return
	}

	// Fetch policy statement
	ps, err := h.service.GetPolicyStatement(ctx, id)
	if err != nil {
		handleError(w, err)
		return
	}

	// Return response
	writeJSON(w, http.StatusOK, NewPolicyStatementDTO(ps))
}

// parseIntParam parses an integer query parameter with a default value.
func parseIntParam(r *http.Request, name string, defaultValue int) int {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// handleError maps domain errors to HTTP responses.
func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, controls.ErrNoConnection):
		writeError(w, http.StatusPreconditionFailed, "no_connection", "No ServiceNow connection configured. Please configure a connection first.")
	case errors.Is(err, controls.ErrAuthFailed):
		writeError(w, http.StatusUnauthorized, "auth_failed", "ServiceNow authentication failed. Please check your credentials.")
	case errors.Is(err, controls.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "Policy statement not found")
	case errors.Is(err, controls.ErrServiceNowError):
		writeError(w, http.StatusBadGateway, "servicenow_error", "Failed to communicate with ServiceNow")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred")
	}
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, status int, errorCode, message string) {
	writeJSON(w, status, &ErrorResponse{
		Error:   errorCode,
		Message: message,
	})
}
