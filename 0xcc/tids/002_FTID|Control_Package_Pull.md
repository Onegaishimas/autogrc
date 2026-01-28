# Technical Implementation Document: Control Package Pull (F2)

**Document Version:** 1.0
**Date:** 2026-01-27
**Feature ID:** F2
**Related PRD:** 002_FPRD|Control_Package_Pull.md
**Related TDD:** 002_FTDD|Control_Package_Pull.md

---

## 1. Implementation Overview

### Summary
This TID provides implementation guidance for pulling control packages from ServiceNow GRC, including system discovery, control import with pagination handling, statement synchronization, and sync state tracking.

### Key Implementation Principles
- **Background Processing:** Long-running pulls executed as jobs, not blocking HTTP requests
- **Idempotent Operations:** Re-running a pull updates existing records, doesn't create duplicates
- **Incremental Sync:** Track `sys_updated_on` to enable delta syncs
- **Transactional Integrity:** Database operations wrapped in transactions

### Integration Points
- **F1 Dependency:** Uses connection config from ServiceNow Connection feature
- **Database:** PostgreSQL for systems, controls, statements, sync state
- **ServiceNow API:** Table API with pagination support
- **Frontend:** React components for system selection and pull status

---

## 2. File Structure and Organization

### Backend Files to Create

```
backend/
├── internal/
│   ├── api/
│   │   └── handlers/
│   │       └── sync/
│   │           ├── handler.go            # HTTP handlers (create)
│   │           ├── handler_test.go       # Handler tests (create)
│   │           └── schemas.go            # Request/response DTOs (create)
│   ├── domain/
│   │   ├── system/
│   │   │   ├── models.go                 # System domain models (create)
│   │   │   ├── service.go                # System business logic (create)
│   │   │   ├── service_test.go           # Service tests (create)
│   │   │   └── repository.go             # DB interface (create)
│   │   ├── control/
│   │   │   ├── models.go                 # Control domain models (create)
│   │   │   ├── service.go                # Control business logic (create)
│   │   │   ├── service_test.go           # Service tests (create)
│   │   │   └── repository.go             # DB interface (create)
│   │   └── statement/
│   │       ├── models.go                 # Statement domain models (create)
│   │       ├── service.go                # Statement business logic (create)
│   │       ├── service_test.go           # Service tests (create)
│   │       └── repository.go             # DB interface (create)
│   ├── jobs/
│   │   └── pull/
│   │       ├── executor.go               # Pull job execution (create)
│   │       ├── executor_test.go          # Executor tests (create)
│   │       └── progress.go               # Progress tracking (create)
│   └── infrastructure/
│       ├── database/
│       │   └── queries/
│       │       ├── system.sql            # System queries (create)
│       │       ├── control.sql           # Control queries (create)
│       │       └── statement.sql         # Statement queries (create)
│       └── servicenow/
│           ├── client.go                 # Add pull methods (modify)
│           └── pagination.go             # Pagination handler (create)
└── migrations/
    └── 20260127_002_create_pull_tables.sql  # Schema (create)
```

### Frontend Files to Create

```
frontend/src/
├── features/
│   ├── systems/
│   │   ├── components/
│   │   │   ├── SystemList.tsx            # System listing (create)
│   │   │   ├── SystemCard.tsx            # Individual system card (create)
│   │   │   └── SystemSelector.tsx        # Multi-select for pull (create)
│   │   ├── hooks/
│   │   │   └── useSystems.ts             # System query hooks (create)
│   │   ├── api/
│   │   │   └── systemsApi.ts             # API client (create)
│   │   └── types.ts                      # TypeScript types (create)
│   └── pull/
│       ├── components/
│       │   ├── PullWizard.tsx            # Multi-step pull flow (create)
│       │   ├── PullProgress.tsx          # Progress indicator (create)
│       │   └── PullResults.tsx           # Results summary (create)
│       ├── hooks/
│       │   ├── usePull.ts                # Pull mutation hooks (create)
│       │   └── usePullStatus.ts          # Status polling hook (create)
│       ├── api/
│       │   └── pullApi.ts                # API client (create)
│       └── types.ts                      # TypeScript types (create)
└── pages/
    └── PullPage.tsx                      # Pull workflow page (create)
```

---

## 3. Component Implementation Hints

### Pull Job Pattern (Backend)

