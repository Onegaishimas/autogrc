# Technical Implementation Document: Control Package Push (F4)

**Document Version:** 1.0
**Date:** 2026-01-27
**Feature ID:** F4
**Related PRD:** 004_FPRD|Control_Package_Push.md
**Related TDD:** 004_FTDD|Control_Package_Push.md

---

## 1. Implementation Overview

### Summary
This TID provides implementation guidance for pushing locally modified control statements back to ServiceNow GRC, including conflict detection, selective push with preview, and comprehensive result reporting.

### Key Implementation Principles
- **Optimistic Concurrency:** Compare `sys_updated_on` to detect conflicts before push
- **Selective Push:** Users choose which modified statements to push
- **Atomic Operations:** Each statement push is independent (partial success allowed)
- **Conflict Resolution:** User must explicitly resolve conflicts before push

### Integration Points
- **F1 Dependency:** Uses connection config for ServiceNow authentication
- **F2 Dependency:** Uses sync state for conflict detection
- **F3 Dependency:** Reads modified statements from editor
- **F5 Forward:** Generates audit events for all push operations

---

## 2. File Structure and Organization

### Backend Files to Create/Modify

```
backend/
├── internal/
│   ├── api/
│   │   └── handlers/
│   │       └── push/
│   │           ├── handler.go            # HTTP handlers (create)
│   │           ├── handler_test.go       # Handler tests (create)
│   │           └── schemas.go            # Request/response DTOs (create)
│   ├── domain/
│   │   └── push/
│   │       ├── models.go                 # Push domain models (create)
│   │       ├── service.go                # Push business logic (create)
│   │       ├── service_test.go           # Service tests (create)
│   │       ├── conflict.go               # Conflict detection logic (create)
│   │       └── executor.go               # Push execution logic (create)
│   ├── jobs/
│   │   └── push/
│   │       ├── executor.go               # Background push job (create)
│   │       └── executor_test.go          # Executor tests (create)
│   └── infrastructure/
│       ├── database/
│       │   └── queries/
│       │       ├── statement.sql         # Add push queries (modify)
│       │       └── push_job.sql          # Push job queries (create)
│       └── servicenow/
│           └── client.go                 # Add PUT methods (modify)
└── migrations/
    └── 20260127_004_create_push_tables.sql  # Push jobs table (create)
```

### Frontend Files to Create

```
frontend/src/
├── features/
│   └── push/
│       ├── components/
│       │   ├── PushWorkflow.tsx          # Main push flow (create)
│       │   ├── ModifiedList.tsx          # List of modified statements (create)
│       │   ├── ConflictResolution.tsx    # Conflict diff viewer (create)
│       │   ├── PushPreview.tsx           # Pre-push preview (create)
│       │   ├── PushProgress.tsx          # Push progress indicator (create)
│       │   └── PushResults.tsx           # Results summary (create)
│       ├── hooks/
│       │   ├── useModifiedStatements.ts  # Query modified statements (create)
│       │   ├── usePush.ts                # Push mutation hooks (create)
│       │   ├── usePushStatus.ts          # Status polling hook (create)
│       │   └── useConflicts.ts           # Conflict detection hook (create)
│       ├── api/
│       │   └── pushApi.ts                # API client (create)
│       └── types.ts                      # TypeScript types (create)
└── pages/
    └── PushPage.tsx                      # Push workflow page (create)
```

---

## 3. Component Implementation Hints

### Conflict Detection Pattern

```go
// internal/domain/push/conflict.go
type ConflictChecker struct {
    snClient servicenow.Client
    stmtRepo statement.Repository
    logger   zerolog.Logger
}

type ConflictResult struct {
    StatementID   uuid.UUID
    HasConflict   bool
    LocalVersion  time.Time  // Our sys_updated_on
    RemoteVersion time.Time  // ServiceNow's current sys_updated_on
    LocalContent  string
    RemoteContent string
}

func (c *ConflictChecker) CheckConflicts(ctx context.Context, statementIDs []uuid.UUID) ([]ConflictResult, error) {
    results := make([]ConflictResult, 0, len(statementIDs))

    for _, stmtID := range statementIDs {
        stmt, err := c.stmtRepo.GetByID(ctx, stmtID)
        if err != nil {
            return nil, fmt.Errorf("failed to get statement %s: %w", stmtID, err)
        }

        // Fetch current state from ServiceNow
        snStmt, err := c.snClient.GetStatement(ctx, stmt.SNSysID)
        if err != nil {
            return nil, fmt.Errorf("failed to fetch ServiceNow statement %s: %w", stmt.SNSysID, err)
        }

        remoteUpdated, _ := time.Parse(time.RFC3339, snStmt.SysUpdatedOn)

        result := ConflictResult{
            StatementID:   stmtID,
            HasConflict:   remoteUpdated.After(stmt.SNUpdatedOn),
            LocalVersion:  stmt.SNUpdatedOn,
            RemoteVersion: remoteUpdated,
            LocalContent:  *stmt.LocalContent,
            RemoteContent: snStmt.Content,
        }
        results = append(results, result)
    }

    return results, nil
}
```

