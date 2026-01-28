package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/controlcrud/backend/internal/domain/statement"
)

// StatementRepository implements statement.Repository using PostgreSQL.
type StatementRepository struct {
	db *sql.DB
}

// NewStatementRepository creates a new statement repository.
func NewStatementRepository(db *sql.DB) *StatementRepository {
	return &StatementRepository{db: db}
}

// GetByID retrieves a statement by its internal ID.
func (r *StatementRepository) GetByID(ctx context.Context, id uuid.UUID) (*statement.Statement, error) {
	query := `
		SELECT id, control_id, sn_sys_id, statement_type,
		       remote_content, remote_updated_at, local_content, is_modified, modified_at, modified_by,
		       sync_status, conflict_resolved_at, conflict_resolved_by,
		       sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
		FROM statements
		WHERE id = $1
	`

	return r.scanStatement(r.db.QueryRowContext(ctx, query, id))
}

// GetBySNSysID retrieves a statement by control ID and ServiceNow sys_id.
func (r *StatementRepository) GetBySNSysID(ctx context.Context, controlID uuid.UUID, snSysID string) (*statement.Statement, error) {
	query := `
		SELECT id, control_id, sn_sys_id, statement_type,
		       remote_content, remote_updated_at, local_content, is_modified, modified_at, modified_by,
		       sync_status, conflict_resolved_at, conflict_resolved_by,
		       sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
		FROM statements
		WHERE control_id = $1 AND sn_sys_id = $2
	`

	return r.scanStatement(r.db.QueryRowContext(ctx, query, controlID, snSysID))
}

// List retrieves statements with pagination. Filters by control_id OR system_id (joins through controls).
func (r *StatementRepository) List(ctx context.Context, params statement.ListParams) (*statement.ListResult, error) {
	var conditions []string
	var args []interface{}
	argNum := 1
	needsJoin := false

	// Handle system_id filter (joins through controls table)
	if params.SystemID != uuid.Nil {
		needsJoin = true
		conditions = append(conditions, fmt.Sprintf("c.system_id = $%d", argNum))
		args = append(args, params.SystemID)
		argNum++
	} else if params.ControlID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("s.control_id = $%d", argNum))
		args = append(args, params.ControlID)
		argNum++
	}

	if params.SyncStatus != "" {
		conditions = append(conditions, fmt.Sprintf("s.sync_status = $%d", argNum))
		args = append(args, params.SyncStatus)
		argNum++
	}

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(s.remote_content ILIKE $%d OR s.local_content ILIKE $%d)", argNum, argNum))
		args = append(args, "%"+params.Search+"%")
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	fromClause := "FROM statements s"
	if needsJoin {
		fromClause = "FROM statements s JOIN controls c ON s.control_id = c.id"
	}

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) %s %s`, fromClause, whereClause)
	var totalCount int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("failed to count statements: %w", err)
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

	// Fetch statements
	query := fmt.Sprintf(`
		SELECT s.id, s.control_id, s.sn_sys_id, s.statement_type,
		       s.remote_content, s.remote_updated_at, s.local_content, s.is_modified, s.modified_at, s.modified_by,
		       s.sync_status, s.conflict_resolved_at, s.conflict_resolved_by,
		       s.sn_updated_on, s.last_pull_at, s.last_push_at, s.created_at, s.updated_at
		%s
		%s
		ORDER BY s.created_at ASC
		LIMIT $%d OFFSET $%d
	`, fromClause, whereClause, argNum, argNum+1)

	args = append(args, params.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list statements: %w", err)
	}
	defer rows.Close()

	statements := make([]statement.Statement, 0)
	for rows.Next() {
		s, err := r.scanStatementFromRows(rows)
		if err != nil {
			return nil, err
		}
		statements = append(statements, *s)
	}

	return &statement.ListResult{
		Statements: statements,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ListByControl retrieves all statements for a control.
func (r *StatementRepository) ListByControl(ctx context.Context, controlID uuid.UUID) ([]statement.Statement, error) {
	query := `
		SELECT id, control_id, sn_sys_id, statement_type,
		       remote_content, remote_updated_at, local_content, is_modified, modified_at, modified_by,
		       sync_status, conflict_resolved_at, conflict_resolved_by,
		       sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
		FROM statements
		WHERE control_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, controlID)
	if err != nil {
		return nil, fmt.Errorf("failed to list statements: %w", err)
	}
	defer rows.Close()

	statements := make([]statement.Statement, 0)
	for rows.Next() {
		s, err := r.scanStatementFromRows(rows)
		if err != nil {
			return nil, err
		}
		statements = append(statements, *s)
	}

	return statements, nil
}

