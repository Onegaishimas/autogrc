# Technical Design Document: Control Package Push (F4)

**Document Version:** 1.0
**Date:** 2026-01-27
**Feature ID:** F4
**Related PRD:** 004_FPRD|Control_Package_Push.md

---

## 1. Executive Summary

This TDD defines the technical architecture for pushing modified implementation statements from ControlCRUD back to ServiceNow GRC. The feature handles conflict detection, selective push, and provides robust error recovery.

**Key Technical Decisions:**
- Pre-push conflict detection (optimistic concurrency)
- Per-statement push with retry logic
- Background job for bulk operations
- Comprehensive result reporting

---

## 2. System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           FRONTEND                                       │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────┐ │
│  │ Push Preview    │  │ Push Progress   │  │ Conflict Resolution     │ │
│  │ Modal           │  │ Display         │  │ Modal                   │ │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────────────┘ │
└───────────┼────────────────────┼────────────────────┼───────────────────┘
            │                    │                    │
            ▼                    ▼                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           BACKEND (Go)                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                         Push Service                                ││
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌───────────┐ ││
│  │  │ Conflict    │  │ Push        │  │ Result      │  │ Audit     │ ││
│  │  │ Detector    │  │ Executor    │  │ Aggregator  │  │ Logger    │ ││
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └───────────┘ ││
│  └─────────────────────────────────────────────────────────────────────┘│
│           │                    │                                         │
│           ▼                    ▼                                         │
│  ┌─────────────────┐  ┌─────────────────────────────────────────────┐  │
│  │ ServiceNow      │  │              PostgreSQL                      │  │
│  │ Client          │  │  statements | push_jobs | sync_log          │  │
│  └─────────────────┘  └─────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
```

### Push Operation Flow

```
1. User requests push preview
       │
       ▼
2. Backend identifies modified statements
       │
       ▼
3. Backend checks ServiceNow for conflicts
       │
       ├──► No conflicts: Ready to push
       │
       └──► Conflicts found: Return conflict details
              │
              ▼
       4. User resolves conflicts (skip/overwrite)
              │
              ▼
5. User confirms push
       │
       ▼
6. Backend creates push job
       │
       ▼
7. For each statement:
       ├──► Fetch current ServiceNow state
       ├──► Compare versions
       ├──► If OK: PUT to ServiceNow
       ├──► Update local state
       └──► Log result
       │
       ▼
8. Return push results
```

---

## 3. Technical Stack

### Backend Components

| Component | Technology | Justification |
|-----------|------------|---------------|
| Background Jobs | Go routines | Simple, in-process |
| Retry Logic | go-retryablehttp | Exponential backoff |
| Diff Generation | sergi/go-diff | Content comparison |
| Concurrent Push | Semaphore pattern | Controlled parallelism |

### Frontend Components

| Component | Technology | Justification |
|-----------|------------|---------------|
| Diff Viewer | react-diff-viewer-continued | Side-by-side diff |
| Modal | Shadcn Dialog | Consistent UI |
| Progress | Shadcn Progress | Animated progress |

---

## 4. Data Design

### Database Schema Additions

```sql
-- migrations/20260127_004_create_push_jobs.up.sql

CREATE TYPE push_status AS ENUM ('pending', 'started', 'completed', 'failed', 'partial');
CREATE TYPE push_item_result AS ENUM ('success', 'failed', 'skipped', 'conflict');
CREATE TYPE conflict_resolution AS ENUM ('overwrite', 'skip', 'pull');

-- Push job tracking
CREATE TABLE push_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    system_id UUID NOT NULL REFERENCES systems(id),
    status push_status DEFAULT 'pending',

    -- Counts
    statements_total INTEGER DEFAULT 0,
    statements_succeeded INTEGER DEFAULT 0,
    statements_failed INTEGER DEFAULT 0,
    statements_skipped INTEGER DEFAULT 0,
    conflicts_count INTEGER DEFAULT 0,

    -- Timing
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,

    -- Error summary
    error_message TEXT,

    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_push_jobs_system ON push_jobs(system_id);

-- Individual push results
CREATE TABLE push_job_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    push_job_id UUID NOT NULL REFERENCES push_jobs(id) ON DELETE CASCADE,
    statement_id UUID NOT NULL REFERENCES implementation_statements(id),
    control_number VARCHAR(20) NOT NULL,

    result push_item_result,
    error_code VARCHAR(50),
    error_message TEXT,

    -- Conflict details
    had_conflict BOOLEAN DEFAULT false,
    conflict_resolution conflict_resolution,
    local_content_hash VARCHAR(64),
    remote_content_hash VARCHAR(64),

    -- ServiceNow response
    servicenow_response JSONB,

    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_push_items_job ON push_job_items(push_job_id);