**Key Hints:**
- Check conflicts before showing push preview
- Compare timestamps to millisecond precision
- Fetch remote content for diff display
- Handle network errors gracefully (retry or fail the check)

### Push Executor Pattern

```go
// internal/domain/push/executor.go
type PushExecutor struct {
    snClient  servicenow.Client
    stmtRepo  statement.Repository
    auditRepo audit.Repository
    logger    zerolog.Logger
}

type PushResult struct {
    StatementID uuid.UUID
    Success     bool
    Error       *string
    PushedAt    *time.Time
}

func (e *PushExecutor) ExecutePush(ctx context.Context, job *PushJob) []PushResult {
    results := make([]PushResult, 0, len(job.StatementIDs))

    // Use semaphore for concurrency control
    sem := make(chan struct{}, 5) // Max 5 concurrent pushes

    var wg sync.WaitGroup
    var mu sync.Mutex

    for _, stmtID := range job.StatementIDs {
        wg.Add(1)
        go func(id uuid.UUID) {
            defer wg.Done()
            sem <- struct{}{}        // Acquire
            defer func() { <-sem }() // Release

            result := e.pushStatement(ctx, id, job.UserID)

            mu.Lock()
            results = append(results, result)
            mu.Unlock()
        }(stmtID)
    }

    wg.Wait()
    return results
}

func (e *PushExecutor) pushStatement(ctx context.Context, stmtID uuid.UUID, userID uuid.UUID) PushResult {
    stmt, err := e.stmtRepo.GetByID(ctx, stmtID)
    if err != nil {
        return PushResult{StatementID: stmtID, Success: false, Error: ptr(err.Error())}
    }

    // Push to ServiceNow
    err = e.snClient.UpdateStatement(ctx, stmt.SNSysID, *stmt.LocalContent)
    if err != nil {
        e.logger.Error().Err(err).Str("statement_id", stmtID.String()).Msg("push failed")
        return PushResult{StatementID: stmtID, Success: false, Error: ptr(err.Error())}
    }

    // Update local state
    now := time.Now()
    err = e.stmtRepo.MarkPushed(ctx, stmtID, now)
    if err != nil {
        e.logger.Error().Err(err).Msg("failed to mark statement as pushed")
    }

    // Record audit event
    e.auditRepo.RecordPush(ctx, audit.PushEvent{
        StatementID: stmtID,
        UserID:      userID,
        PushedAt:    now,
        Content:     *stmt.LocalContent,
    })

    return PushResult{StatementID: stmtID, Success: true, PushedAt: &now}
}
```

---

## 4. Database Implementation Approach

### Push Job Schema

```sql
-- migrations/20260127_004_create_push_tables.sql

CREATE TYPE push_status AS ENUM ('pending', 'running', 'completed', 'failed');

CREATE TABLE push_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status push_status NOT NULL DEFAULT 'pending',
    statement_ids UUID[] NOT NULL,
    results JSONB DEFAULT '[]',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_push_jobs_status ON push_jobs(status)
    WHERE status IN ('pending', 'running');
CREATE INDEX idx_push_jobs_user ON push_jobs(created_by);
```

### Statement Push Tracking

```sql
-- name: MarkStatementPushed :exec
UPDATE statements SET
    content = local_content,           -- Promote local to canonical
    content_html = local_content_html,
    local_content = NULL,              -- Clear local edits
    local_content_html = NULL,
    is_modified = false,
    sync_status = 'synced',
    sn_updated_on = $2,                -- Updated timestamp from ServiceNow response
    updated_at = NOW()
WHERE id = $1;

-- name: GetModifiedStatements :many
SELECT s.*, c.control_id, c.control_name, sys.name as system_name
FROM statements s
JOIN controls c ON s.control_id = c.id
JOIN systems sys ON c.system_id = sys.id
WHERE s.is_modified = true
ORDER BY sys.name, c.control_id;

-- name: GetConflictedStatements :many
SELECT * FROM statements
WHERE sync_status = 'conflict'
ORDER BY updated_at DESC;
```

