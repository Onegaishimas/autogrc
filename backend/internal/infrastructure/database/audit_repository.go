package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/controlcrud/backend/internal/domain/audit"
)

// AuditRepository implements the audit.Repository interface.
type AuditRepository struct {
	db *sql.DB
}

// NewAuditRepository creates a new audit repository.
func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Insert creates a new audit event.
func (r *AuditRepository) Insert(ctx context.Context, event *audit.Event) error {
	detailsJSON, err := json.Marshal(event.Details)
	if err != nil {
		detailsJSON = []byte("{}")
	}

	query := `
		INSERT INTO audit_events (id, event_type, entity_type, entity_id, action, status, details, user_email, ip_address, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.ExecContext(ctx, query,
		event.ID,
		event.EventType,
		event.EntityType,
		event.EntityID,
		event.Action,
		event.Status,
		detailsJSON,
		event.UserEmail,
		event.IPAddress,
		event.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert audit event: %w", err)
	}

	return nil
}

// GetByID retrieves an audit event by ID.
func (r *AuditRepository) GetByID(ctx context.Context, id uuid.UUID) (*audit.Event, error) {
	query := `
		SELECT id, event_type, entity_type, entity_id, action, status, details, user_email, ip_address, created_at
		FROM audit_events
		WHERE id = $1
	`

	var event audit.Event
	var detailsJSON []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.EventType,
		&event.EntityType,
		&event.EntityID,
		&event.Action,
		&event.Status,
		&detailsJSON,
		&event.UserEmail,
		&event.IPAddress,
		&event.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("audit event not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get audit event: %w", err)
	}

	if len(detailsJSON) > 0 {
		json.Unmarshal(detailsJSON, &event.Details)
	}

	return &event, nil
}

// Query retrieves audit events based on filters.
func (r *AuditRepository) Query(ctx context.Context, filters audit.QueryFilters) (*audit.QueryResult, error) {
	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argNum := 1

	if len(filters.EventTypes) > 0 {
		placeholders := make([]string, len(filters.EventTypes))
		for i, et := range filters.EventTypes {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, string(et))
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("event_type IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filters.EntityTypes) > 0 {
		placeholders := make([]string, len(filters.EntityTypes))
		for i, et := range filters.EntityTypes {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, et)
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("entity_type IN (%s)", strings.Join(placeholders, ",")))
	}

	if filters.EntityID != nil && *filters.EntityID != "" {
		conditions = append(conditions, fmt.Sprintf("entity_id = $%d", argNum))
		args = append(args, *filters.EntityID)
		argNum++
	}

	if filters.Status != nil && *filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argNum))
		args = append(args, *filters.Status)
		argNum++
	}

	if filters.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argNum))
		args = append(args, *filters.StartDate)
		argNum++
	}

	if filters.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argNum))
		args = append(args, *filters.EndDate)
		argNum++
	}

	if filters.Search != nil && *filters.Search != "" {
		searchPattern := "%" + *filters.Search + "%"
		conditions = append(conditions, fmt.Sprintf("(user_email ILIKE $%d OR action ILIKE $%d OR entity_id ILIKE $%d)", argNum, argNum+1, argNum+2))
		args = append(args, searchPattern, searchPattern, searchPattern)
		argNum += 3
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_events %s", whereClause)
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count audit events: %w", err)
	}

	// Calculate pagination
	offset := (filters.Page - 1) * filters.PageSize
	totalPages := (totalCount + filters.PageSize - 1) / filters.PageSize

	// Query events
	query := fmt.Sprintf(`
		SELECT id, event_type, entity_type, entity_id, action, status, details, user_email, ip_address, created_at
		FROM audit_events
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)

	args = append(args, filters.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit events: %w", err)
	}
	defer rows.Close()

	var events []audit.Event
	for rows.Next() {
		var event audit.Event
		var detailsJSON []byte
		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.EntityType,
			&event.EntityID,
			&event.Action,
			&event.Status,
			&detailsJSON,
			&event.UserEmail,
			&event.IPAddress,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit event: %w", err)
		}

		if len(detailsJSON) > 0 {
			json.Unmarshal(detailsJSON, &event.Details)
		}

		events = append(events, event)
	}

	return &audit.QueryResult{
		Events:     events,
		TotalCount: totalCount,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetStats retrieves audit statistics.
func (r *AuditRepository) GetStats(ctx context.Context) (*audit.Stats, error) {
	stats := &audit.Stats{
		EventsByType:   make(map[string]int),
		EventsByStatus: make(map[string]int),
	}

	// Total events
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM audit_events").Scan(&stats.TotalEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to count total events: %w", err)
	}

	// Events by type
	rows, err := r.db.QueryContext(ctx, "SELECT event_type, COUNT(*) FROM audit_events GROUP BY event_type")
	if err != nil {
		return nil, fmt.Errorf("failed to count events by type: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var eventType string
		var count int
		rows.Scan(&eventType, &count)
		stats.EventsByType[eventType] = count
	}

	// Events by status
	rows, err = r.db.QueryContext(ctx, "SELECT status, COUNT(*) FROM audit_events GROUP BY status")
	if err != nil {
		return nil, fmt.Errorf("failed to count events by status: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		rows.Scan(&status, &count)
		stats.EventsByStatus[status] = count
	}

	// Events today
	today := time.Now().Truncate(24 * time.Hour)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM audit_events WHERE created_at >= $1", today).Scan(&stats.EventsToday)

	// Events this week
	weekAgo := time.Now().AddDate(0, 0, -7)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM audit_events WHERE created_at >= $1", weekAgo).Scan(&stats.EventsThisWeek)

	// Events this month
	monthAgo := time.Now().AddDate(0, -1, 0)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM audit_events WHERE created_at >= $1", monthAgo).Scan(&stats.EventsThisMonth)

	return stats, nil
}