-- Add conflict tracking to statements
ALTER TABLE implementation_statements
    ADD COLUMN IF NOT EXISTS push_conflict_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS push_conflict_remote_content TEXT;
```

### Conflict Detection Model

```go
// internal/domain/push/conflict.go

type ConflictInfo struct {
    StatementID     uuid.UUID `json:"statement_id"`
    ControlNumber   string    `json:"control_number"`
    LocalModifiedAt time.Time `json:"local_modified_at"`
    LocalModifiedBy string    `json:"local_modified_by"`
    RemoteModifiedAt time.Time `json:"remote_modified_at"`
    RemoteModifiedBy string    `json:"remote_modified_by"`
    LocalContent    string    `json:"local_content"`
    RemoteContent   string    `json:"remote_content"`
}

type PushPreview struct {
    StatementsToProcess int            `json:"statements_to_push"`
    Statements          []PushItem     `json:"statements"`
    Conflicts           []ConflictInfo `json:"conflicts"`
    HasConflicts        bool           `json:"has_conflicts"`
}
```

---

## 5. API Design

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/systems/:id/push/preview | Get push preview with conflicts |
| POST | /api/v1/systems/:id/push | Start push operation |
| GET | /api/v1/systems/:id/push/:jobId | Get push job status |
| GET | /api/v1/statements/:id/diff | Get local vs remote diff |

### Push Preview

```go
// GET /api/v1/systems/:id/push/preview

type PushPreviewResponse struct {
    StatementsToProcess int `json:"statements_to_push"`
    Statements []struct {
        ID            uuid.UUID `json:"id"`
        ControlNumber string    `json:"control_number"`
        ModifiedAt    time.Time `json:"modified_at"`
        HasConflict   bool      `json:"has_conflict"`
    } `json:"statements"`
    Conflicts []ConflictInfo `json:"conflicts"`
}
```

### Start Push

```go
// POST /api/v1/systems/:id/push

type PushRequest struct {
    StatementIDs        []uuid.UUID                      `json:"statement_ids"`
    ConflictResolutions map[uuid.UUID]ConflictResolution `json:"conflict_resolutions"`
}

type ConflictResolution string
const (
    ResolutionOverwrite ConflictResolution = "overwrite"
    ResolutionSkip      ConflictResolution = "skip"
)

type PushResponse struct {
    PushID    uuid.UUID `json:"push_id"`
    Status    string    `json:"status"`
    StartedAt time.Time `json:"started_at"`
}
```

### Push Status

```go
// GET /api/v1/systems/:id/push/:jobId

type PushStatusResponse struct {
    PushID            uuid.UUID      `json:"push_id"`
    Status            string         `json:"status"`
    Progress          int            `json:"progress"`
    StatementsTotal   int            `json:"statements_total"`
    StatementsSucceeded int          `json:"statements_succeeded"`
    StatementsFailed  int            `json:"statements_failed"`
    StatementsSkipped int            `json:"statements_skipped"`
    Results           []PushItemResult `json:"results,omitempty"`
}

type PushItemResult struct {
    StatementID   uuid.UUID `json:"statement_id"`
    ControlNumber string    `json:"control_number"`
    Result        string    `json:"result"`
    Error         string    `json:"error,omitempty"`
}
```

---

## 6. Component Architecture

### Backend Package Structure

```
internal/
├── domain/
│   └── push/
│       ├── service.go         # Push orchestration
│       ├── conflict.go        # Conflict detection
│       ├── executor.go        # Push execution
│       ├── repository.go      # Database access
│       └── service_test.go
└── api/
    └── handlers/
        └── push/
            ├── handler.go     # HTTP handlers
            └── schemas.go     # Request/response types
```

### Push Service Implementation

```go
// internal/domain/push/service.go

type PushService struct {
    snClient      servicenow.Client
    statementRepo StatementRepository
    pushJobRepo   PushJobRepository
    syncLogRepo   SyncLogRepository
    logger        zerolog.Logger
}

func (s *PushService) GetPreview(ctx context.Context, systemID uuid.UUID) (*PushPreview, error) {
    // 1. Get locally modified statements
    modified, err := s.statementRepo.GetModified(ctx, systemID)
    if err != nil {
        return nil, fmt.Errorf("get modified statements: %w", err)
    }

    // 2. Check each for conflicts
    preview := &PushPreview{
        StatementsToProcess: len(modified),
        Statements:          make([]PushItem, len(modified)),
        Conflicts:           []ConflictInfo{},
    }

    for i, stmt := range modified {
        conflict, err := s.checkConflict(ctx, stmt)
        if err != nil {
            return nil, fmt.Errorf("check conflict for %s: %w", stmt.ID, err)
        }

        preview.Statements[i] = PushItem{
            ID:            stmt.ID,
            ControlNumber: stmt.ControlNumber,
            ModifiedAt:    stmt.LocalModifiedAt,
            HasConflict:   conflict != nil,
        }

        if conflict != nil {
            preview.Conflicts = append(preview.Conflicts, *conflict)
            preview.HasConflicts = true
        }
    }

    return preview, nil
}