```go
// internal/jobs/pull/executor.go
type PullExecutor struct {
    snClient    servicenow.Client
    systemRepo  system.Repository
    controlRepo control.Repository
    stmtRepo    statement.Repository
    logger      zerolog.Logger
}

type PullJob struct {
    ID        uuid.UUID
    SystemIDs []uuid.UUID
    Status    JobStatus // pending, running, completed, failed
    Progress  PullProgress
    StartedAt *time.Time
    Error     *string
}

func (e *PullExecutor) Execute(ctx context.Context, job *PullJob) error {
    job.Status = JobStatusRunning
    job.StartedAt = ptr(time.Now())

    for _, sysID := range job.SystemIDs {
        if err := e.pullSystem(ctx, job, sysID); err != nil {
            job.Status = JobStatusFailed
            job.Error = ptr(err.Error())
            return err
        }
    }

    job.Status = JobStatusCompleted
    return nil
}
```

**Key Patterns:**
- Jobs stored in database for persistence across restarts
- Progress tracked per-system for UI updates
- Context cancellation respected for graceful shutdown
- Partial success allowed (some systems succeed, others fail)

### ServiceNow Pagination Pattern

```go
// internal/infrastructure/servicenow/pagination.go
type PaginatedResponse struct {
    Result []json.RawMessage `json:"result"`
}

func (c *Client) FetchAllPages(ctx context.Context, table string, query string) ([]json.RawMessage, error) {
    var allResults []json.RawMessage
    offset := 0
    limit := 100 // ServiceNow default page size

    for {
        url := fmt.Sprintf("%s/api/now/table/%s?sysparm_query=%s&sysparm_offset=%d&sysparm_limit=%d",
            c.baseURL, table, url.QueryEscape(query), offset, limit)

        resp, err := c.doRequest(ctx, "GET", url)
        if err != nil {
            return nil, fmt.Errorf("failed to fetch page at offset %d: %w", offset, err)
        }

        var page PaginatedResponse
        if err := json.Unmarshal(resp, &page); err != nil {
            return nil, fmt.Errorf("failed to parse response: %w", err)
        }

        allResults = append(allResults, page.Result...)

        if len(page.Result) < limit {
            break // Last page
        }
        offset += limit
    }

    return allResults, nil
}
```

**Key Hints:**
- Use `sysparm_offset` and `sysparm_limit` for pagination
- ServiceNow max page size is typically 1000
- Check `X-Total-Count` header for progress calculation if available
- Handle rate limiting with exponential backoff

---

## 4. Database Implementation Approach

### Schema Design

```sql
-- migrations/20260127_002_create_pull_tables.sql

-- Systems table
CREATE TABLE systems (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sn_sys_id VARCHAR(32) UNIQUE NOT NULL,  -- ServiceNow sys_id
    name VARCHAR(255) NOT NULL,
    description TEXT,
    acronym VARCHAR(50),
    owner VARCHAR(255),
    status VARCHAR(50),

    -- Sync metadata
    sn_updated_on TIMESTAMPTZ,
    last_pull_at TIMESTAMPTZ,
    last_push_at TIMESTAMPTZ,

    -- Audit
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Controls table
CREATE TABLE controls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    system_id UUID NOT NULL REFERENCES systems(id) ON DELETE CASCADE,
    sn_sys_id VARCHAR(32) NOT NULL,

    -- Control metadata from NIST 800-53
    control_id VARCHAR(20) NOT NULL,        -- e.g., "AC-1"
    control_name VARCHAR(255) NOT NULL,
    control_family VARCHAR(50) NOT NULL,    -- e.g., "Access Control"
    baseline VARCHAR(20),                    -- LOW, MODERATE, HIGH

    -- Implementation details
    implementation_status VARCHAR(50),
    responsible_role VARCHAR(255),

    -- Sync metadata
    sn_updated_on TIMESTAMPTZ,

    -- Audit
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(system_id, sn_sys_id)
);

-- Statements table
CREATE TABLE statements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    control_id UUID NOT NULL REFERENCES controls(id) ON DELETE CASCADE,
    sn_sys_id VARCHAR(32) NOT NULL,

    -- Statement content
    statement_type VARCHAR(50) NOT NULL,    -- implementation, assessment, etc.
    content TEXT NOT NULL,
    content_html TEXT,                       -- Rendered HTML for display

    -- Local editing state
    local_content TEXT,
    local_content_html TEXT,
    is_modified BOOLEAN DEFAULT false,
    modified_at TIMESTAMPTZ,

    -- Sync metadata
    sn_updated_on TIMESTAMPTZ,
    sync_status VARCHAR(20) DEFAULT 'synced', -- synced, modified, conflict

    -- Audit
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(control_id, sn_sys_id)
);

-- Pull jobs table
CREATE TABLE pull_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    system_ids UUID[] NOT NULL,
    progress JSONB DEFAULT '{}',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_controls_system ON controls(system_id);
CREATE INDEX idx_controls_control_id ON controls(control_id);
CREATE INDEX idx_statements_control ON statements(control_id);
CREATE INDEX idx_statements_modified ON statements(is_modified) WHERE is_modified = true;
CREATE INDEX idx_pull_jobs_status ON pull_jobs(status) WHERE status IN ('pending', 'running');
```