---

## 5. API Implementation Strategy

### Endpoints

```go
// internal/api/handlers/push/handler.go
func (h *Handler) RegisterRoutes(r chi.Router) {
    r.Route("/api/v1/push", func(r chi.Router) {
        // Pre-push operations
        r.Get("/modified", h.GetModifiedStatements)
        r.Post("/check-conflicts", h.CheckConflicts)
        r.Post("/resolve-conflict/{id}", h.ResolveConflict)

        // Push operations
        r.Post("/", h.StartPush)
        r.Get("/{jobId}", h.GetPushStatus)
        r.Get("/{jobId}/results", h.GetPushResults)
    })
}
```

### Request/Response Schemas

```go
// internal/api/handlers/push/schemas.go

type ModifiedStatementsResponse struct {
    Statements []ModifiedStatement `json:"statements"`
    TotalCount int                 `json:"total_count"`
}

type ModifiedStatement struct {
    ID           uuid.UUID `json:"id"`
    ControlID    string    `json:"control_id"`
    ControlName  string    `json:"control_name"`
    SystemName   string    `json:"system_name"`
    ModifiedAt   time.Time `json:"modified_at"`
    SyncStatus   string    `json:"sync_status"` // modified, conflict
    ContentPreview string  `json:"content_preview"` // First 200 chars
}

type CheckConflictsRequest struct {
    StatementIDs []uuid.UUID `json:"statement_ids" validate:"required,min=1"`
}

type CheckConflictsResponse struct {
    Conflicts []ConflictInfo `json:"conflicts"`
}

type ConflictInfo struct {
    StatementID    uuid.UUID `json:"statement_id"`
    ControlID      string    `json:"control_id"`
    HasConflict    bool      `json:"has_conflict"`
    LocalContent   string    `json:"local_content"`
    RemoteContent  string    `json:"remote_content"`
    LocalUpdatedAt time.Time `json:"local_updated_at"`
    RemoteUpdatedAt time.Time `json:"remote_updated_at"`
}

type ResolveConflictRequest struct {
    Resolution string `json:"resolution" validate:"required,oneof=keep_local keep_remote"`
}

type StartPushRequest struct {
    StatementIDs []uuid.UUID `json:"statement_ids" validate:"required,min=1,max=100"`
}

type StartPushResponse struct {
    JobID uuid.UUID `json:"job_id"`
}

type PushStatusResponse struct {
    JobID       uuid.UUID     `json:"job_id"`
    Status      string        `json:"status"`
    Total       int           `json:"total"`
    Completed   int           `json:"completed"`
    Succeeded   int           `json:"succeeded"`
    Failed      int           `json:"failed"`
    StartedAt   *time.Time    `json:"started_at,omitempty"`
    CompletedAt *time.Time    `json:"completed_at,omitempty"`
}

type PushResultsResponse struct {
    Results []PushResultItem `json:"results"`
}

type PushResultItem struct {
    StatementID uuid.UUID  `json:"statement_id"`
    ControlID   string     `json:"control_id"`
    Success     bool       `json:"success"`
    Error       *string    `json:"error,omitempty"`
    PushedAt    *time.Time `json:"pushed_at,omitempty"`
}
```

---

## 6. Frontend Implementation Approach

### Modified Statements List

```typescript
// features/push/components/ModifiedList.tsx
interface ModifiedListProps {
  onSelect: (ids: string[]) => void;
  selectedIds: string[];
}

export function ModifiedList({ onSelect, selectedIds }: ModifiedListProps) {
  const { data: statements, isLoading } = useModifiedStatements();

  const toggleSelect = (id: string) => {
    const next = selectedIds.includes(id)
      ? selectedIds.filter(x => x !== id)
      : [...selectedIds, id];
    onSelect(next);
  };

  const selectAll = () => {
    if (!statements) return;
    const allIds = statements.filter(s => s.syncStatus !== 'conflict').map(s => s.id);
    onSelect(allIds);
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-medium">Modified Statements</h3>
        <Button variant="outline" size="sm" onClick={selectAll}>
          Select All (Non-Conflicted)
        </Button>
      </div>

      <div className="space-y-2">
        {statements?.map((stmt) => (
          <Card
            key={stmt.id}
            className={cn(
              "cursor-pointer transition-colors",
              selectedIds.includes(stmt.id) && "ring-2 ring-primary",
              stmt.syncStatus === 'conflict' && "border-destructive"
            )}
            onClick={() => toggleSelect(stmt.id)}
          >
            <CardContent className="p-4">
              <div className="flex items-start justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <Checkbox
                      checked={selectedIds.includes(stmt.id)}
                      disabled={stmt.syncStatus === 'conflict'}
                    />
                    <span className="font-medium">{stmt.controlId}</span>
                    <span className="text-muted-foreground">{stmt.controlName}</span>
                  </div>
                  <p className="text-sm text-muted-foreground mt-1">
                    {stmt.systemName} • Modified {formatDistanceToNow(new Date(stmt.modifiedAt))} ago
                  </p>
                </div>
                {stmt.syncStatus === 'conflict' && (
                  <Badge variant="destructive">Conflict</Badge>
                )}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
```