func (s *PushService) checkConflict(ctx context.Context, stmt *Statement) (*ConflictInfo, error) {
    // Fetch current state from ServiceNow
    remote, err := s.snClient.GetStatement(ctx, stmt.ServiceNowSysID)
    if err != nil {
        return nil, err
    }

    // Compare sys_updated_on
    if remote.SysUpdatedOn.After(stmt.ServiceNowUpdatedAt) {
        return &ConflictInfo{
            StatementID:      stmt.ID,
            ControlNumber:    stmt.ControlNumber,
            LocalModifiedAt:  stmt.LocalModifiedAt,
            RemoteModifiedAt: remote.SysUpdatedOn,
            LocalContent:     stmt.ContentHTML,
            RemoteContent:    remote.Content,
        }, nil
    }

    return nil, nil
}
```

### Push Executor

```go
// internal/domain/push/executor.go

func (s *PushService) executePush(ctx context.Context, job *PushJob, req *PushRequest) {
    defer s.finalizePush(ctx, job)

    // Semaphore for concurrent requests (limit to 3)
    sem := make(chan struct{}, 3)
    results := make(chan PushItemResult, len(req.StatementIDs))

    for _, stmtID := range req.StatementIDs {
        sem <- struct{}{}
        go func(id uuid.UUID) {
            defer func() { <-sem }()
            result := s.pushStatement(ctx, id, req.ConflictResolutions[id])
            results <- result
        }(stmtID)
    }

    // Collect results
    for i := 0; i < len(req.StatementIDs); i++ {
        result := <-results
        s.recordResult(ctx, job, result)
    }
}

func (s *PushService) pushStatement(ctx context.Context, stmtID uuid.UUID, resolution ConflictResolution) PushItemResult {
    stmt, err := s.statementRepo.Get(ctx, stmtID)
    if err != nil {
        return PushItemResult{StatementID: stmtID, Result: "failed", Error: err.Error()}
    }

    // Check if should skip
    if resolution == ResolutionSkip {
        return PushItemResult{StatementID: stmtID, Result: "skipped"}
    }

    // Push to ServiceNow
    err = s.snClient.UpdateStatement(ctx, stmt.ServiceNowSysID, stmt.ContentHTML)
    if err != nil {
        return PushItemResult{StatementID: stmtID, Result: "failed", Error: err.Error()}
    }

    // Clear local modified flag
    s.statementRepo.ClearModified(ctx, stmtID)

    return PushItemResult{StatementID: stmtID, Result: "success"}
}
```

### ServiceNow Client Push Method

```go
// internal/infrastructure/servicenow/client.go