### sqlc Query Patterns

```sql
-- name: UpsertSystem :one
INSERT INTO systems (sn_sys_id, name, description, acronym, owner, status, sn_updated_on, last_pull_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
ON CONFLICT (sn_sys_id)
DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    acronym = EXCLUDED.acronym,
    owner = EXCLUDED.owner,
    status = EXCLUDED.status,
    sn_updated_on = EXCLUDED.sn_updated_on,
    last_pull_at = NOW(),
    updated_at = NOW()
RETURNING *;

-- name: UpsertControl :one
INSERT INTO controls (system_id, sn_sys_id, control_id, control_name, control_family, baseline, implementation_status, responsible_role, sn_updated_on)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (system_id, sn_sys_id)
DO UPDATE SET
    control_id = EXCLUDED.control_id,
    control_name = EXCLUDED.control_name,
    control_family = EXCLUDED.control_family,
    baseline = EXCLUDED.baseline,
    implementation_status = EXCLUDED.implementation_status,
    responsible_role = EXCLUDED.responsible_role,
    sn_updated_on = EXCLUDED.sn_updated_on,
    updated_at = NOW()
RETURNING *;

-- name: UpsertStatement :one
INSERT INTO statements (control_id, sn_sys_id, statement_type, content, content_html, sn_updated_on)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (control_id, sn_sys_id)
DO UPDATE SET
    content = CASE
        WHEN statements.is_modified THEN statements.content  -- Preserve local if modified
        ELSE EXCLUDED.content
    END,
    content_html = CASE
        WHEN statements.is_modified THEN statements.content_html
        ELSE EXCLUDED.content_html
    END,
    sn_updated_on = EXCLUDED.sn_updated_on,
    sync_status = CASE
        WHEN statements.is_modified AND EXCLUDED.sn_updated_on > statements.sn_updated_on THEN 'conflict'
        WHEN statements.is_modified THEN 'modified'
        ELSE 'synced'
    END,
    updated_at = NOW()
RETURNING *;
```

**Key Hints:**
- Use UPSERT (ON CONFLICT) for idempotent pulls
- Preserve local modifications during pull (check `is_modified`)
- Detect conflicts by comparing `sn_updated_on` timestamps
- Cascade deletes from system → controls → statements

---

## 5. API Implementation Strategy

### Endpoints

```go
// internal/api/handlers/sync/handler.go
func (h *Handler) RegisterRoutes(r chi.Router) {
    r.Route("/api/v1/sync", func(r chi.Router) {
        // System discovery
        r.Get("/systems/discover", h.DiscoverSystems)  // Fetch from ServiceNow
        r.Get("/systems", h.ListSystems)                // List local systems

        // Pull operations
        r.Post("/pull", h.StartPull)                    // Start pull job
        r.Get("/pull/{jobId}", h.GetPullStatus)         // Get job status
        r.Delete("/pull/{jobId}", h.CancelPull)         // Cancel running job
    })
}
```

### Request/Response Schemas