// ListModified retrieves all statements with local modifications.
func (r *StatementRepository) ListModified(ctx context.Context) ([]statement.Statement, error) {
	query := `
		SELECT id, control_id, sn_sys_id, statement_type,
		       remote_content, remote_updated_at, local_content, is_modified, modified_at, modified_by,
		       sync_status, conflict_resolved_at, conflict_resolved_by,
		       sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
		FROM statements
		WHERE is_modified = true
		ORDER BY modified_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list modified statements: %w", err)
	}
	defer rows.Close()

	statements := make([]statement.Statement, 0)
	for rows.Next() {
		s, err := r.scanStatementFromRows(rows)
		if err != nil {
			return nil, err
		}
		statements = append(statements, *s)
	}

	return statements, nil
}

// ListConflicts retrieves all statements with sync conflicts.
func (r *StatementRepository) ListConflicts(ctx context.Context) ([]statement.Statement, error) {
	query := `
		SELECT id, control_id, sn_sys_id, statement_type,
		       remote_content, remote_updated_at, local_content, is_modified, modified_at, modified_by,
		       sync_status, conflict_resolved_at, conflict_resolved_by,
		       sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
		FROM statements
		WHERE sync_status = 'conflict'
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list conflicts: %w", err)
	}
	defer rows.Close()

	statements := make([]statement.Statement, 0)
	for rows.Next() {
		s, err := r.scanStatementFromRows(rows)
		if err != nil {
			return nil, err
		}
		statements = append(statements, *s)
	}

	return statements, nil
}

// Upsert creates or updates a statement from ServiceNow.
func (r *StatementRepository) Upsert(ctx context.Context, input statement.UpsertInput) (*statement.Statement, error) {
	// Check if statement exists and has local modifications
	existing, _ := r.GetBySNSysID(ctx, input.ControlID, input.SNSysID)

	var query string
	if existing != nil && existing.IsModified {
		// Detect conflict: if remote content changed while we have local changes
		if existing.RemoteContent != input.RemoteContent {
			query = `
				UPDATE statements SET
					remote_content = $3,
					remote_updated_at = $4,
					sn_updated_on = $5,
					sync_status = 'conflict',
					last_pull_at = NOW(),
					updated_at = NOW()
				WHERE control_id = $1 AND sn_sys_id = $2
				RETURNING id, control_id, sn_sys_id, statement_type,
				          remote_content, remote_updated_at, local_content, is_modified, modified_at, modified_by,
				          sync_status, conflict_resolved_at, conflict_resolved_by,
				          sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
			`
			return r.scanStatement(r.db.QueryRowContext(ctx, query,
				input.ControlID, input.SNSysID, input.RemoteContent, time.Now(), input.SNUpdatedOn,
			))
		}
		// No conflict - remote hasn't changed
		return existing, nil
	}

	// Normal upsert (new or no local modifications)
	query = `
		INSERT INTO statements (control_id, sn_sys_id, statement_type, remote_content, remote_updated_at, sn_updated_on, last_pull_at)
		VALUES ($1, $2, $3, $4, NOW(), $5, NOW())
		ON CONFLICT (control_id, sn_sys_id)
		DO UPDATE SET
			remote_content = EXCLUDED.remote_content,
			remote_updated_at = NOW(),
			sn_updated_on = EXCLUDED.sn_updated_on,
			last_pull_at = NOW(),
			updated_at = NOW()
		RETURNING id, control_id, sn_sys_id, statement_type,
		          remote_content, remote_updated_at, local_content, is_modified, modified_at, modified_by,
		          sync_status, conflict_resolved_at, conflict_resolved_by,
		          sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
	`

	stmtType := input.StatementType
	if stmtType == "" {
		stmtType = "implementation"
	}

	return r.scanStatement(r.db.QueryRowContext(ctx, query,
		input.ControlID, input.SNSysID, stmtType, input.RemoteContent, input.SNUpdatedOn,
	))
}

