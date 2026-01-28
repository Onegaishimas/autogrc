# Technical Design Document: Control Package Pull (F2)

**Document Version:** 1.0
**Date:** 2026-01-27
**Feature ID:** F2
**Related PRD:** 002_FPRD|Control_Package_Pull.md

---

## 1. Executive Summary

This TDD defines the technical architecture for pulling control packages from ServiceNow GRC into ControlCRUD. The feature handles system selection, control data import, implementation statement retrieval, and sync state tracking.

**Key Technical Decisions:**
- Background job pattern for long-running pulls
- Pagination handling for large datasets
- Optimistic sync state with conflict detection
- ServiceNow sys_id as primary external reference

---

## 2. System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        FRONTEND                                  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ System          │  │ Pull            │  │ Pull            │ │
│  │ Selector        │  │ Button          │  │ Progress        │ │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘ │
└───────────┼────────────────────┼────────────────────┼───────────┘
            │                    │                    │
            ▼                    ▼                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                        BACKEND (Go)                             │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ System          │  │ Pull            │  │ Pull            │ │
│  │ Handler         │  │ Handler         │  │ Status Handler  │ │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘ │
│           │                    │                    │           │
│           ▼                    ▼                    ▼           │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                      Pull Service                           ││
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        ││
│  │  │ System      │  │ Control     │  │ Sync State  │        ││
│  │  │ Fetcher     │  │ Importer    │  │ Manager     │        ││
│  │  └─────────────┘  └─────────────┘  └─────────────┘        ││
│  └─────────────────────────────────────────────────────────────┘│
│           │                    │                                 │
│           ▼                    ▼                                 │
│  ┌─────────────────┐  ┌─────────────────────────────────────┐  │
│  │ ServiceNow      │  │           PostgreSQL                 │  │
│  │ Client          │  │ systems | controls | statements     │  │
│  └─────────────────┘  └─────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### Pull Operation Flow

```
1. User initiates pull
       │
       ▼
2. Create pull job record (status: started)
       │
       ▼
3. Fetch controls from ServiceNow (paginated)
       │
       ├──► For each page:
       │    ├── Transform ServiceNow → local schema
       │    ├── Upsert controls to database
       │    └── Update progress
       │
       ▼
4. Fetch statements for each control
       │
       ├──► For each control:
       │    ├── Fetch implementation statement
       │    ├── Fetch evidence references
       │    └── Store with sync metadata
       │
       ▼
5. Update pull job record (status: completed)
       │
       ▼
6. Return summary to user
```

---

## 3. Technical Stack

### Backend Components

| Component | Technology | Justification |
|-----------|------------|---------------|
| Background Jobs | Go routines + channels | Simple, no external queue needed |
| Progress Tracking | In-memory + DB | Hybrid for performance |
| Data Transformation | Custom mappers | Explicit control over mapping |
| Bulk Insert | pgx batch | Efficient bulk operations |

### ServiceNow API Usage

| Table | Purpose | Fields |
|-------|---------|--------|
| sn_grc_profile | Systems | sys_id, name, owner, description |
| sn_compliance_control | Controls | sys_id, number, name, policy |
| sn_compliance_policy_statement | Statements | sys_id, statement, control |
| sn_grc_evidence | Evidence | sys_id, name, reference_url |

---

## 4. Data Design

### Database Schema