```go
// internal/api/handlers/sync/schemas.go

type DiscoverSystemsResponse struct {
    Systems []DiscoveredSystem `json:"systems"`
}

type DiscoveredSystem struct {
    SNSysID     string `json:"sn_sys_id"`
    Name        string `json:"name"`
    Acronym     string `json:"acronym,omitempty"`
    Description string `json:"description,omitempty"`
    IsImported  bool   `json:"is_imported"` // Already in local DB
}

type StartPullRequest struct {
    SystemIDs []uuid.UUID `json:"system_ids" validate:"required,min=1,max=10"`
}

type StartPullResponse struct {
    JobID uuid.UUID `json:"job_id"`
}

type PullStatusResponse struct {
    JobID       uuid.UUID    `json:"job_id"`
    Status      string       `json:"status"`
    Progress    PullProgress `json:"progress"`
    StartedAt   *time.Time   `json:"started_at,omitempty"`
    CompletedAt *time.Time   `json:"completed_at,omitempty"`
    Error       *string      `json:"error,omitempty"`
}

type PullProgress struct {
    TotalSystems     int `json:"total_systems"`
    CompletedSystems int `json:"completed_systems"`
    TotalControls    int `json:"total_controls"`
    ProcessedControls int `json:"processed_controls"`
    TotalStatements  int `json:"total_statements"`
    ProcessedStatements int `json:"processed_statements"`
}
```

---

## 6. Frontend Implementation Approach

### System Selection Component

```typescript
// features/systems/components/SystemSelector.tsx
interface SystemSelectorProps {
  onSelect: (systemIds: string[]) => void;
  maxSelection?: number;
}

export function SystemSelector({ onSelect, maxSelection = 10 }: SystemSelectorProps) {
  const { data: discoveredSystems, isLoading } = useDiscoverSystems();
  const [selected, setSelected] = useState<Set<string>>(new Set());

  const toggleSystem = (sysId: string) => {
    const next = new Set(selected);
    if (next.has(sysId)) {
      next.delete(sysId);
    } else if (next.size < maxSelection) {
      next.add(sysId);
    }
    setSelected(next);
    onSelect(Array.from(next));
  };

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {discoveredSystems?.map((system) => (
        <SystemCard
          key={system.sn_sys_id}
          system={system}
          selected={selected.has(system.sn_sys_id)}
          onToggle={() => toggleSystem(system.sn_sys_id)}
        />
      ))}
    </div>
  );
}
```

### Pull Progress Component

```typescript
// features/pull/components/PullProgress.tsx
export function PullProgress({ jobId }: { jobId: string }) {
  const { data: status, isLoading } = usePullStatus(jobId, {
    refetchInterval: (data) =>
      data?.status === 'running' ? 2000 : false, // Poll every 2s while running
  });

  if (isLoading || !status) return <Skeleton className="h-20" />;

  const systemProgress = status.progress.totalSystems > 0
    ? (status.progress.completedSystems / status.progress.totalSystems) * 100
    : 0;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          {status.status === 'running' && <Loader2 className="animate-spin" />}
          {status.status === 'completed' && <CheckCircle className="text-green-500" />}
          {status.status === 'failed' && <XCircle className="text-red-500" />}
          Pull {status.status}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Progress value={systemProgress} className="mb-2" />
        <p className="text-sm text-muted-foreground">
          {status.progress.completedSystems} of {status.progress.totalSystems} systems
        </p>
        {status.error && (
          <Alert variant="destructive" className="mt-4">
            <AlertDescription>{status.error}</AlertDescription>
          </Alert>
        )}
      </CardContent>
    </Card>
  );
}
```

### Pull Wizard Flow

```typescript
// features/pull/components/PullWizard.tsx
type Step = 'select' | 'confirm' | 'progress' | 'results';

export function PullWizard() {
  const [step, setStep] = useState<Step>('select');
  const [selectedSystems, setSelectedSystems] = useState<string[]>([]);
  const [jobId, setJobId] = useState<string | null>(null);
  const pullMutation = useStartPull();

  const handleStartPull = async () => {
    const result = await pullMutation.mutateAsync({ systemIds: selectedSystems });
    setJobId(result.jobId);
    setStep('progress');
  };

  return (
    <div className="space-y-6">
      <Stepper activeStep={step} steps={['Select Systems', 'Confirm', 'Progress', 'Results']} />

      {step === 'select' && (
        <SystemSelector onSelect={setSelectedSystems} />
      )}
      {step === 'confirm' && (
        <PullConfirmation systems={selectedSystems} onConfirm={handleStartPull} />
      )}
      {step === 'progress' && jobId && (
        <PullProgress jobId={jobId} onComplete={() => setStep('results')} />
      )}
      {step === 'results' && jobId && (
        <PullResults jobId={jobId} />
      )}

      <div className="flex justify-between">
        {step !== 'select' && step !== 'progress' && (
          <Button variant="outline" onClick={() => setStep('select')}>Back</Button>
        )}
        {step === 'select' && selectedSystems.length > 0 && (
          <Button onClick={() => setStep('confirm')}>Continue</Button>
        )}
      </div>
    </div>
  );
}
```