### Conflict Resolution Component

```typescript
// features/push/components/ConflictResolution.tsx
import { DiffEditor } from '@monaco-editor/react';

interface ConflictResolutionProps {
  conflict: ConflictInfo;
  onResolve: (resolution: 'keep_local' | 'keep_remote') => void;
}

export function ConflictResolution({ conflict, onResolve }: ConflictResolutionProps) {
  return (
    <Dialog>
      <DialogContent className="max-w-4xl">
        <DialogHeader>
          <DialogTitle>Resolve Conflict: {conflict.controlId}</DialogTitle>
          <DialogDescription>
            This statement was modified both locally and in ServiceNow.
            Choose which version to keep.
          </DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-2 gap-4 text-sm mb-4">
          <div>
            <span className="font-medium">Local:</span>
            <span className="text-muted-foreground ml-2">
              {format(new Date(conflict.localUpdatedAt), 'PPpp')}
            </span>
          </div>
          <div>
            <span className="font-medium">ServiceNow:</span>
            <span className="text-muted-foreground ml-2">
              {format(new Date(conflict.remoteUpdatedAt), 'PPpp')}
            </span>
          </div>
        </div>

        <div className="h-[400px] border rounded">
          <DiffEditor
            original={conflict.remoteContent}
            modified={conflict.localContent}
            language="html"
            options={{
              readOnly: true,
              renderSideBySide: true,
              minimap: { enabled: false },
            }}
          />
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onResolve('keep_remote')}>
            Keep ServiceNow Version
          </Button>
          <Button onClick={() => onResolve('keep_local')}>
            Keep Local Version
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
```

### Push Workflow Component

```typescript
// features/push/components/PushWorkflow.tsx
type Step = 'select' | 'conflicts' | 'preview' | 'progress' | 'results';

export function PushWorkflow() {
  const [step, setStep] = useState<Step>('select');
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [jobId, setJobId] = useState<string | null>(null);

  const conflictsQuery = useConflicts(selectedIds, { enabled: step === 'conflicts' });
  const pushMutation = useStartPush();

  const handleCheckConflicts = async () => {
    setStep('conflicts');
  };

  const handleStartPush = async () => {
    const result = await pushMutation.mutateAsync({ statementIds: selectedIds });
    setJobId(result.jobId);
    setStep('progress');
  };

  return (
    <div className="space-y-6">
      <Stepper activeStep={step} steps={['Select', 'Conflicts', 'Preview', 'Push', 'Results']} />

      {step === 'select' && (
        <>
          <ModifiedList selectedIds={selectedIds} onSelect={setSelectedIds} />
          <div className="flex justify-end">
            <Button
              disabled={selectedIds.length === 0}
              onClick={handleCheckConflicts}
            >
              Check for Conflicts
            </Button>
          </div>
        </>
      )}

      {step === 'conflicts' && (
        <ConflictStep
          conflicts={conflictsQuery.data?.conflicts ?? []}
          onResolved={() => setStep('preview')}
          onBack={() => setStep('select')}
        />
      )}

      {step === 'preview' && (
        <PushPreview
          statementIds={selectedIds}
          onConfirm={handleStartPush}
          onBack={() => setStep('select')}
        />
      )}

      {step === 'progress' && jobId && (
        <PushProgress
          jobId={jobId}
          onComplete={() => setStep('results')}
        />
      )}

      {step === 'results' && jobId && (
        <PushResults
          jobId={jobId}
          onDone={() => {
            setStep('select');
            setSelectedIds([]);
            setJobId(null);
          }}
        />
      )}
    </div>
  );
}
```

---

## 7. Business Logic Implementation Hints

