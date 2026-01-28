package connection

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for connection data persistence.
type Repository interface {
	// GetActive returns the currently active connection, if any.
	// Returns ErrConnectionNotFound if no active connection exists.
	GetActive(ctx context.Context) (*Connection, error)

	// GetByID returns a connection by its ID.
	// Returns ErrConnectionNotFound if the connection does not exist.
	GetByID(ctx context.Context, id uuid.UUID) (*Connection, error)

	// Upsert creates or updates a connection.
	// If an active connection exists, it will be deactivated first.
	Upsert(ctx context.Context, conn *Connection) error

	// UpdateTestStatus updates the connection's test status fields.
	UpdateTestStatus(ctx context.Context, id uuid.UUID, status ConnectionStatus, message string, version string) error

	// Delete removes a connection by its ID.
	// Returns ErrConnectionNotFound if the connection does not exist.
	Delete(ctx context.Context, id uuid.UUID) error

	// DeactivateAll deactivates all connections.
	DeactivateAll(ctx context.Context) error
}