```sql
-- migrations/20260127_002_create_systems_controls.up.sql

CREATE TYPE impact_level AS ENUM ('low', 'moderate', 'high');
CREATE TYPE control_responsibility AS ENUM ('common', 'hybrid', 'system_specific');
CREATE TYPE statement_status AS ENUM ('draft', 'review', 'approved');
CREATE TYPE pull_status AS ENUM ('started', 'completed', 'failed', 'partial');

-- Systems table
CREATE TABLE systems (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    servicenow_sys_id VARCHAR(32) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner VARCHAR(255),
    impact_level impact_level,

    -- Sync metadata
    last_pull_at TIMESTAMPTZ,
    last_pull_status pull_status,
    last_pull_by UUID REFERENCES users(id),

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Controls table
CREATE TABLE controls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    system_id UUID NOT NULL REFERENCES systems(id) ON DELETE CASCADE,
    servicenow_sys_id VARCHAR(32) NOT NULL,

    -- Control metadata
    control_family VARCHAR(10) NOT NULL,  -- AC, AU, etc.
    control_number VARCHAR(20) NOT NULL,  -- AC-1, AC-2(1)
    title VARCHAR(500) NOT NULL,
    description TEXT,
    responsibility control_responsibility DEFAULT 'system_specific',
    baseline_impact impact_level,

    -- Sync metadata
    servicenow_updated_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(system_id, servicenow_sys_id)
);

CREATE INDEX idx_controls_system ON controls(system_id);
CREATE INDEX idx_controls_family ON controls(control_family);
CREATE INDEX idx_controls_number ON controls(control_number);

-- Implementation statements table
CREATE TABLE implementation_statements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    control_id UUID NOT NULL REFERENCES controls(id) ON DELETE CASCADE,
    servicenow_sys_id VARCHAR(32),

    -- Content
    content_html TEXT,
    content_plain TEXT,  -- For search
    word_count INTEGER DEFAULT 0,
    status statement_status DEFAULT 'draft',

    -- Sync tracking
    servicenow_updated_at TIMESTAMPTZ,
    local_modified_at TIMESTAMPTZ,  -- NULL = not modified locally
    local_modified_by UUID REFERENCES users(id),
    original_content_html TEXT,  -- Snapshot at last pull
    version INTEGER DEFAULT 1,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(control_id)  -- One statement per control
);

CREATE INDEX idx_statements_control ON implementation_statements(control_id);
CREATE INDEX idx_statements_modified ON implementation_statements(local_modified_at)
    WHERE local_modified_at IS NOT NULL;

-- Evidence references table
CREATE TABLE evidence_references (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    statement_id UUID NOT NULL REFERENCES implementation_statements(id) ON DELETE CASCADE,
    servicenow_sys_id VARCHAR(32),

    name VARCHAR(255) NOT NULL,
    description TEXT,
    reference_url VARCHAR(2048),
    evidence_type VARCHAR(50),

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_evidence_statement ON evidence_references(statement_id);

-- Pull job tracking
CREATE TABLE pull_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    system_id UUID NOT NULL REFERENCES systems(id),
    status pull_status DEFAULT 'started',

    -- Progress
    controls_total INTEGER DEFAULT 0,
    controls_processed INTEGER DEFAULT 0,
    statements_total INTEGER DEFAULT 0,
    statements_processed INTEGER DEFAULT 0,
    errors_count INTEGER DEFAULT 0,

    -- Timing
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,

    -- Error details
    error_message TEXT,
    error_details JSONB,

    created_by UUID REFERENCES users(id)
);

CREATE INDEX idx_pull_jobs_system ON pull_jobs(system_id);
CREATE INDEX idx_pull_jobs_status ON pull_jobs(status);
```

### Data Transformation

```go
// internal/domain/pull/transformer.go

type ServiceNowControl struct {
    SysID       string `json:"sys_id"`
    Number      string `json:"number"`
    Name        string `json:"name"`
    Description string `json:"description"`
    SysUpdatedOn string `json:"sys_updated_on"`
}

func TransformControl(snControl ServiceNowControl, systemID uuid.UUID) (*Control, error) {
    family, number := parseControlNumber(snControl.Number) // "AC-1" → "AC", "AC-1"

    return &Control{
        ID:                uuid.New(),
        SystemID:          systemID,
        ServiceNowSysID:   snControl.SysID,
        ControlFamily:     family,
        ControlNumber:     number,
        Title:             snControl.Name,
        Description:       snControl.Description,
        ServiceNowUpdatedAt: parseServiceNowTime(snControl.SysUpdatedOn),
    }, nil
}
```

---

## 5. API Design

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/systems | List available systems |
| GET | /api/v1/systems/:id | Get system details |
| POST | /api/v1/systems/:id/pull | Initiate pull |
| GET | /api/v1/systems/:id/pull/:pullId | Get pull status |
| GET | /api/v1/systems/:id/controls | List controls for system |

### Pull Initiation

```go
// POST /api/v1/systems/:id/pull

type PullRequest struct {
    // No body needed for MVP
    // Future: selective control families
}

type PullResponse struct {
    PullID    uuid.UUID `json:"pull_id"`
    Status    string    `json:"status"`
    StartedAt time.Time `json:"started_at"`
}
```

### Pull Status

```go
// GET /api/v1/systems/:id/pull/:pullId

type PullStatusResponse struct {
    PullID             uuid.UUID `json:"pull_id"`
    Status             string    `json:"status"`
    Progress           int       `json:"progress"` // 0-100
    ControlsTotal      int       `json:"controls_total"`
    ControlsProcessed  int       `json:"controls_processed"`
    StatementsTotal    int       `json:"statements_total"`
    StatementsProcessed int      `json:"statements_processed"`
    ErrorsCount        int       `json:"errors_count"`
    Errors             []PullError `json:"errors,omitempty"`
}
```

