package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/controlcrud/backend/internal/domain/control"
)

// ControlRepository implements control.Repository using PostgreSQL.
type ControlRepository struct {
	db *sql.DB
}

// NewControlRepository creates a new control repository.
func NewControlRepository(db *sql.DB) *ControlRepository {
	return &ControlRepository{db: db}
}

// GetByID retrieves a control by its internal ID.
func (r *ControlRepository) GetByID(ctx context.Context, id uuid.UUID) (*control.Control, error) {
	query := `
		SELECT id, system_id, sn_sys_id, control_id, control_name, control_family,
		       description, implementation_status, responsible_role,
		       sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
		FROM controls
		WHERE id = $1
	`

	return r.scanControl(r.db.QueryRowContext(ctx, query, id))
}

// GetBySNSysID retrieves a control by system ID and ServiceNow sys_id.
func (r *ControlRepository) GetBySNSysID(ctx context.Context, systemID uuid.UUID, snSysID string) (*control.Control, error) {
	query := `
		SELECT id, system_id, sn_sys_id, control_id, control_name, control_family,
		       description, implementation_status, responsible_role,
		       sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
		FROM controls
		WHERE system_id = $1 AND sn_sys_id = $2
	`

	return r.scanControl(r.db.QueryRowContext(ctx, query, systemID, snSysID))
}

