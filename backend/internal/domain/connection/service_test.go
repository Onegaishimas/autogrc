package connection

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// mockRepository implements Repository for testing.
type mockRepository struct {
	activeConn *Connection
	conns      map[uuid.UUID]*Connection
	err        error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		conns: make(map[uuid.UUID]*Connection),
	}
}

func (m *mockRepository) GetActive(ctx context.Context) (*Connection, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.activeConn == nil {
		return nil, ErrConnectionNotFound
	}
	return m.activeConn, nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uuid.UUID) (*Connection, error) {
	if m.err != nil {
		return nil, m.err
	}
	conn, ok := m.conns[id]
	if !ok {
		return nil, ErrConnectionNotFound
	}
	return conn, nil
}

func (m *mockRepository) Upsert(ctx context.Context, conn *Connection) error {
	if m.err != nil {
		return m.err
	}
	m.conns[conn.ID] = conn
	if conn.IsActive {
		m.activeConn = conn
	}
	return nil
}

func (m *mockRepository) UpdateTestStatus(ctx context.Context, id uuid.UUID, status ConnectionStatus, message string, version string) error {
	if m.err != nil {
		return m.err
	}
	conn, ok := m.conns[id]
	if !ok {
		return ErrConnectionNotFound
	}
	now := time.Now()
	conn.LastTestAt = &now
	conn.LastTestStatus = status
	conn.LastTestMessage = message
	conn.LastTestInstanceVersion = version
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.err != nil {
		return m.err
	}
	delete(m.conns, id)
	if m.activeConn != nil && m.activeConn.ID == id {
		m.activeConn = nil
	}
	return nil
}

func (m *mockRepository) DeactivateAll(ctx context.Context) error {
	if m.err != nil {
		return m.err
	}
	for _, conn := range m.conns {
		conn.IsActive = false
	}
	m.activeConn = nil
	return nil
}

// mockCrypto implements crypto.CryptoService for testing.
type mockCrypto struct {
	encryptErr error
	decryptErr error
}

func (m *mockCrypto) Encrypt(plaintext []byte) ([]byte, []byte, error) {
	if m.encryptErr != nil {
		return nil, nil, m.encryptErr
	}
	// Simple "encryption" for testing - just reverse the bytes
	encrypted := make([]byte, len(plaintext))
	for i, b := range plaintext {
		encrypted[len(plaintext)-1-i] = b
	}
	nonce := []byte("testnonce123")
	return encrypted, nonce, nil
}

func (m *mockCrypto) Decrypt(ciphertext []byte, nonce []byte) ([]byte, error) {
	if m.decryptErr != nil {
		return nil, m.decryptErr
	}
	// Reverse the "encryption"
	decrypted := make([]byte, len(ciphertext))
	for i, b := range ciphertext {
		decrypted[len(ciphertext)-1-i] = b
	}
	return decrypted, nil
}

