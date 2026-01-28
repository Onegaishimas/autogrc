package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/controlcrud/backend/internal/domain/system"
)

// SystemRepository implements system.Repository using PostgreSQL.
type SystemRepository struct {
	db *sql.DB
}

// NewSystemRepository creates a new system repository.
func NewSystemRepository(db *sql.DB) *SystemRepository {
	return &SystemRepository{db: db}
}

// GetByID retrieves a system by its internal ID.
func (r *SystemRepository) GetByID(ctx context.Context, id uuid.UUID) (*system.System, error) {
	query := `
		SELECT id, sn_sys_id, name, description, acronym, owner, status,
		       sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
		FROM systems
		WHERE id = $1
	`

	var s system.System
	var description, acronym, owner sql.NullString
	var snUpdatedOn, lastPullAt, lastPushAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.SNSysID, &s.Name, &description, &acronym, &owner, &s.Status,
		&snUpdatedOn, &lastPullAt, &lastPushAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get system: %w", err)
	}

	s.Description = description.String
	s.Acronym = acronym.String
	s.Owner = owner.String
	if snUpdatedOn.Valid {
		s.SNUpdatedOn = &snUpdatedOn.Time
	}
	if lastPullAt.Valid {
		s.LastPullAt = &lastPullAt.Time
	}
	if lastPushAt.Valid {
		s.LastPushAt = &lastPushAt.Time
	}

	return &s, nil
}

// GetBySNSysID retrieves a system by its ServiceNow sys_id.
func (r *SystemRepository) GetBySNSysID(ctx context.Context, snSysID string) (*system.System, error) {
	query := `
		SELECT id, sn_sys_id, name, description, acronym, owner, status,
		       sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
		FROM systems
		WHERE sn_sys_id = $1
	`

	var s system.System
	var description, acronym, owner sql.NullString
	var snUpdatedOn, lastPullAt, lastPushAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, snSysID).Scan(
		&s.ID, &s.SNSysID, &s.Name, &description, &acronym, &owner, &s.Status,
		&snUpdatedOn, &lastPullAt, &lastPushAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get system: %w", err)
	}

	s.Description = description.String
	s.Acronym = acronym.String
	s.Owner = owner.String
	if snUpdatedOn.Valid {
		s.SNUpdatedOn = &snUpdatedOn.Time
	}
	if lastPullAt.Valid {
		s.LastPullAt = &lastPullAt.Time
	}
	if lastPushAt.Valid {
		s.LastPushAt = &lastPushAt.Time
	}

	return &s, nil
}

