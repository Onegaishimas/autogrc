# Technical Design Document: ServiceNow GRC Connection (F1)

**Document Version:** 1.0
**Date:** 2026-01-27
**Feature ID:** F1
**Related PRD:** 001_FPRD|ServiceNow_Connection.md

---

## 1. Executive Summary

This TDD defines the technical architecture for establishing and managing the connection between ControlCRUD and ServiceNow GRC instances. The feature provides secure credential storage, connection testing, and status monitoring - serving as the foundation for all sync operations (F2-F4).

**Key Technical Decisions:**
- AES-256-GCM encryption for credential storage
- Connection pooling and retry logic for resilience
- Status caching with TTL for performance
- RESTful API following project conventions

---

## 2. System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        FRONTEND                                  │
│  ┌─────────────────┐  ┌─────────────────┐                       │
│  │ Connection      │  │ Status          │                       │
│  │ Settings Page   │  │ Indicator       │                       │
│  └────────┬────────┘  └────────┬────────┘                       │
│           │                    │                                 │
└───────────┼────────────────────┼─────────────────────────────────┘
            │ REST API           │ REST API
            ▼                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                        BACKEND (Go)                             │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ Connection      │  │ ServiceNow      │  │ Crypto          │ │
│  │ Handler         │  │ Client          │  │ Service         │ │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘ │
│           │                    │                    │           │
│           ▼                    ▼                    ▼           │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                    Connection Service                       ││
│  │  - Config management                                        ││
│  │  - Connection testing                                       ││
│  │  - Status monitoring                                        ││
│  └─────────────────────────────────────────────────────────────┘│
└───────────┬─────────────────────────────────────────────────────┘
            │
            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      PostgreSQL                                  │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ servicenow_connections (encrypted credentials)              ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
            │
            ▼ HTTPS
┌─────────────────────────────────────────────────────────────────┐
│                    ServiceNow Instance                           │
│                    (REST API)                                    │
└─────────────────────────────────────────────────────────────────┘
```

### Component Relationships

| Component | Responsibility | Dependencies |
|-----------|---------------|--------------|
| Connection Handler | HTTP endpoint handling | Connection Service |
| Connection Service | Business logic | ServiceNow Client, Crypto Service, DB |
| ServiceNow Client | External API calls | net/http |
| Crypto Service | Encryption/decryption | crypto/aes |

---

## 3. Technical Stack

### Backend Components

| Component | Technology | Justification |
|-----------|------------|---------------|
| HTTP Router | Chi | Per ADR, lightweight, middleware support |
| Database Access | sqlc + pgx | Type-safe SQL, native PostgreSQL driver |
| HTTP Client | net/http + retryablehttp | Stdlib with retry capability |
| Encryption | crypto/aes | Go stdlib, AES-256-GCM |
| Logging | zerolog | Structured JSON logging per ADR |
| Validation | go-playground/validator | Request validation |

### Frontend Components

| Component | Technology | Justification |
|-----------|------------|---------------|
| Framework | React 18 + TypeScript | Per ADR |
| State | TanStack Query | Server state management |
| Forms | React Hook Form + Zod | Form handling with validation |
| UI | Shadcn/ui | Per ADR, accessible components |

### Dependencies

```go
// go.mod additions
require (
    github.com/hashicorp/go-retryablehttp v0.7.5
    github.com/go-playground/validator/v10 v10.18.0
)
```

---

## 4. Data Design

### Database Schema

```sql
-- migrations/20260127_001_create_servicenow_connections.up.sql

CREATE TYPE auth_method AS ENUM ('basic', 'oauth');
CREATE TYPE connection_status AS ENUM ('success', 'failure', 'pending', 'unknown');

CREATE TABLE servicenow_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_url VARCHAR(255) NOT NULL,
    auth_method auth_method NOT NULL DEFAULT 'basic',

    -- Basic Auth (encrypted)
    username VARCHAR(255),
    password_encrypted BYTEA,
    password_nonce BYTEA,

    -- OAuth (encrypted)
    oauth_client_id VARCHAR(255),
    oauth_client_secret_encrypted BYTEA,
    oauth_client_secret_nonce BYTEA,
    oauth_token_url VARCHAR(255),

    -- Status tracking
    is_active BOOLEAN DEFAULT true,
    last_test_at TIMESTAMPTZ,
    last_test_status connection_status DEFAULT 'unknown',
    last_test_message TEXT,
    last_test_instance_version VARCHAR(50),

    -- Audit
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),

    CONSTRAINT single_active_connection UNIQUE (is_active)
        WHERE is_active = true
);

