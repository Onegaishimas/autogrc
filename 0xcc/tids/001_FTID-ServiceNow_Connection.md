# Technical Implementation Document: ServiceNow GRC Connection (F1)

**Document Version:** 1.0
**Date:** 2026-01-27
**Feature ID:** F1
**Related PRD:** 001_FPRD|ServiceNow_Connection.md
**Related TDD:** 001_FTDD|ServiceNow_Connection.md

---

## 1. Implementation Overview

### Summary
This TID provides specific implementation guidance for the ServiceNow GRC Connection feature, establishing secure credential storage, connection testing, and status monitoring capabilities.

### Key Implementation Principles
- **Security First:** AES-256-GCM encryption for all credentials before storage
- **Explicit Error Handling:** Wrap all errors with context using `fmt.Errorf`
- **Interface-Based Design:** Define interfaces for testability (CryptoService, SNClient)
- **Stateless Backend:** No in-memory session state; rely on database and short-lived caches

### Integration Points
- Database: PostgreSQL via sqlc + pgx
- External: ServiceNow REST API via net/http + retryablehttp
- Frontend: React components consuming REST API via TanStack Query

---

## 2. File Structure and Organization

### Backend Files to Create

```
backend/
├── cmd/server/main.go                    # Add connection routes (modify)
├── internal/
│   ├── api/
│   │   └── handlers/
│   │       └── connection/
│   │           ├── handler.go            # HTTP handlers (create)
│   │           ├── handler_test.go       # Handler tests (create)
│   │           └── schemas.go            # Request/response DTOs (create)
│   ├── domain/
│   │   └── connection/
│   │       ├── models.go                 # Domain models (create)
│   │       ├── service.go                # Business logic (create)
│   │       ├── service_test.go           # Service tests (create)
│   │       └── repository.go             # DB interface (create)
│   └── infrastructure/
│       ├── crypto/
│       │   ├── aes.go                    # AES-256-GCM impl (create)
│       │   └── aes_test.go               # Crypto tests (create)
│       ├── database/
│       │   └── queries/
│       │       └── connection.sql        # sqlc queries (create)
│       └── servicenow/
│           ├── client.go                 # SN API client (create)
│           ├── client_test.go            # Client tests (create)
│           └── models.go                 # API response types (create)
└── migrations/
    └── 20260127_001_create_servicenow_connections.sql  # Schema (create)
```

### Frontend Files to Create

```
frontend/src/
├── features/
│   └── connection/
│       ├── components/
│       │   ├── ConnectionSettings.tsx    # Main settings page (create)
│       │   ├── ConnectionForm.tsx        # Config form (create)
│       │   ├── ConnectionTest.tsx        # Test button/results (create)
│       │   └── StatusIndicator.tsx       # Header status badge (create)
│       ├── hooks/
│       │   ├── useConnection.ts          # Query/mutation hooks (create)
│       │   └── useConnectionTest.ts      # Test mutation hook (create)
│       ├── api/
│       │   └── connectionApi.ts          # API client functions (create)
│       └── types.ts                      # TypeScript types (create)
└── components/
    └── layout/
        └── Header.tsx                    # Add StatusIndicator (modify)
```

### Naming Conventions
- Go: `snake_case` for files, `PascalCase` for exported types
- TypeScript: `PascalCase` for components, `camelCase` for functions/hooks
- API routes: `/api/v1/connection/{action}`

---

## 3. Component Implementation Hints

### Go Service Layer Pattern

```go
// internal/domain/connection/service.go
type Service struct {
    repo    Repository
    crypto  crypto.CryptoService
    snClient servicenow.Client
    logger  zerolog.Logger
}

func NewService(repo Repository, crypto crypto.CryptoService, snClient servicenow.Client, logger zerolog.Logger) *Service {
    return &Service{repo: repo, crypto: crypto, snClient: snClient, logger: logger}
}
```

**Key Patterns:**
- Constructor injection for all dependencies
- Logger passed as dependency for testing
- Methods return `(result, error)` - never panic
- Use `context.Context` as first parameter on all public methods

### React Component Hierarchy

```
ConnectionSettings (page container)
├── ConnectionForm (form with validation)
│   ├── Input (instance URL)
│   ├── RadioGroup (auth method)
│   ├── Input (username) - conditional
│   ├── Input (password) - conditional
│   ├── Input (OAuth client ID) - conditional
│   └── Input (OAuth client secret) - conditional
├── ConnectionTest (test button + results)
│   └── Alert (success/failure display)
└── StatusIndicator (connection status badge)
```

**Key Patterns:**
- Use React Hook Form for form state
- Zod schemas for validation (shared with API types)
- Shadcn/ui components for all UI elements
- TanStack Query mutations for form submission

---

## 4. Database Implementation Approach

### Migration Strategy