func (c *client) UpdateStatement(ctx context.Context, sysID string, content string) error {
    url := fmt.Sprintf("%s/api/now/table/sn_compliance_policy_statement/%s", c.baseURL, sysID)

    body := map[string]string{"statement": content}
    jsonBody, _ := json.Marshal(body)

    req, _ := retryablehttp.NewRequest("PUT", url, bytes.NewReader(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    c.auth.AddAuth(req)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("servicenow request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == 409 {
        return ErrConflict
    }
    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("servicenow error %d: %s", resp.StatusCode, body)
    }

    return nil
}
```

---

## 7. State Management

### Frontend Push State

```typescript
// src/features/push/stores/pushStore.ts

interface PushStore {
  // Preview state
  preview: PushPreview | null;
  selectedStatements: Set<string>;
  conflictResolutions: Record<string, ConflictResolution>;

  // Push state
  pushJob: PushJob | null;
  pushStatus: PushStatus | null;

  // Actions
  setPreview: (preview: PushPreview) => void;
  toggleStatement: (id: string) => void;
  setConflictResolution: (id: string, resolution: ConflictResolution) => void;
  startPush: () => void;
  clearPush: () => void;
}
```

### Push Modal Flow

```typescript
// src/features/push/components/PushModal.tsx

export function PushModal({ systemId, onClose }: Props) {
  const { data: preview, isLoading } = usePushPreview(systemId);
  const pushMutation = usePushMutation(systemId);
  const [step, setStep] = useState<'preview' | 'conflicts' | 'progress' | 'results'>('preview');

  // Step logic
  useEffect(() => {
    if (preview?.hasConflicts && step === 'preview') {
      setStep('conflicts');
    }
  }, [preview]);

  useEffect(() => {
    if (pushMutation.isSuccess) {
      setStep('progress');
    }
  }, [pushMutation.isSuccess]);

  return (
    <Dialog open onOpenChange={onClose}>
      {step === 'preview' && <PushPreviewStep preview={preview} onConfirm={handlePush} />}
      {step === 'conflicts' && <ConflictResolutionStep conflicts={preview.conflicts} />}
      {step === 'progress' && <PushProgressStep jobId={pushMutation.data.push_id} />}
      {step === 'results' && <PushResultsStep jobId={pushMutation.data.push_id} />}
    </Dialog>
  );
}
```

---

## 8. Security Considerations

### Authorization
- Only users who modified statements can push them
- Admin can push any statement
- Push requires write access to ServiceNow

### Audit Trail
- All push operations logged to sync_log
- Individual statement results logged
- Conflict resolutions recorded

### Data Validation
- Content sanitized before push (same as save)
- ServiceNow sys_id validated
- Version checks prevent stale writes

---

## 9. Performance & Scalability

### Performance Targets

| Operation | Target | Approach |
|-----------|--------|----------|
| Push preview | < 5s | Parallel conflict checks |
| Push single statement | < 5s | Direct API call |
| Push 50 statements | < 2 min | Controlled parallelism |

### Optimization Strategies

1. **Parallel Conflict Detection**
   ```go
   // Check conflicts in parallel (max 5 concurrent)
   sem := make(chan struct{}, 5)
   for _, stmt := range statements {
       sem <- struct{}{}
       go func(s Statement) {
           defer func() { <-sem }()
           checkConflict(s)
       }(stmt)
   }
   ```

2. **Batch Progress Updates**
   ```go
   // Update progress every 5 statements, not every one
   const progressBatchSize = 5
   if processed % progressBatchSize == 0 {
       updateProgress(job)
   }
   ```

3. **Connection Reuse**
   - HTTP client with keep-alive
   - Connection pool for ServiceNow

---

## 10. Testing Strategy

### Unit Tests

```go
// internal/domain/push/service_test.go

func TestPushService_CheckConflict(t *testing.T) {
    tests := []struct {
        name           string
        localUpdatedAt time.Time
        remoteUpdatedAt time.Time
        expectConflict bool
    }{
        {"no conflict", time.Now(), time.Now().Add(-1*time.Hour), false},
        {"has conflict", time.Now().Add(-1*time.Hour), time.Now(), true},
    }
    // ...
}

func TestPushService_PushStatement(t *testing.T) {
    // Test successful push
    // Test ServiceNow error handling
    // Test conflict resolution
}
```

### Integration Tests

```go
func TestPushService_FullPushFlow(t *testing.T) {
    // Setup: Modified statements, mock ServiceNow
    // Execute: Full push
    // Verify: Results, local state cleared, audit logged
}
```

### E2E Tests

```typescript
test('user can push changes', async ({ page }) => {
  // Modify a statement
  await page.goto('/workspace/system-1/AC-1');
  await page.fill('.editor', 'New content');
  await page.click('button:has-text("Save")');

  // Open push modal
  await page.click('button:has-text("Push to ServiceNow")');

  // Confirm and push
  await expect(page.locator('.push-preview')).toContainText('1 statement');
  await page.click('button:has-text("Confirm Push")');

  // Wait for completion
  await expect(page.locator('.push-results')).toContainText('1 succeeded');
});

test('user can resolve conflicts', async ({ page }) => {
  // Trigger conflict scenario
  // Verify conflict UI appears
  // Resolve and push
});
```

---

## 11. Deployment & DevOps

### Configuration

```bash
# Push settings
PUSH_MAX_CONCURRENT=3
PUSH_TIMEOUT_PER_STATEMENT=30s
PUSH_RETRY_MAX=3
PUSH_RETRY_DELAY=1s
```

### Monitoring

- Log push start/complete/fail
- Track push duration metrics
- Alert on high failure rate

---

## 12. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| ServiceNow rate limiting | Medium | Medium | Exponential backoff, limits |
| Data inconsistency | Low | High | Version checks, transactions |
| Partial push failure | Medium | Medium | Clear reporting, retry |
| Conflict resolution errors | Low | Medium | Clear UI, undo capability |

---

## 13. Development Phases

### Phase 1: Preview & Conflict Detection (3 days)
- Preview endpoint
- Conflict detection logic
- Diff generation

### Phase 2: Push Execution (3 days)
- Push service
- ServiceNow update client
- Result tracking

### Phase 3: Frontend (2 days)
- Push modal
- Progress display
- Conflict resolution UI

### Phase 4: Testing & Polish (2 days)
- Integration tests
- E2E tests
- Error handling refinement

**Estimated Total: 10 days**

---

## 14. Open Technical Questions

1. **Atomic Push:** Should we support all-or-nothing push (rollback on failure)?
2. **Push Approval:** Should pushes require approval from ISSO before executing?
3. **Conflict Auto-resolution:** Should we offer auto-merge for non-overlapping changes?