---

## 7. Business Logic Implementation Hints

### Data Transformation

```go
// internal/jobs/pull/executor.go

type SNControlResponse struct {
    SysID              string `json:"sys_id"`
    Number             string `json:"number"`          // e.g., "AC-1"
    ShortDescription   string `json:"short_description"`
    ControlFamily      string `json:"u_control_family"`
    ImplementationStatus string `json:"u_implementation_status"`
}

func (e *PullExecutor) transformControl(snCtrl SNControlResponse, systemID uuid.UUID) *control.Control {
    return &control.Control{
        SystemID:             systemID,
        SNSysID:              snCtrl.SysID,
        ControlID:            snCtrl.Number,
        ControlName:          snCtrl.ShortDescription,
        ControlFamily:        extractFamily(snCtrl.Number), // "AC-1" → "Access Control"
        ImplementationStatus: snCtrl.ImplementationStatus,
    }
}

func extractFamily(controlID string) string {
    // Map control ID prefix to NIST 800-53 family name
    families := map[string]string{
        "AC": "Access Control",
        "AT": "Awareness and Training",
        "AU": "Audit and Accountability",
        "CA": "Assessment, Authorization, and Monitoring",
        "CM": "Configuration Management",
        "CP": "Contingency Planning",
        "IA": "Identification and Authentication",
        "IR": "Incident Response",
        "MA": "Maintenance",
        "MP": "Media Protection",
        "PE": "Physical and Environmental Protection",
        "PL": "Planning",
        "PM": "Program Management",
        "PS": "Personnel Security",
        "PT": "PII Processing and Transparency",
        "RA": "Risk Assessment",
        "SA": "System and Services Acquisition",
        "SC": "System and Communications Protection",
        "SI": "System and Information Integrity",
        "SR": "Supply Chain Risk Management",
    }

    prefix := strings.Split(controlID, "-")[0]
    if family, ok := families[prefix]; ok {
        return family
    }
    return "Unknown"
}
```

### Conflict Detection

```go
// internal/domain/statement/service.go

func (s *Service) DetectConflict(local *Statement, remote SNStatementResponse) ConflictType {
    remoteUpdated, _ := time.Parse(time.RFC3339, remote.SysUpdatedOn)

    // No local modifications - safe to overwrite
    if !local.IsModified {
        return ConflictNone
    }

    // Local modified, remote unchanged - safe to push later
    if remoteUpdated.Equal(local.SNUpdatedOn) || remoteUpdated.Before(local.SNUpdatedOn) {
        return ConflictNone
    }

    // Both modified - conflict!
    return ConflictBothModified
}
```

---

## 8. Testing Implementation Approach

### Mock ServiceNow Responses

```go
// internal/jobs/pull/executor_test.go

func TestPullExecutor_Execute(t *testing.T) {
    mockSN := &MockSNClient{}
    mockSN.On("FetchSystems", mock.Anything).Return([]SNSystemResponse{
        {SysID: "sys1", Name: "Test System"},
    }, nil)
    mockSN.On("FetchControls", mock.Anything, "sys1").Return([]SNControlResponse{
        {SysID: "ctrl1", Number: "AC-1", ShortDescription: "Access Control Policy"},
    }, nil)

    executor := NewPullExecutor(mockSN, mockRepo, logger)

    job := &PullJob{ID: uuid.New(), SystemIDs: []uuid.UUID{systemID}}
    err := executor.Execute(context.Background(), job)

    assert.NoError(t, err)
    assert.Equal(t, JobStatusCompleted, job.Status)
}
```

### Frontend Test Pattern

```typescript
// features/pull/components/PullProgress.test.tsx
describe('PullProgress', () => {
  it('polls status while running', async () => {
    const { result } = renderHook(() => usePullStatus('job-1'), {
      wrapper: createQueryWrapper(),
    });

    await waitFor(() => {
      expect(result.current.data?.status).toBe('running');
    });

    // Advance timers and verify refetch
    jest.advanceTimersByTime(2000);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
  });

  it('stops polling when completed', async () => {
    mockPullStatus({ status: 'completed' });

    const { result } = renderHook(() => usePullStatus('job-1'));

    await waitFor(() => {
      expect(result.current.data?.status).toBe('completed');
    });

    jest.advanceTimersByTime(5000);
    expect(fetchMock).toHaveBeenCalledTimes(1); // No additional calls
  });
});
```