CREATE INDEX idx_connections_active ON servicenow_connections(is_active)
    WHERE is_active = true;
```

### Data Models (Go)

```go
// internal/domain/connection/models.go

type Connection struct {
    ID                  uuid.UUID       `db:"id"`
    InstanceURL         string          `db:"instance_url"`
    AuthMethod          AuthMethod      `db:"auth_method"`
    Username            *string         `db:"username"`
    LastTestAt          *time.Time      `db:"last_test_at"`
    LastTestStatus      ConnectionStatus `db:"last_test_status"`
    LastTestMessage     *string         `db:"last_test_message"`
    IsActive            bool            `db:"is_active"`
    CreatedAt           time.Time       `db:"created_at"`
    UpdatedAt           time.Time       `db:"updated_at"`
}

type AuthMethod string
const (
    AuthMethodBasic AuthMethod = "basic"
    AuthMethodOAuth AuthMethod = "oauth"
)

type ConnectionStatus string
const (
    StatusSuccess ConnectionStatus = "success"
    StatusFailure ConnectionStatus = "failure"
    StatusPending ConnectionStatus = "pending"
    StatusUnknown ConnectionStatus = "unknown"
)
```

### Encryption Strategy

```go
// internal/infrastructure/crypto/aes.go

type CryptoService interface {
    Encrypt(plaintext []byte) (ciphertext, nonce []byte, err error)
    Decrypt(ciphertext, nonce []byte) (plaintext []byte, err error)
}

// Implementation uses AES-256-GCM
// Key derived from environment variable via PBKDF2
// Unique nonce per encryption operation
```

---

## 5. API Design

### Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| GET | /api/v1/connection/status | Get connection status | User |
| POST | /api/v1/connection/config | Create/update connection | Admin |
| POST | /api/v1/connection/test | Test connection | Admin |
| DELETE | /api/v1/connection | Remove connection | Admin |

### Request/Response Schemas

```go
// internal/api/handlers/connection/schemas.go

type ConfigRequest struct {
    InstanceURL string     `json:"instance_url" validate:"required,url,startswith=https://"`
    AuthMethod  AuthMethod `json:"auth_method" validate:"required,oneof=basic oauth"`
    Username    *string    `json:"username" validate:"required_if=AuthMethod basic"`
    Password    *string    `json:"password" validate:"required_if=AuthMethod basic"`
    OAuthClientID     *string `json:"oauth_client_id" validate:"required_if=AuthMethod oauth"`
    OAuthClientSecret *string `json:"oauth_client_secret" validate:"required_if=AuthMethod oauth"`
}

type StatusResponse struct {
    Status           ConnectionStatus `json:"status"`
    InstanceURL      string           `json:"instance_url,omitempty"`
    LastTestAt       *time.Time       `json:"last_test_at,omitempty"`
    LastTestStatus   ConnectionStatus `json:"last_test_status"`
    InstanceVersion  *string          `json:"instance_version,omitempty"`
}

type TestResponse struct {
    Status      string             `json:"status"`
    Message     string             `json:"message"`
    InstanceInfo *InstanceInfo     `json:"instance_info,omitempty"`
    Permissions  map[string]string `json:"permissions,omitempty"`
}
```

### Error Handling

```go
// Standardized error responses
type ErrorCode string
const (
    ErrInvalidURL        ErrorCode = "INVALID_URL"
    ErrAuthFailed        ErrorCode = "AUTH_FAILED"
    ErrConnectionTimeout ErrorCode = "CONNECTION_TIMEOUT"
    ErrPermissionDenied  ErrorCode = "PERMISSION_DENIED"
    ErrRateLimited       ErrorCode = "RATE_LIMITED"
)
```

---

## 6. Component Architecture

### Backend Package Structure

```
internal/
├── api/
│   └── handlers/
│       └── connection/
│           ├── handler.go      # HTTP handlers
│           ├── schemas.go      # Request/response types
│           └── handler_test.go
├── domain/
│   └── connection/
│       ├── models.go           # Domain models
│       ├── service.go          # Business logic
│       ├── repository.go       # Database interface
│       └── service_test.go
└── infrastructure/
    ├── crypto/
    │   ├── aes.go              # Encryption service
    │   └── aes_test.go
    ├── database/
    │   └── queries/
    │       └── connection.sql  # sqlc queries
    └── servicenow/
        ├── client.go           # ServiceNow API client
        ├── client_test.go
        └── models.go           # API response types