```sql
-- Use IF NOT EXISTS for enum creation
DO $$ BEGIN
    CREATE TYPE auth_method AS ENUM ('basic', 'oauth');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;
```

**Key Hints:**
- Single migration file for initial schema
- Use UUID primary key with `gen_random_uuid()`
- Partial unique index for single active connection: `WHERE is_active = true`
- Store nonce alongside encrypted data (never reuse nonces)

### sqlc Query Patterns

```sql
-- name: GetActiveConnection :one
SELECT * FROM servicenow_connections
WHERE is_active = true
LIMIT 1;

-- name: UpsertConnection :one
INSERT INTO servicenow_connections (
    instance_url, auth_method, username, password_encrypted, password_nonce,
    is_active, created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, true, $6, $6)
ON CONFLICT (is_active) WHERE is_active = true
DO UPDATE SET
    instance_url = EXCLUDED.instance_url,
    auth_method = EXCLUDED.auth_method,
    username = EXCLUDED.username,
    password_encrypted = EXCLUDED.password_encrypted,
    password_nonce = EXCLUDED.password_nonce,
    updated_at = NOW(),
    updated_by = EXCLUDED.updated_by
RETURNING *;
```

**Key Hints:**
- Use UPSERT for single-connection model
- Return full row after mutations for cache invalidation
- Use `TIMESTAMPTZ` for all timestamps (timezone-aware)

---

## 5. API Implementation Strategy

### Handler Structure

```go
// internal/api/handlers/connection/handler.go
type Handler struct {
    service *connection.Service
    logger  zerolog.Logger
}

func (h *Handler) RegisterRoutes(r chi.Router) {
    r.Route("/api/v1/connection", func(r chi.Router) {
        r.Get("/status", h.GetStatus)
        r.Post("/config", h.SaveConfig)  // Admin only
        r.Post("/test", h.TestConnection) // Admin only
        r.Delete("/", h.DeleteConnection) // Admin only
    })
}
```

### Request Validation Pattern

```go
func (h *Handler) SaveConfig(w http.ResponseWriter, r *http.Request) {
    var req ConfigRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
        return
    }

    if err := validator.Validate(req); err != nil {
        h.respondValidationError(w, err)
        return
    }

    // Process request...
}
```

**Key Hints:**
- Validate URL starts with `https://`
- Validate required fields based on auth_method
- Return structured error responses with codes
- Log request IDs for traceability

### Response Wrapper Pattern

```go
type Response struct {
    Data  interface{} `json:"data,omitempty"`
    Error *ErrorBody  `json:"error,omitempty"`
}

type ErrorBody struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

---

## 6. Frontend Implementation Approach

### TanStack Query Setup

```typescript
// features/connection/hooks/useConnection.ts
export function useConnectionStatus() {
  return useQuery({
    queryKey: ['connection', 'status'],
    queryFn: connectionApi.getStatus,
    staleTime: 60_000, // 1 minute
    refetchOnWindowFocus: true,
  });
}

export function useSaveConnection() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: connectionApi.saveConfig,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['connection'] });
    },
  });
}
```

### Form Implementation Pattern

```typescript
// features/connection/components/ConnectionForm.tsx
const formSchema = z.object({
  instanceUrl: z.string().url().startsWith('https://'),
  authMethod: z.enum(['basic', 'oauth']),
  username: z.string().optional(),
  password: z.string().optional(),
}).refine(/* conditional validation */);

export function ConnectionForm() {
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
  });
  const mutation = useSaveConnection();

  const onSubmit = (data: z.infer<typeof formSchema>) => {
    mutation.mutate(data);
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)}>
        {/* Form fields */}
      </form>
    </Form>
  );
}
```

**Key Hints:**
- Use `zodResolver` to connect Zod schemas to React Hook Form
- Show/hide fields based on `authMethod` selection
- Mask password fields after initial entry (show dots, not value)
- Disable form during mutation pending state

### Status Indicator Pattern

```typescript
// features/connection/components/StatusIndicator.tsx
export function StatusIndicator() {
  const { data, isLoading } = useConnectionStatus();

  if (isLoading) return <Badge variant="outline">...</Badge>;
  if (!data) return <Badge variant="destructive">Not Configured</Badge>;

  const variants = {
    success: 'default',
    failure: 'destructive',
    pending: 'secondary',
    unknown: 'outline',
  } as const;

  return <Badge variant={variants[data.status]}>{data.status}</Badge>;
}
```

---

## 7. Business Logic Implementation Hints

### Encryption Service Implementation

```go
// internal/infrastructure/crypto/aes.go
type AESCryptoService struct {
    key []byte // 32 bytes for AES-256
}

func NewAESCryptoService(keyBase64 string) (*AESCryptoService, error) {
    key, err := base64.StdEncoding.DecodeString(keyBase64)
    if err != nil {
        return nil, fmt.Errorf("failed to decode encryption key: %w", err)
    }
    if len(key) != 32 {
        return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
    }
    return &AESCryptoService{key: key}, nil
}