// UpsertBatch creates or updates multiple statements.
func (r *StatementRepository) UpsertBatch(ctx context.Context, inputs []statement.UpsertInput) ([]statement.Statement, error) {
	if len(inputs) == 0 {
		return []statement.Statement{}, nil
	}

	statements := make([]statement.Statement, 0, len(inputs))
	for _, input := range inputs {
		s, err := r.Upsert(ctx, input)
		if err != nil {
			return nil, err
		}
		statements = append(statements, *s)
	}

	return statements, nil
}

// UpdateLocal updates the local content of a statement.
func (r *StatementRepository) UpdateLocal(ctx context.Context, input statement.UpdateInput) (*statement.Statement, error) {
	query := `
		UPDATE statements SET
			local_content = $2,
			is_modified = true,
			modified_at = NOW(),
			modified_by = $3,
			sync_status = 'modified',
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, control_id, sn_sys_id, statement_type,
		          remote_content, remote_updated_at, local_content, is_modified, modified_at, modified_by,
		          sync_status, conflict_resolved_at, conflict_resolved_by,
		          sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
	`

	return r.scanStatement(r.db.QueryRowContext(ctx, query, input.ID, input.LocalContent, input.ModifiedBy))
}

// ResolveConflict resolves a sync conflict.
func (r *StatementRepository) ResolveConflict(ctx context.Context, input statement.ResolveConflictInput) (*statement.Statement, error) {
	var query string
	var args []interface{}

	switch input.Resolution {
	case statement.ConflictResolutionKeepLocal:
		query = `
			UPDATE statements SET
				sync_status = 'modified',
				conflict_resolved_at = NOW(),
				conflict_resolved_by = $2,
				updated_at = NOW()
			WHERE id = $1
			RETURNING id, control_id, sn_sys_id, statement_type,
			          remote_content, remote_updated_at, local_content, is_modified, modified_at, modified_by,
			          sync_status, conflict_resolved_at, conflict_resolved_by,
			          sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
		`
		args = []interface{}{input.ID, input.ResolvedBy}

	case statement.ConflictResolutionKeepRemote:
		query = `
			UPDATE statements SET
				local_content = remote_content,
				is_modified = false,
				sync_status = 'synced',
				conflict_resolved_at = NOW(),
				conflict_resolved_by = $2,
				updated_at = NOW()
			WHERE id = $1
			RETURNING id, control_id, sn_sys_id, statement_type,
			          remote_content, remote_updated_at, local_content, is_modified, modified_at, modified_by,
			          sync_status, conflict_resolved_at, conflict_resolved_by,
			          sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
		`
		args = []interface{}{input.ID, input.ResolvedBy}

	case statement.ConflictResolutionMerge:
		query = `
			UPDATE statements SET
				local_content = $2,
				is_modified = true,
				sync_status = 'modified',
				conflict_resolved_at = NOW(),
				conflict_resolved_by = $3,
				updated_at = NOW()
			WHERE id = $1
			RETURNING id, control_id, sn_sys_id, statement_type,
			          remote_content, remote_updated_at, local_content, is_modified, modified_at, modified_by,
			          sync_status, conflict_resolved_at, conflict_resolved_by,
			          sn_updated_on, last_pull_at, last_push_at, created_at, updated_at
		`
		args = []interface{}{input.ID, input.MergedContent, input.ResolvedBy}

	default:
		return nil, fmt.Errorf("invalid conflict resolution: %s", input.Resolution)
	}

	return r.scanStatement(r.db.QueryRowContext(ctx, query, args...))
}