// List retrieves controls for a system with pagination.
func (r *ControlRepository) List(ctx context.Context, params control.ListParams) (*control.ListResult, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, fmt.Sprintf("c.system_id = $%d", argNum))
	args = append(args, params.SystemID)
	argNum++

	if params.ControlFamily != "" {
		conditions = append(conditions, fmt.Sprintf("c.control_family = $%d", argNum))
		args = append(args, params.ControlFamily)
		argNum++
	}

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(c.control_id ILIKE $%d OR c.control_name ILIKE $%d)", argNum, argNum))
		args = append(args, "%"+params.Search+"%")
		argNum++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM controls c %s`, whereClause)
	var totalCount int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("failed to count controls: %w", err)
	}

	// Calculate pagination
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	offset := (params.Page - 1) * params.PageSize
	totalPages := (totalCount + params.PageSize - 1) / params.PageSize

	// Fetch controls with stats
	query := fmt.Sprintf(`
		SELECT c.id, c.system_id, c.sn_sys_id, c.control_id, c.control_name, c.control_family,
		       c.description, c.implementation_status, c.responsible_role,
		       c.sn_updated_on, c.last_pull_at, c.last_push_at, c.created_at, c.updated_at,
		       COALESCE((SELECT COUNT(*) FROM statements s WHERE s.control_id = c.id), 0) as statement_count,
		       COALESCE((SELECT COUNT(*) FROM statements s WHERE s.control_id = c.id AND s.is_modified = true), 0) as modified_count
		FROM controls c
		%s
		ORDER BY c.control_id ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)

	args = append(args, params.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list controls: %w", err)
	}
	defer rows.Close()

	controls := make([]control.ControlWithStats, 0)
	for rows.Next() {
		var c control.ControlWithStats
		var description, responsibleRole sql.NullString
		var snUpdatedOn, lastPullAt, lastPushAt sql.NullTime

		err := rows.Scan(
			&c.ID, &c.SystemID, &c.SNSysID, &c.ControlID, &c.ControlName, &c.ControlFamily,
			&description, &c.ImplementationStatus, &responsibleRole,
			&snUpdatedOn, &lastPullAt, &lastPushAt, &c.CreatedAt, &c.UpdatedAt,
			&c.StatementCount, &c.ModifiedCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan control: %w", err)
		}

		c.Description = description.String
		c.ResponsibleRole = responsibleRole.String
		if snUpdatedOn.Valid {
			c.SNUpdatedOn = &snUpdatedOn.Time
		}
		if lastPullAt.Valid {
			c.LastPullAt = &lastPullAt.Time
		}
		if lastPushAt.Valid {
			c.LastPushAt = &lastPushAt.Time
		}

		controls = append(controls, c)
	}

	return &control.ListResult{
		Controls:   controls,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ListBySystem retrieves all controls for a system.
func (r *ControlRepository) ListBySystem(ctx context.Context, systemID uuid.UUID) ([]control.Control, error) {
	query := `
		SELECT id, system_id, sn_sys_id, control_id, control_name, control_family,
		       description, implementation_status, responsible_role,
		       sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
		FROM controls
		WHERE system_id = $1
		ORDER BY control_id ASC
	`

	rows, err := r.db.QueryContext(ctx, query, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to list controls: %w", err)
	}
	defer rows.Close()

	controls := make([]control.Control, 0)
	for rows.Next() {
		c, err := r.scanControlFromRows(rows)
		if err != nil {
			return nil, err
		}
		controls = append(controls, *c)
	}

	return controls, nil
}

// Upsert creates or updates a control.
func (r *ControlRepository) Upsert(ctx context.Context, input control.UpsertInput) (*control.Control, error) {
	query := `
		INSERT INTO controls (system_id, sn_sys_id, control_id, control_name, control_family,
		                      description, implementation_status, responsible_role, sn_updated_on, last_pull_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (system_id, sn_sys_id)
		DO UPDATE SET
			control_id = EXCLUDED.control_id,
			control_name = EXCLUDED.control_name,
			control_family = EXCLUDED.control_family,
			description = EXCLUDED.description,
			implementation_status = EXCLUDED.implementation_status,
			responsible_role = EXCLUDED.responsible_role,
			sn_updated_on = EXCLUDED.sn_updated_on,
			last_pull_at = NOW(),
			updated_at = NOW()
		RETURNING id, system_id, sn_sys_id, control_id, control_name, control_family,
		          description, implementation_status, responsible_role,
		          sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
	`

	status := input.ImplementationStatus
	if status == "" {
		status = "not_assessed"
	}

	return r.scanControl(r.db.QueryRowContext(ctx, query,
		input.SystemID, input.SNSysID, input.ControlID, input.ControlName, input.ControlFamily,
		input.Description, status, input.ResponsibleRole, input.SNUpdatedOn,
	))
}

// UpsertBatch creates or updates multiple controls.
func (r *ControlRepository) UpsertBatch(ctx context.Context, inputs []control.UpsertInput) ([]control.Control, error) {
	if len(inputs) == 0 {
		return []control.Control{}, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	controls := make([]control.Control, 0, len(inputs))

	for _, input := range inputs {
		query := `
			INSERT INTO controls (system_id, sn_sys_id, control_id, control_name, control_family,
			                      description, implementation_status, responsible_role, sn_updated_on, last_pull_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
			ON CONFLICT (system_id, sn_sys_id)
			DO UPDATE SET
				control_id = EXCLUDED.control_id,
				control_name = EXCLUDED.control_name,
				control_family = EXCLUDED.control_family,
				description = EXCLUDED.description,
				implementation_status = EXCLUDED.implementation_status,
				responsible_role = EXCLUDED.responsible_role,
				sn_updated_on = EXCLUDED.sn_updated_on,
				last_pull_at = NOW(),
				updated_at = NOW()
			RETURNING id, system_id, sn_sys_id, control_id, control_name, control_family,
			          description, implementation_status, responsible_role,
			          sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
		`

		status := input.ImplementationStatus
		if status == "" {
			status = "not_assessed"
		}

		var c control.Control
		var description, responsibleRole sql.NullString
		var snUpdatedOn, lastPullAt, lastPushAt sql.NullTime

		err := tx.QueryRowContext(ctx, query,
			input.SystemID, input.SNSysID, input.ControlID, input.ControlName, input.ControlFamily,
			input.Description, status, input.ResponsibleRole, input.SNUpdatedOn,
		).Scan(
			&c.ID, &c.SystemID, &c.SNSysID, &c.ControlID, &c.ControlName, &c.ControlFamily,
			&description, &c.ImplementationStatus, &responsibleRole,
			&snUpdatedOn, &lastPullAt, &lastPushAt, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert control: %w", err)
		}

		c.Description = description.String
		c.ResponsibleRole = responsibleRole.String
		if snUpdatedOn.Valid {
			c.SNUpdatedOn = &snUpdatedOn.Time
		}
		if lastPullAt.Valid {
			c.LastPullAt = &lastPullAt.Time
		}
		if lastPushAt.Valid {
			c.LastPushAt = &lastPushAt.Time
		}

		controls = append(controls, c)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return controls, nil
}

// Delete removes a control and its statements.
func (r *ControlRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM controls WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete control: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return control.ErrNotFound
	}

	return nil
}

// DeleteBySystem removes all controls for a system.
func (r *ControlRepository) DeleteBySystem(ctx context.Context, systemID uuid.UUID) error {
	query := `DELETE FROM controls WHERE system_id = $1`
	_, err := r.db.ExecContext(ctx, query, systemID)
	if err != nil {
		return fmt.Errorf("failed to delete controls: %w", err)
	}
	return nil
}

// Helper functions

func (r *ControlRepository) scanControl(row *sql.Row) (*control.Control, error) {
	var c control.Control
	var description, responsibleRole sql.NullString
	var snUpdatedOn, lastPullAt, lastPushAt sql.NullTime

	err := row.Scan(
		&c.ID, &c.SystemID, &c.SNSysID, &c.ControlID, &c.ControlName, &c.ControlFamily,
		&description, &c.ImplementationStatus, &responsibleRole,
		&snUpdatedOn, &lastPullAt, &lastPushAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan control: %w", err)
	}

	c.Description = description.String
	c.ResponsibleRole = responsibleRole.String
	if snUpdatedOn.Valid {
		c.SNUpdatedOn = &snUpdatedOn.Time
	}
	if lastPullAt.Valid {
		c.LastPullAt = &lastPullAt.Time
	}
	if lastPushAt.Valid {
		c.LastPushAt = &lastPushAt.Time
	}

	return &c, nil
}

func (r *ControlRepository) scanControlFromRows(rows *sql.Rows) (*control.Control, error) {
	var c control.Control
	var description, responsibleRole sql.NullString
	var snUpdatedOn, lastPullAt, lastPushAt sql.NullTime

	err := rows.Scan(
		&c.ID, &c.SystemID, &c.SNSysID, &c.ControlID, &c.ControlName, &c.ControlFamily,
		&description, &c.ImplementationStatus, &responsibleRole,
		&snUpdatedOn, &lastPullAt, &lastPushAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan control: %w", err)
	}

	c.Description = description.String
	c.ResponsibleRole = responsibleRole.String
	if snUpdatedOn.Valid {
		c.SNUpdatedOn = &snUpdatedOn.Time
	}
	if lastPullAt.Valid {
		c.LastPullAt = &lastPullAt.Time
	}
	if lastPushAt.Valid {
		c.LastPushAt = &lastPushAt.Time
	}

	return &c, nil
}
