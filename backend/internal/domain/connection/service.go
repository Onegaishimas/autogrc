package connection

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/controlcrud/backend/internal/infrastructure/crypto"
	"github.com/controlcrud/backend/internal/infrastructure/servicenow"
)

// Service provides business logic for connection management.
type Service struct {
	repo     Repository
	crypto   crypto.CryptoService
	snClient servicenow.Client
}

// NewService creates a new connection service.
func NewService(repo Repository, cryptoSvc crypto.CryptoService) *Service {
	return &Service{
		repo:   repo,
		crypto: cryptoSvc,
	}
}

// GetStatus returns the current connection status.
func (s *Service) GetStatus(ctx context.Context) (*Status, error) {
	conn, err := s.repo.GetActive(ctx)
	if err == ErrConnectionNotFound {
		return &Status{
			IsConfigured:   false,
			LastTestStatus: StatusUnknown,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active connection: %w", err)
	}

	return &Status{
		IsConfigured:            true,
		InstanceURL:             conn.InstanceURL,
		AuthMethod:              conn.AuthMethod,
		LastTestAt:              conn.LastTestAt,
		LastTestStatus:          conn.LastTestStatus,
		LastTestMessage:         conn.LastTestMessage,
		LastTestInstanceVersion: conn.LastTestInstanceVersion,
	}, nil
}

// SaveConfig saves a new connection configuration.
func (s *Service) SaveConfig(ctx context.Context, input *ConfigInput, userID *uuid.UUID) (*Connection, error) {
	// Validate input
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// Create new connection
	conn := &Connection{
		ID:             uuid.New(),
		InstanceURL:    input.InstanceURL,
		AuthMethod:     input.AuthMethod,
		IsActive:       true,
		LastTestStatus: StatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      userID,
		UpdatedBy:      userID,
	}

	// Encrypt credentials based on auth method
	switch input.AuthMethod {
	case AuthMethodBasic:
		conn.Username = input.Username

		// Encrypt password
		encrypted, nonce, err := s.crypto.Encrypt([]byte(input.Password))
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
		}
		conn.PasswordEncrypted = encrypted
		conn.PasswordNonce = nonce

	case AuthMethodOAuth:
		conn.OAuthClientID = input.OAuthClientID
		conn.OAuthTokenURL = input.OAuthTokenURL

		// Encrypt client secret
		encrypted, nonce, err := s.crypto.Encrypt([]byte(input.OAuthClientSecret))
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
		}
		conn.OAuthClientSecretEncrypted = encrypted
		conn.OAuthClientSecretNonce = nonce
	}

	// Deactivate existing connections and save new one
	if err := s.repo.DeactivateAll(ctx); err != nil {
		return nil, fmt.Errorf("failed to deactivate existing connections: %w", err)
	}

	if err := s.repo.Upsert(ctx, conn); err != nil {
		return nil, fmt.Errorf("failed to save connection: %w", err)
	}

	return conn, nil
}

// TestConnection tests the active connection and updates its status.
func (s *Service) TestConnection(ctx context.Context) (*TestResult, error) {
	// Get active connection
	conn, err := s.repo.GetActive(ctx)
	if err == ErrConnectionNotFound {
		return &TestResult{
			Success:      false,
			ErrorMessage: "no connection configured",
			TestedAt:     time.Now(),
		}, ErrConnectionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active connection: %w", err)
	}

	// Create ServiceNow client
	snConfig := servicenow.DefaultConfig(conn.InstanceURL)
	snClient, err := servicenow.NewSNClient(snConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create ServiceNow client: %w", err)
	}

	// Set authentication
	auth, err := s.getAuthProvider(conn)
	if err != nil {
		return nil, err
	}
	snClient.SetAuth(auth)

	// Test connection
	result, err := snClient.TestConnection(ctx)

	// Convert to domain result
	testResult := &TestResult{
		Success:        result.Success,
		ResponseTimeMs: result.ResponseTimeMs,
		TestedAt:       result.TestedAt,
	}

	if result.Success {
		testResult.InstanceVersion = result.InstanceInfo.Version
		testResult.BuildTag = result.InstanceInfo.BuildTag
	} else {
		testResult.ErrorMessage = result.ErrorMessage
	}

	// Update connection status in database
	status := StatusSuccess
	if !result.Success {
		status = StatusFailure
	}

	updateErr := s.repo.UpdateTestStatus(ctx, conn.ID, status, result.ErrorMessage, result.InstanceInfo.Version)
	if updateErr != nil {
		// Log but don't fail the test result
		// TODO: Add proper logging
	}

	return testResult, err
}

// DeleteConnection deletes the active connection.
func (s *Service) DeleteConnection(ctx context.Context) error {
	conn, err := s.repo.GetActive(ctx)
	if err == ErrConnectionNotFound {
		return nil // Already deleted
	}
	if err != nil {
		return fmt.Errorf("failed to get active connection: %w", err)
	}

	return s.repo.Delete(ctx, conn.ID)
}

// GetSNClient returns a configured ServiceNow client for the active connection.
// This method is used by other services that need to interact with ServiceNow.
func (s *Service) GetSNClient(ctx context.Context) (servicenow.Client, error) {
	conn, err := s.repo.GetActive(ctx)
	if err == ErrConnectionNotFound {
		return nil, ErrConnectionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active connection: %w", err)
	}

	// Create ServiceNow client
	snConfig := servicenow.DefaultConfig(conn.InstanceURL)
	snClient, err := servicenow.NewSNClient(snConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create ServiceNow client: %w", err)
	}

	// Set authentication
	auth, err := s.getAuthProvider(conn)
	if err != nil {
		return nil, err
	}
	snClient.SetAuth(auth)

	return snClient, nil
}

// getAuthProvider creates an auth provider for the connection.
func (s *Service) getAuthProvider(conn *Connection) (servicenow.AuthProvider, error) {
	switch conn.AuthMethod {
	case AuthMethodBasic:
		// Decrypt password
		password, err := s.crypto.Decrypt(conn.PasswordEncrypted, conn.PasswordNonce)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
		}
		return &servicenow.BasicAuthProvider{
			Username: conn.Username,
			Password: string(password),
		}, nil

	case AuthMethodOAuth:
		// Decrypt client secret
		secret, err := s.crypto.Decrypt(conn.OAuthClientSecretEncrypted, conn.OAuthClientSecretNonce)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
		}
		return &servicenow.OAuthProvider{
			ClientID:     conn.OAuthClientID,
			ClientSecret: string(secret),
			TokenURL:     conn.OAuthTokenURL,
		}, nil

	default:
		return nil, ErrInvalidAuthMethod
	}
}