// Delete removes a statement.
func (r *StatementRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM statements WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete statement: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return statement.ErrNotFound
	}

	return nil
}

// DeleteByControl removes all statements for a control.
func (r *StatementRepository) DeleteByControl(ctx context.Context, controlID uuid.UUID) error {
	query := `DELETE FROM statements WHERE control_id = $1`
	_, err := r.db.ExecContext(ctx, query, controlID)
	if err != nil {
		return fmt.Errorf("failed to delete statements: %w", err)
	}
	return nil
}

// MarkAsSynced marks a statement as synced after push.
func (r *StatementRepository) MarkAsSynced(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE statements SET
			is_modified = false,
			sync_status = 'synced',
			last_push_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark as synced: %w", err)
	}
	return nil
}

// Helper functions

func (r *StatementRepository) scanStatement(row *sql.Row) (*statement.Statement, error) {
	var s statement.Statement
	var remoteContent, localContent sql.NullString
	var remoteUpdatedAt, modifiedAt, conflictResolvedAt, snUpdatedOn, lastPullAt, lastPushAt sql.NullTime
	var modifiedBy, conflictResolvedBy sql.NullString

	err := row.Scan(
		&s.ID, &s.ControlID, &s.SNSysID, &s.StatementType,
		&remoteContent, &remoteUpdatedAt, &localContent, &s.IsModified, &modifiedAt, &modifiedBy,
		&s.SyncStatus, &conflictResolvedAt, &conflictResolvedBy,
		&snUpdatedOn, &lastPullAt, &lastPushAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan statement: %w", err)
	}

	s.RemoteContent = remoteContent.String
	s.LocalContent = localContent.String
	if remoteUpdatedAt.Valid {
		s.RemoteUpdatedAt = &remoteUpdatedAt.Time
	}
	if modifiedAt.Valid {
		s.ModifiedAt = &modifiedAt.Time
	}
	if modifiedBy.Valid {
		if id, err := uuid.Parse(modifiedBy.String); err == nil {
			s.ModifiedBy = &id
		}
	}
	if conflictResolvedAt.Valid {
		s.ConflictResolvedAt = &conflictResolvedAt.Time
	}
	if conflictResolvedBy.Valid {
		if id, err := uuid.Parse(conflictResolvedBy.String); err == nil {
			s.ConflictResolvedBy = &id
		}
	}
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

func (r *StatementRepository) scanStatementFromRows(rows *sql.Rows) (*statement.Statement, error) {
	var s statement.Statement
	var remoteContent, localContent sql.NullString
	var remoteUpdatedAt, modifiedAt, conflictResolvedAt, snUpdatedOn, lastPullAt, lastPushAt sql.NullTime
	var modifiedBy, conflictResolvedBy sql.NullString

	err := rows.Scan(
		&s.ID, &s.ControlID, &s.SNSysID, &s.StatementType,
		&remoteContent, &remoteUpdatedAt, &localContent, &s.IsModified, &modifiedAt, &modifiedBy,
		&s.SyncStatus, &conflictResolvedAt, &conflictResolvedBy,
		&snUpdatedOn, &lastPullAt, &lastPushAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan statement: %w", err)
	}

	s.RemoteContent = remoteContent.String
	s.LocalContent = localContent.String
	if remoteUpdatedAt.Valid {
		s.RemoteUpdatedAt = &remoteUpdatedAt.Time
	}
	if modifiedAt.Valid {
		s.ModifiedAt = &modifiedAt.Time
	}
	if modifiedBy.Valid {
		if id, err := uuid.Parse(modifiedBy.String); err == nil {
			s.ModifiedBy = &id
		}
	}
	if conflictResolvedAt.Valid {
		s.ConflictResolvedAt = &conflictResolvedAt.Time
	}
	if conflictResolvedBy.Valid {
		if id, err := uuid.Parse(conflictResolvedBy.String); err == nil {
			s.ConflictResolvedBy = &id
		}
	}
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