### ServiceNow Update API

```go
// internal/infrastructure/servicenow/client.go

func (c *Client) UpdateStatement(ctx context.Context, sysID string, content string) error {
    url := fmt.Sprintf("%s/api/now/table/sn_compliance_statement/%s", c.baseURL, sysID)

    payload := map[string]interface{}{
        "u_implementation_statement": content, // Field name may vary by instance
    }

    body, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal payload: %w", err)
    }

    req, err := retryablehttp.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    c.auth.ApplyAuth(req)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == 404 {
        return ErrStatementNotFound
    }
    if resp.StatusCode != 200 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
    }

    return nil
}
```

### Conflict Resolution Logic

```go
// internal/domain/push/service.go

func (s *Service) ResolveConflict(ctx context.Context, stmtID uuid.UUID, resolution string, userID uuid.UUID) error {
    stmt, err := s.stmtRepo.GetByID(ctx, stmtID)
    if err != nil {
        return fmt.Errorf("failed to get statement: %w", err)
    }

    if stmt.SyncStatus != "conflict" {
        return fmt.Errorf("statement is not in conflict state")
    }

    switch resolution {
    case "keep_local":
        // Keep local content, clear conflict status, ready for push
        err = s.stmtRepo.ClearConflict(ctx, stmtID)

    case "keep_remote":
        // Fetch current ServiceNow content and overwrite local
        snStmt, err := s.snClient.GetStatement(ctx, stmt.SNSysID)
        if err != nil {
            return fmt.Errorf("failed to fetch remote: %w", err)
        }
        err = s.stmtRepo.OverwriteLocal(ctx, stmtID, snStmt.Content, snStmt.SysUpdatedOn)

    default:
        return fmt.Errorf("invalid resolution: %s", resolution)
    }

    // Record resolution in audit log
    s.auditRepo.RecordConflictResolution(ctx, audit.ConflictResolutionEvent{
        StatementID: stmtID,
        UserID:      userID,
        Resolution:  resolution,
        ResolvedAt:  time.Now(),
    })

    return err
}
```

---

## 8. Testing Implementation Approach

### Conflict Detection Tests

```go
// internal/domain/push/conflict_test.go
func TestConflictChecker_CheckConflicts(t *testing.T) {
    tests := []struct {
        name            string
        localUpdatedOn  time.Time
        remoteUpdatedOn string
        wantConflict    bool
    }{
        {
            name:            "no conflict when remote unchanged",
            localUpdatedOn:  time.Date(2026, 1, 27, 10, 0, 0, 0, time.UTC),
            remoteUpdatedOn: "2026-01-27T10:00:00Z",
            wantConflict:    false,
        },
        {
            name:            "conflict when remote is newer",
            localUpdatedOn:  time.Date(2026, 1, 27, 10, 0, 0, 0, time.UTC),
            remoteUpdatedOn: "2026-01-27T11:00:00Z",
            wantConflict:    true,
        },
        {
            name:            "no conflict when local is newer",
            localUpdatedOn:  time.Date(2026, 1, 27, 12, 0, 0, 0, time.UTC),
            remoteUpdatedOn: "2026-01-27T10:00:00Z",
            wantConflict:    false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks and run test...
        })
    }
}
```

### Push Executor Tests

```go
// internal/domain/push/executor_test.go
func TestPushExecutor_ExecutePush(t *testing.T) {
    t.Run("successful push updates local state", func(t *testing.T) {
        mockSN := &MockSNClient{}
        mockSN.On("UpdateStatement", mock.Anything, "sn-sys-1", mock.Anything).Return(nil)

        mockRepo := &MockStmtRepo{}
        mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(&Statement{
            ID:           uuid.New(),
            SNSysID:      "sn-sys-1",
            LocalContent: ptr("<p>Updated</p>"),
        }, nil)
        mockRepo.On("MarkPushed", mock.Anything, mock.Anything, mock.Anything).Return(nil)

        executor := NewPushExecutor(mockSN, mockRepo, mockAudit, logger)
        job := &PushJob{StatementIDs: []uuid.UUID{uuid.New()}}

        results := executor.ExecutePush(context.Background(), job)

        assert.Len(t, results, 1)
        assert.True(t, results[0].Success)
        mockRepo.AssertCalled(t, "MarkPushed", mock.Anything, mock.Anything, mock.Anything)
    })

    t.Run("failed push records error", func(t *testing.T) {
        mockSN := &MockSNClient{}
        mockSN.On("UpdateStatement", mock.Anything, mock.Anything, mock.Anything).
            Return(errors.New("network error"))

        // ... setup and assertions
    })
}
```