---

## 6. Component Architecture

### Pull Service Design

```go
// internal/domain/pull/service.go

type PullService struct {
    snClient     servicenow.Client
    systemRepo   SystemRepository
    controlRepo  ControlRepository
    statementRepo StatementRepository
    pullJobRepo  PullJobRepository
    logger       zerolog.Logger
}

func (s *PullService) StartPull(ctx context.Context, systemID uuid.UUID, userID uuid.UUID) (*PullJob, error) {
    // 1. Create pull job record
    job := &PullJob{
        ID:        uuid.New(),
        SystemID:  systemID,
        Status:    "started",
        StartedAt: time.Now(),
        CreatedBy: userID,
    }
    if err := s.pullJobRepo.Create(ctx, job); err != nil {
        return nil, fmt.Errorf("create pull job: %w", err)
    }

    // 2. Start background pull
    go s.executePull(context.Background(), job)

    return job, nil
}

func (s *PullService) executePull(ctx context.Context, job *PullJob) {
    defer s.finalizePull(ctx, job)

    // Fetch controls with pagination
    offset := 0
    limit := 100
    for {
        controls, total, err := s.snClient.GetControls(ctx, job.SystemID, offset, limit)
        if err != nil {
            job.ErrorMessage = err.Error()
            job.Status = "failed"
            return
        }

        job.ControlsTotal = total

        // Process batch
        for _, snControl := range controls {
            if err := s.processControl(ctx, job, snControl); err != nil {
                job.ErrorsCount++
                // Continue with next control
            }
            job.ControlsProcessed++
            s.pullJobRepo.UpdateProgress(ctx, job)
        }

        offset += limit
        if offset >= total {
            break
        }
    }

    job.Status = "completed"
}
```

### ServiceNow Client

```go
// internal/infrastructure/servicenow/client.go

type Client interface {
    GetSystems(ctx context.Context) ([]System, error)
    GetControls(ctx context.Context, systemID string, offset, limit int) ([]Control, int, error)
    GetStatement(ctx context.Context, controlID string) (*Statement, error)
    GetEvidence(ctx context.Context, statementID string) ([]Evidence, error)
}

type client struct {
    baseURL    string
    httpClient *retryablehttp.Client
    auth       AuthProvider
}

func (c *client) GetControls(ctx context.Context, systemID string, offset, limit int) ([]Control, int, error) {
    url := fmt.Sprintf("%s/api/now/table/sn_compliance_control", c.baseURL)

    req, _ := retryablehttp.NewRequest("GET", url, nil)
    q := req.URL.Query()
    q.Set("sysparm_query", fmt.Sprintf("profile=%s", systemID))
    q.Set("sysparm_limit", strconv.Itoa(limit))
    q.Set("sysparm_offset", strconv.Itoa(offset))
    q.Set("sysparm_fields", "sys_id,number,name,description,sys_updated_on")
    req.URL.RawQuery = q.Encode()

    resp, err := c.httpClient.Do(req)
    // ... handle response
}
```

---

## 7. State Management

### Pull Progress Tracking

```go
// In-memory progress for fast polling
type PullProgressTracker struct {
    mu       sync.RWMutex
    progress map[uuid.UUID]*PullProgress
}

type PullProgress struct {
    Status            string
    ControlsTotal     int
    ControlsProcessed int
    StatementsTotal   int
    StatementsProcessed int
    LastUpdated       time.Time
}

// Update progress every N controls (not every single one)
const progressUpdateInterval = 10
```

### Frontend State

```typescript
// src/features/workspace/hooks/usePull.ts

export function usePull(systemId: string) {
  const queryClient = useQueryClient();

  const pullMutation = useMutation({
    mutationFn: () => pullApi.startPull(systemId),
    onSuccess: (data) => {
      // Start polling for progress
      queryClient.invalidateQueries(['pull', systemId]);
    },
  });

  const pullStatus = useQuery({
    queryKey: ['pull', systemId, 'status'],
    queryFn: () => pullApi.getPullStatus(systemId, pullId),
    enabled: !!pullId && status !== 'completed',
    refetchInterval: 2000, // Poll every 2 seconds
  });

  return { pullMutation, pullStatus };
}
```

---

## 8. Security Considerations