// List retrieves systems with pagination and optional filters.
func (r *SystemRepository) List(ctx context.Context, params system.ListParams) (*system.ListResult, error) {
	// Build query with filters
	var conditions []string
	var args []interface{}
	argNum := 1

	if params.Status != "" {
		conditions = append(conditions, fmt.Sprintf("s.status = $%d", argNum))
		args = append(args, params.Status)
		argNum++
	}

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(s.name ILIKE $%d OR s.description ILIKE $%d OR s.acronym ILIKE $%d)", argNum, argNum, argNum))
		args = append(args, "%"+params.Search+"%")
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM systems s %s`, whereClause)
	var totalCount int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("failed to count systems: %w", err)
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.PageSize
	totalPages := (totalCount + params.PageSize - 1) / params.PageSize

	// Fetch systems with stats
	query := fmt.Sprintf(`
		SELECT s.id, s.sn_sys_id, s.name, s.description, s.acronym, s.owner, s.status,
		       s.sn_updated_on, s.last_pull_at, s.last_push_at, s.created_at, s.updated_at,
		       COALESCE((SELECT COUNT(*) FROM controls c WHERE c.system_id = s.id), 0) as control_count,
		       COALESCE((SELECT COUNT(*) FROM statements st
		                 JOIN controls c ON st.control_id = c.id
		                 WHERE c.system_id = s.id), 0) as statement_count,
		       COALESCE((SELECT COUNT(*) FROM statements st
		                 JOIN controls c ON st.control_id = c.id
		                 WHERE c.system_id = s.id AND st.is_modified = true), 0) as modified_count
		FROM systems s
		%s
		ORDER BY s.name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)

	args = append(args, params.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list systems: %w", err)
	}
	defer rows.Close()

	systems := make([]system.SystemWithStats, 0)
	for rows.Next() {
		var s system.SystemWithStats
		var description, acronym, owner sql.NullString
		var snUpdatedOn, lastPullAt, lastPushAt sql.NullTime

		err := rows.Scan(
			&s.ID, &s.SNSysID, &s.Name, &description, &acronym, &owner, &s.Status,
			&snUpdatedOn, &lastPullAt, &lastPushAt, &s.CreatedAt, &s.UpdatedAt,
			&s.ControlCount, &s.StatementCount, &s.ModifiedCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan system: %w", err)
		}

		s.Description = description.String
		s.Acronym = acronym.String
		s.Owner = owner.String
		if snUpdatedOn.Valid {
			s.SNUpdatedOn = &snUpdatedOn.Time
		}
		if lastPullAt.Valid {
			s.LastPullAt = &lastPullAt.Time
		}
		if lastPushAt.Valid {
			s.LastPushAt = &lastPushAt.Time
		}

		systems = append(systems, s)
	}

	return &system.ListResult{
		Systems:    systems,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ListAll retrieves all systems without pagination.
func (r *SystemRepository) ListAll(ctx context.Context) ([]system.System, error) {
	query := `
		SELECT id, sn_sys_id, name, description, acronym, owner, status,
		       sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
		FROM systems
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list systems: %w", err)
	}
	defer rows.Close()

	systems := make([]system.System, 0)
	for rows.Next() {
		var s system.System
		var description, acronym, owner sql.NullString
		var snUpdatedOn, lastPullAt, lastPushAt sql.NullTime

		err := rows.Scan(
			&s.ID, &s.SNSysID, &s.Name, &description, &acronym, &owner, &s.Status,
			&snUpdatedOn, &lastPullAt, &lastPushAt, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan system: %w", err)
		}

		s.Description = description.String
		s.Acronym = acronym.String
		s.Owner = owner.String
		if snUpdatedOn.Valid {
			s.SNUpdatedOn = &snUpdatedOn.Time
		}
		if lastPullAt.Valid {
			s.LastPullAt = &lastPullAt.Time
		}
		if lastPushAt.Valid {
			s.LastPushAt = &lastPushAt.Time
		}

		systems = append(systems, s)
	}

	return systems, nil
}

// Upsert creates or updates a system based on sn_sys_id.
func (r *SystemRepository) Upsert(ctx context.Context, input system.UpsertInput) (*system.System, error) {
	query := `
		INSERT INTO systems (sn_sys_id, name, description, acronym, owner, status, sn_updated_on, last_pull_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (sn_sys_id)
		DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			acronym = EXCLUDED.acronym,
			owner = EXCLUDED.owner,
			status = EXCLUDED.status,
			sn_updated_on = EXCLUDED.sn_updated_on,
			last_pull_at = NOW(),
			updated_at = NOW()
		RETURNING id, sn_sys_id, name, description, acronym, owner, status,
		          sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
	`

	status := input.Status
	if status == "" {
		status = "active"
	}

	var s system.System
	var description, acronym, owner sql.NullString
	var snUpdatedOn, lastPullAt, lastPushAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query,
		input.SNSysID, input.Name, input.Description, input.Acronym, input.Owner, status, input.SNUpdatedOn,
	).Scan(
		&s.ID, &s.SNSysID, &s.Name, &description, &acronym, &owner, &s.Status,
		&snUpdatedOn, &lastPullAt, &lastPushAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert system: %w", err)
	}

	s.Description = description.String
	s.Acronym = acronym.String
	s.Owner = owner.String
	if snUpdatedOn.Valid {
		s.SNUpdatedOn = &snUpdatedOn.Time
	}
	if lastPullAt.Valid {
		s.LastPullAt = &lastPullAt.Time
	}
	if lastPushAt.Valid {
		s.LastPushAt = &lastPushAt.Time
	}

	return &s, nil
}

// UpsertBatch creates or updates multiple systems.
func (r *SystemRepository) UpsertBatch(ctx context.Context, inputs []system.UpsertInput) ([]system.System, error) {
	if len(inputs) == 0 {
		return []system.System{}, nil
	}

	// Use a transaction for batch operations
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	systems := make([]system.System, 0, len(inputs))

	for _, input := range inputs {
		query := `
			INSERT INTO systems (sn_sys_id, name, description, acronym, owner, status, sn_updated_on, last_pull_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
			ON CONFLICT (sn_sys_id)
			DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				acronym = EXCLUDED.acronym,
				owner = EXCLUDED.owner,
				status = EXCLUDED.status,
				sn_updated_on = EXCLUDED.sn_updated_on,
				last_pull_at = NOW(),
				updated_at = NOW()
			RETURNING id, sn_sys_id, name, description, acronym, owner, status,
			          sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
		`

		status := input.Status
		if status == "" {
			status = "active"
		}

		var s system.System
		var description, acronym, owner sql.NullString
		var snUpdatedOn, lastPullAt, lastPushAt sql.NullTime

		err := tx.QueryRowContext(ctx, query,
			input.SNSysID, input.Name, input.Description, input.Acronym, input.Owner, status, input.SNUpdatedOn,
		).Scan(
			&s.ID, &s.SNSysID, &s.Name, &description, &acronym, &owner, &s.Status,
			&snUpdatedOn, &lastPullAt, &lastPushAt, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert system %s: %w", input.SNSysID, err)
		}

		s.Description = description.String
		s.Acronym = acronym.String
		s.Owner = owner.String
		if snUpdatedOn.Valid {
			s.SNUpdatedOn = &snUpdatedOn.Time
		}
		if lastPullAt.Valid {
			s.LastPullAt = &lastPullAt.Time
		}
		if lastPushAt.Valid {
			s.LastPushAt = &lastPushAt.Time
		}

		systems = append(systems, s)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return systems, nil
}

// Delete removes a system and all its related controls/statements.
func (r *SystemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// CASCADE will handle controls and statements
	query := `DELETE FROM systems WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete system: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return system.ErrNotFound
	}

	return nil
}

// UpdateLastPullAt updates the last pull timestamp.
func (r *SystemRepository) UpdateLastPullAt(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE systems SET last_pull_at = $1, updated_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update last_pull_at: %w", err)
	}
	return nil
}

// GetAllSNSysIDs returns all ServiceNow sys_ids for existing systems.
func (r *SystemRepository) GetAllSNSysIDs(ctx context.Context) ([]string, error) {
	query := `SELECT sn_sys_id FROM systems`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get sys_ids: %w", err)
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan sys_id: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}