```

### Frontend Component Structure

```
src/
├── features/
│   └── connection/
│       ├── components/
│       │   ├── ConnectionSettings.tsx
│       │   ├── ConnectionForm.tsx
│       │   ├── ConnectionTest.tsx
│       │   └── StatusIndicator.tsx
│       ├── hooks/
│       │   ├── useConnection.ts
│       │   └── useConnectionTest.ts
│       ├── api/
│       │   └── connectionApi.ts
│       └── types.ts
└── components/
    └── layout/
        └── Header.tsx  # Includes StatusIndicator
```

---

## 7. State Management

### Backend State
- Connection config stored in PostgreSQL
- Status cached in service (refreshed on demand)
- No in-memory session state (stateless)

### Frontend State

```typescript
// TanStack Query for server state
const { data: status } = useQuery({
  queryKey: ['connection', 'status'],
  queryFn: connectionApi.getStatus,
  staleTime: 60_000, // 1 minute
  refetchOnWindowFocus: true,
});

// Zustand for UI state (none needed for this feature)
```

---

## 8. Security Considerations

### Credential Protection
1. Credentials encrypted with AES-256-GCM before storage
2. Encryption key from environment variable (not in code)
3. Unique nonce per encryption (stored alongside ciphertext)
4. Password never logged or returned in responses
5. Password masked in UI after entry

### API Security
1. Admin role required for config/test endpoints
2. CSRF protection via SameSite cookies
3. Rate limiting on test endpoint (5 per minute)
4. Input validation on all requests

### Audit
1. All config changes logged with user attribution
2. Test results logged (success/failure only, no credentials)

---

## 9. Performance & Scalability

### Performance Targets
| Operation | Target | Approach |
|-----------|--------|----------|
| Status check | < 100ms | Cached status, DB query |
| Connection test | < 10s | HTTP timeout, no retry |
| Config save | < 500ms | Single DB transaction |

### Connection Pooling
- Use pgxpool for database connections (10-20 pool size)
- HTTP client with connection reuse for ServiceNow

### Caching
- Status cached in TanStack Query (1 minute staleTime)
- No server-side caching needed (single connection)

---

## 10. Testing Strategy

### Unit Tests

```go
// internal/domain/connection/service_test.go
func TestConnectionService_SaveConfig(t *testing.T) {
    // Test valid config saves successfully
    // Test invalid URL rejected
    // Test credentials encrypted
}

func TestConnectionService_TestConnection(t *testing.T) {
    // Test successful connection
    // Test auth failure
    // Test timeout handling
}
```

### Integration Tests

```go
// internal/api/handlers/connection/handler_test.go
func TestConnectionHandler_PostConfig(t *testing.T) {
    // Test API request/response
    // Test validation errors
    // Test authorization
}
```

### Mock Strategy
- Mock ServiceNow client for unit tests
- Use httptest for handler tests
- Testcontainers for database integration tests

---

## 11. Deployment & DevOps

### Environment Configuration

```bash
# Required environment variables
ENCRYPTION_KEY=base64-encoded-32-byte-key
DATABASE_URL=postgres://user:pass@localhost/controlcrud

# Optional
SN_CONNECTION_TIMEOUT=10s
SN_RETRY_MAX=3
```

### Health Check
- Include connection status in `/health` endpoint
- Degraded (not failed) if ServiceNow unreachable

---

## 12. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Credentials exposed in logs | Low | High | Never log credentials, code review |
| ServiceNow API changes | Low | Medium | Version-specific client, monitoring |
| Encryption key loss | Low | High | Key backup procedures, rotation plan |
| Connection timeout issues | Medium | Low | Configurable timeout, retry logic |

---

## 13. Development Phases

### Phase 1: Database & Crypto (2 days)
- Database migrations
- Crypto service implementation
- Unit tests for encryption

### Phase 2: Backend API (3 days)
- Connection service
- ServiceNow client (basic connectivity)
- API handlers
- Integration tests

### Phase 3: Frontend (2 days)
- Settings form component
- Test connection UI
- Status indicator
- Integration with header

### Phase 4: Polish & Testing (1 day)
- E2E tests
- Error handling refinement
- Documentation

**Estimated Total: 8 days**

---

## 14. Open Technical Questions

1. **Key Rotation:** How should we handle encryption key rotation?
2. **OAuth Token Refresh:** What's the token lifetime, do we need refresh logic?
3. **Multi-tenancy Future:** Should schema support multiple connections now?
