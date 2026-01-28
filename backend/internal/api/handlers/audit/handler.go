package audit

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/controlcrud/backend/internal/domain/audit"
)

// Handler handles HTTP requests for audit operations.
type Handler struct {
	service *audit.Service
	logger  *slog.Logger
}

// NewHandler creates a new audit handler.
func NewHandler(service *audit.Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers audit routes with the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/audit", h.QueryEvents)
	mux.HandleFunc("GET /api/v1/audit/stats", h.GetStats)
	mux.HandleFunc("GET /api/v1/audit/export", h.ExportEvents)
	mux.HandleFunc("GET /api/v1/audit/{id}", h.GetEvent)
}

// QueryEvents handles GET /api/v1/audit
func (h *Handler) QueryEvents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	filters := audit.QueryFilters{
		Page:     1,
		PageSize: 50,
	}

	// Parse event types
	if eventTypes := query.Get("event_types"); eventTypes != "" {
		for _, et := range strings.Split(eventTypes, ",") {
			filters.EventTypes = append(filters.EventTypes, audit.EventType(strings.TrimSpace(et)))
		}
	}

	// Parse entity types
	if entityTypes := query.Get("entity_types"); entityTypes != "" {
		filters.EntityTypes = strings.Split(entityTypes, ",")
	}

	// Parse entity ID
	if entityID := query.Get("entity_id"); entityID != "" {
		filters.EntityID = &entityID
	}

	// Parse status
	if status := query.Get("status"); status != "" {
		filters.Status = &status
	}

	// Parse dates
	if startDate := query.Get("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			filters.StartDate = &t
		}
	}
	if endDate := query.Get("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			filters.EndDate = &t
		}
	}

	// Parse search
	if search := query.Get("search"); search != "" {
		filters.Search = &search
	}

	// Parse pagination
	if page := query.Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filters.Page = p
		}
	}
	if pageSize := query.Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			filters.PageSize = ps
		}
	}

	result, err := h.service.Query(r.Context(), filters)
	if err != nil {
		h.logger.Error("failed to query audit events", "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to query audit events")
		return
	}

	// Convert to response
	events := make([]EventResponse, len(result.Events))
	for i, e := range result.Events {
		events[i] = EventResponse{
			ID:         e.ID,
			EventType:  string(e.EventType),
			EntityType: e.EntityType,
			EntityID:   e.EntityID,
			Action:     e.Action,
			Status:     e.Status,
			Details:    e.Details,
			UserEmail:  e.UserEmail,
			IPAddress:  e.IPAddress,
			CreatedAt:  e.CreatedAt,
		}
	}

	h.writeJSON(w, http.StatusOK, QueryEventsResponse{
		Events:     events,
		TotalCount: result.TotalCount,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	})
}

// GetEvent handles GET /api/v1/audit/{id}
func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "Invalid event ID format")
		return
	}

	event, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "not_found", "Audit event not found")
		return
	}

	h.writeJSON(w, http.StatusOK, EventResponse{
		ID:         event.ID,
		EventType:  string(event.EventType),
		EntityType: event.EntityType,
		EntityID:   event.EntityID,
		Action:     event.Action,
		Status:     event.Status,
		Details:    event.Details,
		UserEmail:  event.UserEmail,
		IPAddress:  event.IPAddress,
		CreatedAt:  event.CreatedAt,
	})
}

// GetStats handles GET /api/v1/audit/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		h.logger.Error("failed to get audit stats", "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get audit stats")
		return
	}

	h.writeJSON(w, http.StatusOK, StatsResponse{
		TotalEvents:     stats.TotalEvents,
		EventsByType:    stats.EventsByType,
		EventsByStatus:  stats.EventsByStatus,
		EventsToday:     stats.EventsToday,
		EventsThisWeek:  stats.EventsThisWeek,
		EventsThisMonth: stats.EventsThisMonth,
	})
}

// ExportEvents handles GET /api/v1/audit/export
func (h *Handler) ExportEvents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	filters := audit.QueryFilters{
		Page:     1,
		PageSize: 10000,
	}

	// Parse event types
	if eventTypes := query.Get("event_types"); eventTypes != "" {
		for _, et := range strings.Split(eventTypes, ",") {
			filters.EventTypes = append(filters.EventTypes, audit.EventType(strings.TrimSpace(et)))
		}
	}

	// Parse entity types
	if entityTypes := query.Get("entity_types"); entityTypes != "" {
		filters.EntityTypes = strings.Split(entityTypes, ",")
	}

	// Parse dates
	if startDate := query.Get("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			filters.StartDate = &t
		}
	}
	if endDate := query.Get("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			filters.EndDate = &t
		}
	}

	csvData, err := h.service.ExportCSV(r.Context(), filters)
	if err != nil {
		h.logger.Error("failed to export audit events", "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to export audit events")
		return
	}

	filename := "audit_export_" + time.Now().Format("20060102_150405") + ".csv"
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Write(csvData)
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