### Authorization
- Only authenticated users can view systems
- Only users with `author` or higher role can initiate pulls
- System access can be restricted by role (future)

### Data Validation
- ServiceNow sys_id validated (32 char alphanumeric)
- Control numbers validated against 800-53 pattern
- Content sanitized before storage

### Rate Limiting
- Maximum 1 concurrent pull per user
- Maximum 1 concurrent pull per system

---

## 9. Performance & Scalability

### Performance Targets

| Operation | Target | Approach |
|-----------|--------|----------|
| System list | < 3s | Cache in TanStack Query |
| Pull 300 controls | < 2 min | Batch operations, parallel where possible |
| Progress updates | 2s interval | Efficient polling |

### Optimization Strategies

1. **Batch Database Operations**
   ```go
   // Use pgx batch for bulk inserts
   batch := &pgx.Batch{}
   for _, control := range controls {
       batch.Queue(insertControlSQL, control.Fields()...)
   }
   results := conn.SendBatch(ctx, batch)
   ```

2. **Parallel API Calls**
   ```go
   // Fetch statements concurrently (with limits)
   sem := make(chan struct{}, 5) // Max 5 concurrent
   for _, control := range controls {
       sem <- struct{}{}
       go func(c Control) {
           defer func() { <-sem }()
           s.fetchStatement(ctx, c)
       }(control)
   }
   ```

3. **Pagination**
   - Fetch 100 records per ServiceNow API call
   - Process in batches to limit memory

---

## 10. Testing Strategy

### Unit Tests

```go
// internal/domain/pull/transformer_test.go
func TestTransformControl(t *testing.T) {
    tests := []struct {
        name     string
        input    ServiceNowControl
        expected Control
    }{
        {
            name: "standard control",
            input: ServiceNowControl{Number: "AC-1", Name: "Access Control"},
            expected: Control{ControlFamily: "AC", ControlNumber: "AC-1"},
        },
        {
            name: "control enhancement",
            input: ServiceNowControl{Number: "AC-2(1)", Name: "..."},
            expected: Control{ControlFamily: "AC", ControlNumber: "AC-2(1)"},
        },
    }
    // ...
}
```

### Integration Tests

```go
// Test with mock ServiceNow server
func TestPullService_ExecutePull(t *testing.T) {
    mockSN := httptest.NewServer(mockServiceNowHandler())
    defer mockSN.Close()

    // Test full pull flow
}
```

### E2E Tests

```typescript
// tests/e2e/pull.spec.ts
test('user can pull control package', async ({ page }) => {
  await page.goto('/workspace');
  await page.selectOption('#system-select', 'System A');
  await page.click('button:has-text("Pull from ServiceNow")');
  await expect(page.locator('.pull-progress')).toBeVisible();
  await expect(page.locator('.pull-complete')).toBeVisible({ timeout: 120000 });
});
```

---

## 11. Deployment & DevOps

### Configuration

```bash
# ServiceNow API settings
SN_API_TIMEOUT=30s
SN_API_RETRY_MAX=3
SN_API_PAGE_SIZE=100

# Pull settings
PULL_MAX_CONCURRENT_PER_USER=1
PULL_PROGRESS_UPDATE_INTERVAL=10
```

### Monitoring

- Log pull start/complete/fail events
- Track pull duration metrics
- Alert on pull failure rate > 5%

---

## 12. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| ServiceNow timeout | Medium | Medium | Retry logic, configurable timeout |
| Large dataset OOM | Low | High | Streaming/pagination, memory limits |
| Schema mismatch | Medium | Medium | Flexible parsing, error logging |
| Concurrent pull conflicts | Low | Low | Database constraints, locking |

---

## 13. Development Phases

### Phase 1: Data Layer (2 days)
- Database migrations
- Repository implementations
- Unit tests

### Phase 2: ServiceNow Client (2 days)
- API client implementation
- Pagination handling
- Error handling

### Phase 3: Pull Service (3 days)
- Pull orchestration
- Progress tracking
- Background job management

### Phase 4: API & Frontend (2 days)
- API endpoints
- System selector UI
- Pull progress UI

### Phase 5: Testing & Polish (1 day)
- Integration tests
- E2E tests
- Error message refinement

**Estimated Total: 10 days**

---

## 14. Open Technical Questions

1. **ServiceNow Schema:** What are the exact table/field names in ServiceNow GRC?
2. **Incremental Pull:** Should we support delta pulls (only changed records)?
3. **Concurrent Pulls:** How to handle if another user starts pull mid-operation?
