// Package database provides database implementations for repositories.
package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/controlcrud/backend/internal/domain/connection"
	"github.com/google/uuid"
)

// ConnectionRepository implements connection.Repository using PostgreSQL.
type ConnectionRepository struct {
	db *sql.DB
}

// NewConnectionRepository creates a new PostgreSQL connection repository.
func NewConnectionRepository(db *sql.DB) *ConnectionRepository {
	return &ConnectionRepository{db: db}
}

// GetActive retrieves the active connection configuration.
func (r *ConnectionRepository) GetActive(ctx context.Context) (*connection.Connection, error) {
	query := `
		SELECT
			id, instance_url, auth_method,
			username, password_encrypted, password_nonce,
			oauth_client_id, oauth_client_secret_encrypted, oauth_client_secret_nonce, oauth_token_url,
			is_active, last_test_at, last_test_status, last_test_message, last_test_instance_version,
			created_at, updated_at, created_by, updated_by
		FROM servicenow_connections
		WHERE is_active = true
		LIMIT 1
	`

	var conn connection.Connection
	var lastTestAt sql.NullTime
	var lastTestStatus sql.NullString
	var lastTestMessage sql.NullString
	var lastTestInstanceVersion sql.NullString
	var createdBy, updatedBy sql.NullString

	err := r.db.QueryRowContext(ctx, query).Scan(
		&conn.ID, &conn.InstanceURL, &conn.AuthMethod,
		&conn.Username, &conn.PasswordEncrypted, &conn.PasswordNonce,
		&conn.OAuthClientID, &conn.OAuthClientSecretEncrypted, &conn.OAuthClientSecretNonce, &conn.OAuthTokenURL,
		&conn.IsActive, &lastTestAt, &lastTestStatus, &lastTestMessage, &lastTestInstanceVersion,
		&conn.CreatedAt, &conn.UpdatedAt, &createdBy, &updatedBy,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connection.ErrConnectionNotFound
		}
		return nil, err
	}

	if lastTestAt.Valid {
		conn.LastTestAt = &lastTestAt.Time
	}
	if lastTestStatus.Valid {
		conn.LastTestStatus = connection.ConnectionStatus(lastTestStatus.String)
	}
	if lastTestMessage.Valid {
		conn.LastTestMessage = lastTestMessage.String
	}
	if lastTestInstanceVersion.Valid {
		conn.LastTestInstanceVersion = lastTestInstanceVersion.String
	}
	if createdBy.Valid {
		id, _ := uuid.Parse(createdBy.String)
		conn.CreatedBy = &id
	}
	if updatedBy.Valid {
		id, _ := uuid.Parse(updatedBy.String)
		conn.UpdatedBy = &id
	}

	return &conn, nil
}

// GetByID retrieves a connection by its ID.
func (r *ConnectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*connection.Connection, error) {
	query := `
		SELECT
			id, instance_url, auth_method,
			username, password_encrypted, password_nonce,
			oauth_client_id, oauth_client_secret_encrypted, oauth_client_secret_nonce, oauth_token_url,
			is_active, last_test_at, last_test_status, last_test_message, last_test_instance_version,
			created_at, updated_at, created_by, updated_by
		FROM servicenow_connections
		WHERE id = $1
	`

	var conn connection.Connection
	var lastTestAt sql.NullTime
	var lastTestStatus sql.NullString
	var lastTestMessage sql.NullString
	var lastTestInstanceVersion sql.NullString
	var createdBy, updatedBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&conn.ID, &conn.InstanceURL, &conn.AuthMethod,
		&conn.Username, &conn.PasswordEncrypted, &conn.PasswordNonce,
		&conn.OAuthClientID, &conn.OAuthClientSecretEncrypted, &conn.OAuthClientSecretNonce, &conn.OAuthTokenURL,
		&conn.IsActive, &lastTestAt, &lastTestStatus, &lastTestMessage, &lastTestInstanceVersion,
		&conn.CreatedAt, &conn.UpdatedAt, &createdBy, &updatedBy,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connection.ErrConnectionNotFound
		}
		return nil, err
	}

	if lastTestAt.Valid {
		conn.LastTestAt = &lastTestAt.Time
	}
	if lastTestStatus.Valid {
		conn.LastTestStatus = connection.ConnectionStatus(lastTestStatus.String)
	}
	if lastTestMessage.Valid {
		conn.LastTestMessage = lastTestMessage.String
	}
	if lastTestInstanceVersion.Valid {
		conn.LastTestInstanceVersion = lastTestInstanceVersion.String
	}
	if createdBy.Valid {
		id, _ := uuid.Parse(createdBy.String)
		conn.CreatedBy = &id
	}
	if updatedBy.Valid {
		id, _ := uuid.Parse(updatedBy.String)
		conn.UpdatedBy = &id
	}

	return &conn, nil
}