func TestService_GetStatus_NoConnection(t *testing.T) {
	repo := newMockRepository()
	crypto := &mockCrypto{}
	svc := NewService(repo, crypto)

	ctx := context.Background()
	status, err := svc.GetStatus(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.IsConfigured {
		t.Error("expected IsConfigured to be false")
	}
	if status.LastTestStatus != StatusUnknown {
		t.Errorf("expected status Unknown, got %s", status.LastTestStatus)
	}
}

func TestService_GetStatus_WithConnection(t *testing.T) {
	repo := newMockRepository()
	crypto := &mockCrypto{}
	svc := NewService(repo, crypto)

	// Set up active connection
	testTime := time.Now()
	repo.activeConn = &Connection{
		ID:                      uuid.New(),
		InstanceURL:             "https://test.service-now.com",
		AuthMethod:              AuthMethodBasic,
		IsActive:                true,
		LastTestAt:              &testTime,
		LastTestStatus:          StatusSuccess,
		LastTestInstanceVersion: "Tokyo",
	}

	ctx := context.Background()
	status, err := svc.GetStatus(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !status.IsConfigured {
		t.Error("expected IsConfigured to be true")
	}
	if status.InstanceURL != "https://test.service-now.com" {
		t.Errorf("unexpected instance URL: %s", status.InstanceURL)
	}
	if status.AuthMethod != AuthMethodBasic {
		t.Errorf("unexpected auth method: %s", status.AuthMethod)
	}
	if status.LastTestStatus != StatusSuccess {
		t.Errorf("expected status Success, got %s", status.LastTestStatus)
	}
}

func TestService_SaveConfig_BasicAuth(t *testing.T) {
	repo := newMockRepository()
	crypto := &mockCrypto{}
	svc := NewService(repo, crypto)

	input := &ConfigInput{
		InstanceURL: "https://test.service-now.com",
		AuthMethod:  AuthMethodBasic,
		Username:    "admin",
		Password:    "secret123",
	}

	ctx := context.Background()
	userID := uuid.New()
	conn, err := svc.SaveConfig(ctx, input, &userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conn.InstanceURL != input.InstanceURL {
		t.Errorf("unexpected instance URL: %s", conn.InstanceURL)
	}
	if conn.AuthMethod != AuthMethodBasic {
		t.Errorf("unexpected auth method: %s", conn.AuthMethod)
	}
	if conn.Username != "admin" {
		t.Errorf("unexpected username: %s", conn.Username)
	}
	if len(conn.PasswordEncrypted) == 0 {
		t.Error("password should be encrypted")
	}
	if len(conn.PasswordNonce) == 0 {
		t.Error("password nonce should be set")
	}
	if !conn.IsActive {
		t.Error("connection should be active")
	}
	if conn.LastTestStatus != StatusPending {
		t.Errorf("expected pending status, got %s", conn.LastTestStatus)
	}
}

func TestService_SaveConfig_OAuth(t *testing.T) {
	repo := newMockRepository()
	crypto := &mockCrypto{}
	svc := NewService(repo, crypto)

	input := &ConfigInput{
		InstanceURL:       "https://test.service-now.com",
		AuthMethod:        AuthMethodOAuth,
		OAuthClientID:     "client123",
		OAuthClientSecret: "secret456",
		OAuthTokenURL:     "https://test.service-now.com/oauth_token.do",
	}

	ctx := context.Background()
	conn, err := svc.SaveConfig(ctx, input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if conn.AuthMethod != AuthMethodOAuth {
		t.Errorf("unexpected auth method: %s", conn.AuthMethod)
	}
	if conn.OAuthClientID != "client123" {
		t.Errorf("unexpected client ID: %s", conn.OAuthClientID)
	}
	if len(conn.OAuthClientSecretEncrypted) == 0 {
		t.Error("client secret should be encrypted")
	}
	if conn.OAuthTokenURL != input.OAuthTokenURL {
		t.Errorf("unexpected token URL: %s", conn.OAuthTokenURL)
	}
}

func TestService_SaveConfig_ValidationErrors(t *testing.T) {
	repo := newMockRepository()
	crypto := &mockCrypto{}
	svc := NewService(repo, crypto)
	ctx := context.Background()

	tests := []struct {
		name    string
		input   *ConfigInput
		wantErr error
	}{
		{
			name:    "missing instance URL",
			input:   &ConfigInput{AuthMethod: AuthMethodBasic, Username: "admin", Password: "pass"},
			wantErr: ErrInstanceURLRequired,
		},
		{
			name:    "missing auth method",
			input:   &ConfigInput{InstanceURL: "https://test.com"},
			wantErr: ErrAuthMethodRequired,
		},
		{
			name:    "invalid auth method",
			input:   &ConfigInput{InstanceURL: "https://test.com", AuthMethod: "invalid"},
			wantErr: ErrInvalidAuthMethod,
		},
		{
			name:    "basic auth missing username",
			input:   &ConfigInput{InstanceURL: "https://test.com", AuthMethod: AuthMethodBasic, Password: "pass"},
			wantErr: ErrUsernameRequired,
		},
		{
			name:    "basic auth missing password",
			input:   &ConfigInput{InstanceURL: "https://test.com", AuthMethod: AuthMethodBasic, Username: "admin"},
			wantErr: ErrPasswordRequired,
		},
		{
			name:    "oauth missing client ID",
			input:   &ConfigInput{InstanceURL: "https://test.com", AuthMethod: AuthMethodOAuth, OAuthClientSecret: "secret", OAuthTokenURL: "https://test.com/oauth"},
			wantErr: ErrClientIDRequired,
		},
		{
			name:    "oauth missing client secret",
			input:   &ConfigInput{InstanceURL: "https://test.com", AuthMethod: AuthMethodOAuth, OAuthClientID: "client", OAuthTokenURL: "https://test.com/oauth"},
			wantErr: ErrClientSecretRequired,
		},
		{
			name:    "oauth missing token URL",
			input:   &ConfigInput{InstanceURL: "https://test.com", AuthMethod: AuthMethodOAuth, OAuthClientID: "client", OAuthClientSecret: "secret"},
			wantErr: ErrTokenURLRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.SaveConfig(ctx, tt.input, nil)
			if err != tt.wantErr {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestService_DeleteConnection(t *testing.T) {
	repo := newMockRepository()
	crypto := &mockCrypto{}
	svc := NewService(repo, crypto)

	// Set up active connection
	connID := uuid.New()
	repo.activeConn = &Connection{
		ID:       connID,
		IsActive: true,
	}
	repo.conns[connID] = repo.activeConn

	ctx := context.Background()
	err := svc.DeleteConnection(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify connection is deleted
	_, err = repo.GetActive(ctx)
	if err != ErrConnectionNotFound {
		t.Error("expected connection to be deleted")
	}
}

func TestService_DeleteConnection_NoConnection(t *testing.T) {
	repo := newMockRepository()
	crypto := &mockCrypto{}
	svc := NewService(repo, crypto)

	ctx := context.Background()
	err := svc.DeleteConnection(ctx)
	if err != nil {
		t.Errorf("expected no error when deleting non-existent connection, got %v", err)
	}
}

func TestConfigInput_Validate(t *testing.T) {
	validBasic := &ConfigInput{
		InstanceURL: "https://test.service-now.com",
		AuthMethod:  AuthMethodBasic,
		Username:    "admin",
		Password:    "password",
	}

	if err := validBasic.Validate(); err != nil {
		t.Errorf("valid basic config should not have error: %v", err)
	}

	validOAuth := &ConfigInput{
		InstanceURL:       "https://test.service-now.com",
		AuthMethod:        AuthMethodOAuth,
		OAuthClientID:     "client",
		OAuthClientSecret: "secret",
		OAuthTokenURL:     "https://test.service-now.com/oauth_token.do",
	}

	if err := validOAuth.Validate(); err != nil {
		t.Errorf("valid oauth config should not have error: %v", err)
	}
}