---

## 9. Configuration and Environment Strategy

### ServiceNow Table Configuration

```go
// internal/infrastructure/servicenow/tables.go
const (
    TableSystems    = "x_custom_grc_systems"     // Or cmdb_ci_service
    TableControls   = "sn_compliance_control"
    TableStatements = "sn_compliance_statement"
)

// These may need customization per ServiceNow instance
// Consider making configurable via environment or database
```

### Environment Variables

```bash
# Pull configuration
SN_PULL_PAGE_SIZE=100
SN_PULL_TIMEOUT=5m
SN_PULL_MAX_SYSTEMS=10

# Job configuration
PULL_JOB_RETENTION_DAYS=30
```

---

## 10. Integration Strategy

### Job Processing Integration

```go
// cmd/server/main.go
func main() {
    // ... existing setup ...

    // Initialize pull job executor
    pullExecutor := pull.NewExecutor(snClient, systemRepo, controlRepo, stmtRepo, logger)

    // Start background job processor
    jobProcessor := jobs.NewProcessor(pullJobRepo, pullExecutor, logger)
    go jobProcessor.Start(ctx)

    // Register routes
    syncHandler := handlers.NewSyncHandler(pullExecutor, systemRepo, logger)
    syncHandler.RegisterRoutes(r)
}
```

### Graceful Shutdown

```go
// internal/jobs/processor.go
func (p *Processor) Start(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            p.logger.Info().Msg("shutting down job processor")
            return
        default:
            job, err := p.repo.GetNextPending(ctx)
            if err != nil {
                time.Sleep(5 * time.Second)
                continue
            }
            if job != nil {
                p.executor.Execute(ctx, job)
            } else {
                time.Sleep(1 * time.Second)
            }
        }
    }
}
```

---

## 11. Error Handling and Logging Strategy

### Pull Job Error Handling

```go
func (e *PullExecutor) pullSystem(ctx context.Context, job *PullJob, systemID uuid.UUID) error {
    e.logger.Info().
        Str("job_id", job.ID.String()).
        Str("system_id", systemID.String()).
        Msg("starting system pull")

    // Fetch system from ServiceNow
    snSystem, err := e.snClient.GetSystem(ctx, systemID.String())
    if err != nil {
        e.logger.Error().
            Err(err).
            Str("system_id", systemID.String()).
            Msg("failed to fetch system from ServiceNow")
        return fmt.Errorf("failed to fetch system %s: %w", systemID, err)
    }

    // Continue with controls and statements...
}
```

### User-Facing Progress Messages

| Phase | Message |
|-------|---------|
| Starting | "Starting pull for X systems..." |
| Systems | "Fetching system information..." |
| Controls | "Importing controls (X of Y)..." |
| Statements | "Syncing implementation statements..." |
| Completed | "Successfully imported X controls and Y statements" |
| Failed | "Pull failed: {user-friendly error message}" |

---

## 12. Performance Implementation Hints

### Batch Database Operations

```go
// Use batch inserts for controls/statements
func (r *Repository) UpsertControls(ctx context.Context, controls []*Control) error {
    batch := &pgx.Batch{}
    for _, c := range controls {
        batch.Queue(upsertControlSQL, c.SystemID, c.SNSysID, c.ControlID, ...)
    }
    results := r.db.SendBatch(ctx, batch)
    defer results.Close()

    for range controls {
        if _, err := results.Exec(); err != nil {
            return fmt.Errorf("batch upsert failed: %w", err)
        }
    }
    return nil
}
```

### Progress Update Throttling

```go
// Update progress every N records or T seconds, not every record
const progressUpdateInterval = 100 // records
const progressUpdateTimeout = 5 * time.Second

func (e *PullExecutor) updateProgress(job *PullJob, processed int) {
    if processed%progressUpdateInterval == 0 || time.Since(e.lastProgressUpdate) > progressUpdateTimeout {
        e.repo.UpdateProgress(job.ID, job.Progress)
        e.lastProgressUpdate = time.Now()
    }
}
```

---

## 13. Security Checklist

- [ ] Validate system IDs are UUIDs before processing
- [ ] Limit max systems per pull request (prevent DoS)
- [ ] Rate limit pull requests per user
- [ ] Verify user has access to requested systems (future: RBAC)
- [ ] Sanitize ServiceNow response data before storage
- [ ] Log pull operations for audit trail (F5)