### Frontend Tests

```typescript
// features/push/components/ConflictResolution.test.tsx
describe('ConflictResolution', () => {
  it('shows diff between local and remote content', () => {
    const conflict = {
      statementId: 'stmt-1',
      localContent: '<p>Local version</p>',
      remoteContent: '<p>Remote version</p>',
    };

    render(<ConflictResolution conflict={conflict} onResolve={jest.fn()} />);

    expect(screen.getByText('Local version')).toBeInTheDocument();
    expect(screen.getByText('Remote version')).toBeInTheDocument();
  });

  it('calls onResolve with keep_local when keeping local', async () => {
    const onResolve = jest.fn();
    render(<ConflictResolution conflict={mockConflict} onResolve={onResolve} />);

    await userEvent.click(screen.getByRole('button', { name: /keep local/i }));

    expect(onResolve).toHaveBeenCalledWith('keep_local');
  });
});
```

---

## 9. Error Handling Strategy

### Push Error Types

```go
// internal/domain/push/errors.go
var (
    ErrStatementNotFound = errors.New("statement not found in ServiceNow")
    ErrConflictUnresolved = errors.New("statement has unresolved conflict")
    ErrAuthExpired = errors.New("ServiceNow authentication expired")
    ErrRateLimited = errors.New("ServiceNow rate limit exceeded")
)

func categorizeError(err error) (userMessage string, retryable bool) {
    switch {
    case errors.Is(err, ErrStatementNotFound):
        return "Statement no longer exists in ServiceNow", false
    case errors.Is(err, ErrAuthExpired):
        return "ServiceNow session expired, please reconnect", false
    case errors.Is(err, ErrRateLimited):
        return "ServiceNow rate limit reached, try again later", true
    default:
        return "Failed to push statement", true
    }
}
```

### User-Facing Error Display

```typescript
// features/push/components/PushResults.tsx
function getErrorDisplay(error: string | undefined): { message: string; severity: 'warning' | 'error' } {
  if (!error) return { message: '', severity: 'error' };

  if (error.includes('rate limit')) {
    return { message: 'Rate limited - retry in a few minutes', severity: 'warning' };
  }
  if (error.includes('not found')) {
    return { message: 'Statement deleted in ServiceNow', severity: 'error' };
  }
  return { message: 'Push failed - see logs for details', severity: 'error' };
}
```

---

## 10. Performance Implementation Hints

### Concurrent Push Limits

```go
// internal/domain/push/executor.go
const (
    maxConcurrentPushes = 5   // Prevent overwhelming ServiceNow
    pushTimeout         = 30 * time.Second
)

func (e *PushExecutor) pushWithTimeout(ctx context.Context, stmtID uuid.UUID) PushResult {
    ctx, cancel := context.WithTimeout(ctx, pushTimeout)
    defer cancel()

    // Push logic...
}
```

### Progress Update Batching

```go
// Update job progress every N completions, not every statement
const progressUpdateBatch = 10

func (e *PushExecutor) ExecutePush(ctx context.Context, job *PushJob) []PushResult {
    var completed int
    for _, stmtID := range job.StatementIDs {
        result := e.pushStatement(ctx, stmtID, job.UserID)
        results = append(results, result)

        completed++
        if completed%progressUpdateBatch == 0 {
            e.updateJobProgress(job.ID, completed, len(results))
        }
    }
}
```

---

## 11. Integration with Audit Trail (F5)

### Audit Event Recording

```go
// Record push event for audit trail
e.auditRepo.RecordEvent(ctx, audit.Event{
    Type:        audit.EventTypePush,
    EntityType:  "statement",
    EntityID:    stmtID.String(),
    UserID:      userID,
    Action:      "push",
    Details: map[string]interface{}{
        "content_pushed":  content,
        "sn_sys_id":       snSysID,
        "previous_status": previousStatus,
    },
    Timestamp: time.Now(),
})
```

---

## 12. Security Checklist

- [ ] Verify user has write access to statements being pushed
- [ ] Validate statement IDs are UUIDs
- [ ] Limit max statements per push request (100)
- [ ] Rate limit push requests per user
- [ ] Log all push operations for audit
- [ ] Sanitize content before push (prevent XSS in ServiceNow)
- [ ] Handle ServiceNow auth token expiration gracefully
- [ ] Verify connection is active before allowing push
