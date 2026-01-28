package servicenow

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewSNClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *ClientConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig("https://instance.service-now.com"),
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "empty instance URL",
			config: &ClientConfig{
				InstanceURL: "",
				Timeout:     10 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewSNClient(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if client == nil {
					t.Error("expected client, got nil")
				}
			}
		})
	}
}

func TestBasicAuthProvider(t *testing.T) {
	auth := &BasicAuthProvider{
		Username: "admin",
		Password: "secret",
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	err := auth.ApplyAuth(req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	username, password, ok := req.BasicAuth()
	if !ok {
		t.Error("basic auth not set")
	}
	if username != "admin" {
		t.Errorf("expected username 'admin', got '%s'", username)
	}
	if password != "secret" {
		t.Errorf("expected password 'secret', got '%s'", password)
	}

	if auth.Type() != "basic" {
		t.Errorf("expected type 'basic', got '%s'", auth.Type())
	}
}

func TestOAuthProvider(t *testing.T) {
	auth := &OAuthProvider{
		ClientID:     "client123",
		ClientSecret: "secret456",
		TokenURL:     "https://instance.service-now.com/oauth_token.do",
		accessToken:  "test-token",
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
	err := auth.ApplyAuth(req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	authHeader := req.Header.Get("Authorization")
	if authHeader != "Bearer test-token" {
		t.Errorf("expected 'Bearer test-token', got '%s'", authHeader)
	}

	if auth.Type() != "oauth" {
		t.Errorf("expected type 'oauth', got '%s'", auth.Type())
	}
}

func TestTestConnection_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/now/table/sys_properties" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Check auth header
		username, password, ok := r.BasicAuth()
		if !ok || username != "admin" || password != "password" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Return mock response
		response := TableAPIResponse[SysProperty]{
			Result: []SysProperty{
				{Name: "glide.product.version", Value: "Tokyo"},
				{Name: "glide.buildtag", Value: "glide-tokyo-12-15-2022__patch0-hotfix1"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client, err := NewSNClient(&ClientConfig{
		InstanceURL: server.URL,
		Timeout:     5 * time.Second,
		MaxRetries:  1,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Set auth
	client.SetAuth(&BasicAuthProvider{
		Username: "admin",
		Password: "password",
	})

	// Test connection
	ctx := context.Background()
	result, err := client.TestConnection(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Errorf("expected success, got failure: %s", result.ErrorMessage)
	}
	if result.InstanceInfo.Version != "Tokyo" {
		t.Errorf("expected version 'Tokyo', got '%s'", result.InstanceInfo.Version)
	}
	if result.InstanceInfo.BuildTag != "glide-tokyo-12-15-2022__patch0-hotfix1" {
		t.Errorf("unexpected build tag: %s", result.InstanceInfo.BuildTag)
	}
	if result.ResponseTimeMs < 0 {
		t.Error("expected non-negative response time")
	}
}

func TestTestConnection_AuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"User Not Authenticated"}}`))
	}))
	defer server.Close()

	client, _ := NewSNClient(&ClientConfig{
		InstanceURL: server.URL,
		Timeout:     5 * time.Second,
		MaxRetries:  0,
	})
	client.SetAuth(&BasicAuthProvider{
		Username: "wrong",
		Password: "creds",
	})

	ctx := context.Background()
	result, err := client.TestConnection(ctx)

	if err != ErrAuthFailed {
		t.Errorf("expected ErrAuthFailed, got %v", err)
	}
	if result.Success {
		t.Error("expected failure")
	}
	if result.ErrorMessage == "" {
		t.Error("expected error message")
	}
}

func TestTestConnection_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, _ := NewSNClient(&ClientConfig{
		InstanceURL: server.URL,
		Timeout:     5 * time.Second,
		MaxRetries:  0,
	})
	client.SetAuth(&BasicAuthProvider{Username: "admin", Password: "pass"})

	ctx := context.Background()
	result, err := client.TestConnection(ctx)

	if err != ErrServerError {
		t.Errorf("expected ErrServerError, got %v", err)
	}
	if result.Success {
		t.Error("expected failure")
	}
}

func TestTestConnection_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := NewSNClient(&ClientConfig{
		InstanceURL: server.URL,
		Timeout:     50 * time.Millisecond, // Very short timeout
		MaxRetries:  0,
	})
	client.SetAuth(&BasicAuthProvider{Username: "admin", Password: "pass"})

	ctx := context.Background()
	result, err := client.TestConnection(ctx)

	if err == nil {
		t.Error("expected timeout error")
	}
	if result.Success {
		t.Error("expected failure due to timeout")
	}
}

func TestTestConnection_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client, _ := NewSNClient(&ClientConfig{
		InstanceURL: server.URL,
		Timeout:     5 * time.Second,
		MaxRetries:  0,
	})
	client.SetAuth(&BasicAuthProvider{Username: "admin", Password: "pass"})

	ctx := context.Background()
	result, err := client.TestConnection(ctx)

	if err != ErrRateLimited {
		t.Errorf("expected ErrRateLimited, got %v", err)
	}
	if result.Success {
		t.Error("expected failure")
	}
}

func TestTestConnection_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client, _ := NewSNClient(&ClientConfig{
		InstanceURL: server.URL,
		Timeout:     5 * time.Second,
		MaxRetries:  0,
	})
	client.SetAuth(&BasicAuthProvider{Username: "admin", Password: "pass"})

	ctx := context.Background()
	result, err := client.TestConnection(ctx)

	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if result.Success {
		t.Error("expected failure")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig("https://instance.service-now.com")

	if config.InstanceURL != "https://instance.service-now.com" {
		t.Errorf("unexpected instance URL: %s", config.InstanceURL)
	}
	if config.Timeout != 10*time.Second {
		t.Errorf("expected 10s timeout, got %v", config.Timeout)
	}
	if config.MaxRetries != 3 {
		t.Errorf("expected 3 retries, got %d", config.MaxRetries)
	}
}