// Upsert creates or updates a connection configuration.
func (r *ConnectionRepository) Upsert(ctx context.Context, conn *connection.Connection) error {
	query := `
		INSERT INTO servicenow_connections (
			id, instance_url, auth_method,
			username, password_encrypted, password_nonce,
			oauth_client_id, oauth_client_secret_encrypted, oauth_client_secret_nonce, oauth_token_url,
			is_active, last_test_status,
			created_at, updated_at, created_by, updated_by
		) VALUES (
			$1, $2, $3,
			$4, $5, $6,
			$7, $8, $9, $10,
			$11, $12,
			$13, $14, $15, $16
		)
		ON CONFLICT (id) DO UPDATE SET
			instance_url = EXCLUDED.instance_url,
			auth_method = EXCLUDED.auth_method,
			username = EXCLUDED.username,
			password_encrypted = EXCLUDED.password_encrypted,
			password_nonce = EXCLUDED.password_nonce,
			oauth_client_id = EXCLUDED.oauth_client_id,
			oauth_client_secret_encrypted = EXCLUDED.oauth_client_secret_encrypted,
			oauth_client_secret_nonce = EXCLUDED.oauth_client_secret_nonce,
			oauth_token_url = EXCLUDED.oauth_token_url,
			is_active = EXCLUDED.is_active,
			last_test_status = EXCLUDED.last_test_status,
			updated_at = EXCLUDED.updated_at,
			updated_by = EXCLUDED.updated_by
	`

	now := time.Now()
	if conn.CreatedAt.IsZero() {
		conn.CreatedAt = now
	}
	conn.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		conn.ID, conn.InstanceURL, conn.AuthMethod,
		conn.Username, conn.PasswordEncrypted, conn.PasswordNonce,
		conn.OAuthClientID, conn.OAuthClientSecretEncrypted, conn.OAuthClientSecretNonce, conn.OAuthTokenURL,
		conn.IsActive, conn.LastTestStatus,
		conn.CreatedAt, conn.UpdatedAt, conn.CreatedBy, conn.UpdatedBy,
	)

	return err
}

// UpdateTestStatus updates the test status for a connection.
func (r *ConnectionRepository) UpdateTestStatus(ctx context.Context, id uuid.UUID, status connection.ConnectionStatus, message string, version string) error {
	query := `
		UPDATE servicenow_connections
		SET
			last_test_at = $2,
			last_test_status = $3,
			last_test_message = $4,
			last_test_instance_version = $5,
			updated_at = $6
		WHERE id = $1
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, id, now, status, message, version, now)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return connection.ErrConnectionNotFound
	}

	return nil
}

// Delete removes a connection configuration.
func (r *ConnectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM servicenow_connections WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return connection.ErrConnectionNotFound
	}

	return nil
}

// DeactivateAll deactivates all connection configurations.
func (r *ConnectionRepository) DeactivateAll(ctx context.Context) error {
	query := `UPDATE servicenow_connections SET is_active = false, updated_at = $1`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}