func (s *AESCryptoService) Encrypt(plaintext []byte) (ciphertext, nonce []byte, err error) {
    block, err := aes.NewCipher(s.key)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to create cipher: %w", err)
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to create GCM: %w", err)
    }

    nonce = make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
    }

    ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
    return ciphertext, nonce, nil
}
```

**Key Hints:**
- Generate new nonce for every encryption operation
- Store nonce in separate column (not prepended to ciphertext)
- Key loaded from `ENCRYPTION_KEY` environment variable
- Use `crypto/rand` for nonce generation, never `math/rand`

### ServiceNow Client Implementation

```go
// internal/infrastructure/servicenow/client.go
type Client struct {
    baseURL    string
    httpClient *retryablehttp.Client
    auth       AuthProvider
}

func (c *Client) TestConnection(ctx context.Context) (*InstanceInfo, error) {
    req, err := retryablehttp.NewRequestWithContext(ctx, "GET",
        c.baseURL+"/api/now/table/sys_properties?sysparm_limit=1", nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    c.auth.ApplyAuth(req)
    req.Header.Set("Accept", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("connection failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == 401 {
        return nil, ErrAuthFailed
    }
    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    // Parse and return instance info...
}
```

**Key Hints:**
- Use `retryablehttp` for automatic retry with exponential backoff
- Set 10-second timeout for test connections
- Test by querying `sys_properties` table (requires minimal permissions)
- Return meaningful error types (ErrAuthFailed, ErrTimeout, etc.)

---

## 8. Testing Implementation Approach

### Unit Test Pattern (Service)

```go
// internal/domain/connection/service_test.go
func TestService_SaveConfig(t *testing.T) {
    tests := []struct {
        name    string
        input   ConfigInput
        setup   func(*MockRepo, *MockCrypto)
        wantErr bool
    }{
        {
            name: "valid basic auth config saves successfully",
            input: ConfigInput{
                InstanceURL: "https://test.service-now.com",
                AuthMethod:  AuthMethodBasic,
                Username:    ptr("admin"),
                Password:    ptr("secret"),
            },
            setup: func(repo *MockRepo, crypto *MockCrypto) {
                crypto.EXPECT().Encrypt([]byte("secret")).Return([]byte("enc"), []byte("nonce"), nil)
                repo.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(&Connection{}, nil)
            },
            wantErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks and run test...
        })
    }
}
```

### Integration Test Pattern (Handler)

```go
// internal/api/handlers/connection/handler_test.go
func TestHandler_SaveConfig(t *testing.T) {
    // Use httptest.NewServer
    // Use real service with mock dependencies
    // Test full request/response cycle
}
```

### Frontend Test Pattern

```typescript
// features/connection/components/ConnectionForm.test.tsx
describe('ConnectionForm', () => {
  it('shows username/password fields for basic auth', async () => {
    render(<ConnectionForm />);

    await userEvent.click(screen.getByLabelText('Basic Authentication'));

    expect(screen.getByLabelText('Username')).toBeInTheDocument();
    expect(screen.getByLabelText('Password')).toBeInTheDocument();
  });

  it('validates URL is HTTPS', async () => {
    render(<ConnectionForm />);

    await userEvent.type(screen.getByLabelText('Instance URL'), 'http://test.com');
    await userEvent.click(screen.getByRole('button', { name: 'Save' }));

    expect(screen.getByText(/must start with https/i)).toBeInTheDocument();
  });
});
```

**Key Hints:**
- Mock external dependencies at service boundaries
- Use table-driven tests in Go
- Test validation errors explicitly
- Test conditional field display in frontend

---

## 9. Configuration and Environment Strategy

### Environment Variables

```bash
# Required
ENCRYPTION_KEY=base64-encoded-32-byte-key
DATABASE_URL=postgres://user:pass@localhost:5432/controlcrud

# Optional with defaults
SN_CONNECTION_TIMEOUT=10s
SN_RETRY_MAX=3
SN_RETRY_WAIT_MIN=1s
SN_RETRY_WAIT_MAX=30s
```

### Configuration Loading Pattern

```go
// internal/config/config.go
type Config struct {
    EncryptionKey string        `env:"ENCRYPTION_KEY,required"`
    DatabaseURL   string        `env:"DATABASE_URL,required"`
    SNTimeout     time.Duration `env:"SN_CONNECTION_TIMEOUT" envDefault:"10s"`
    SNRetryMax    int           `env:"SN_RETRY_MAX" envDefault:"3"`
}

func Load() (*Config, error) {
    var cfg Config
    if err := env.Parse(&cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }
    return &cfg, nil
}
```

**Key Hints:**
- Use `github.com/caarlos0/env` for environment parsing
- Fail fast on missing required variables
- Log configuration (without secrets) at startup

---

## 10. Integration Strategy

### Route Registration

```go
// cmd/server/main.go
func main() {
    // Load config...

    // Initialize dependencies
    cryptoSvc, err := crypto.NewAESCryptoService(cfg.EncryptionKey)
    if err != nil {
        log.Fatal().Err(err).Msg("failed to initialize crypto service")
    }

    snClient := servicenow.NewClient(cfg.SNTimeout, cfg.SNRetryMax)
    connRepo := database.NewConnectionRepository(db)
    connSvc := connection.NewService(connRepo, cryptoSvc, snClient, logger)
    connHandler := handlers.NewConnectionHandler(connSvc, logger)

    // Register routes
    r := chi.NewRouter()
    connHandler.RegisterRoutes(r)

    // Start server...
}
```

### Header Integration (Frontend)

```typescript
// components/layout/Header.tsx
import { StatusIndicator } from '@/features/connection/components/StatusIndicator';

export function Header() {
  return (
    <header className="flex items-center justify-between px-4 py-2 border-b">
      <Logo />
      <nav>{/* Navigation items */}</nav>
      <div className="flex items-center gap-4">
        <StatusIndicator />
        <UserMenu />
      </div>
    </header>
  );
}
```

---

## 11. Utilities and Helpers Design

### Error Utilities

```go
// internal/pkg/errors/errors.go
type AppError struct {
    Code    string
    Message string
    Err     error
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

var (
    ErrAuthFailed        = &AppError{Code: "AUTH_FAILED", Message: "Authentication failed"}
    ErrConnectionTimeout = &AppError{Code: "CONNECTION_TIMEOUT", Message: "Connection timed out"}
    ErrInvalidURL        = &AppError{Code: "INVALID_URL", Message: "Invalid ServiceNow URL"}
)
```

### Response Helpers

```go
// internal/api/response/response.go
func JSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(Response{Data: data})
}

func Error(w http.ResponseWriter, status int, code, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(Response{Error: &ErrorBody{Code: code, Message: message}})
}
```

---

## 12. Error Handling and Logging Strategy

### Structured Logging Pattern

```go
// Log with context
logger.Info().
    Str("instance_url", config.InstanceURL).
    Str("auth_method", string(config.AuthMethod)).
    Msg("testing ServiceNow connection")

// Log errors with stack
logger.Error().
    Err(err).
    Str("instance_url", config.InstanceURL).
    Msg("connection test failed")
```

### User-Facing Error Messages

| Internal Error | User Message |
|---------------|--------------|
| `ErrAuthFailed` | "Unable to authenticate. Please check your credentials." |
| `ErrConnectionTimeout` | "Connection timed out. Please verify the instance URL is accessible." |
| `ErrInvalidURL` | "Invalid URL. Please enter a valid ServiceNow instance URL starting with https://." |

**Key Hints:**
- Never expose internal error details to users
- Log full error context server-side
- Include request ID in error responses for support

---

## 13. Performance Implementation Hints

### Connection Timeout Configuration

```go
httpClient := &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        10,
        MaxIdleConnsPerHost: 5,
        IdleConnTimeout:     30 * time.Second,
    },
}
```

### Database Connection Pool

```go
poolConfig, _ := pgxpool.ParseConfig(cfg.DatabaseURL)
poolConfig.MaxConns = 20
poolConfig.MinConns = 5
poolConfig.MaxConnLifetime = 30 * time.Minute
poolConfig.MaxConnIdleTime = 5 * time.Minute
```

**Key Hints:**
- No server-side caching for single connection model
- Frontend caches status for 1 minute via TanStack Query
- Use connection pooling for database and HTTP clients

---

## 14. Code Quality and Standards

### Required Linting (Go)

```yaml
# .golangci.yml
linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - gosec
    - gofmt
```

### Required Linting (TypeScript)

```json
// .eslintrc
{
  "extends": ["@typescript-eslint/recommended", "prettier"],
  "rules": {
    "@typescript-eslint/no-explicit-any": "error",
    "@typescript-eslint/explicit-function-return-type": "warn"
  }
}
```

### Documentation Requirements
- Go: Document all exported functions and types
- TypeScript: Document complex component props with JSDoc
- API: OpenAPI spec for all endpoints

---

## 15. Security Checklist

- [ ] Credentials encrypted with AES-256-GCM before database storage
- [ ] Encryption key loaded from environment variable
- [ ] Unique nonce generated for each encryption operation
- [ ] Password never logged or returned in API responses
- [ ] Admin role required for config/test/delete endpoints
- [ ] Input validation on all API requests (URL format, required fields)
- [ ] CSRF protection via SameSite cookies
- [ ] Rate limiting on test endpoint (5 requests/minute)
