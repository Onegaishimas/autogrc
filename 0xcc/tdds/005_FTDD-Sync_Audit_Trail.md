# Technical Design Document: Sync Audit Trail (F5)

**Document Version:** 1.0
**Date:** 2026-01-27
**Feature ID:** F5
**Related PRD:** 005_FPRD|Sync_Audit_Trail.md

---

## 1. Executive Summary

This TDD defines the technical architecture for the comprehensive sync audit trail system. The feature provides logging, querying, and export capabilities for all synchronization operations, supporting compliance requirements and troubleshooting needs.

**Key Technical Decisions:**
- Append-only audit tables (immutable)
- Efficient indexing for time-range queries
- Background export generation for large datasets
- Separate detail tables to optimize query performance

---

## 2. System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           FRONTEND                                       │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                      Audit Log Page                                 ││
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────┐ ││
│  │  │ Filter Panel    │  │ Log Table       │  │ Export Panel        │ ││
│  │  │ - Date range    │  │ - Expandable    │  │ - Format select     │ ││
│  │  │ - Type          │  │ - Paginated     │  │ - Download          │ ││
│  │  │ - Status        │  │ - Sortable      │  │                     │ ││
│  │  └─────────────────┘  └─────────────────┘  └─────────────────────┘ ││
│  └─────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────┘
            │                         │
            ▼                         ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           BACKEND (Go)                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────┐ │
│  │ Audit           │  │ Export          │  │ Statement History       │ │
│  │ Query Handler   │  │ Handler         │  │ Handler                 │ │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────────────┘ │
│           │                    │                    │                   │
│           ▼                    ▼                    ▼                   │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                      Audit Service                                  ││
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                ││
│  │  │ Log         │  │ Query       │  │ Export      │                ││
│  │  │ Writer      │  │ Builder     │  │ Generator   │                ││
│  │  └─────────────┘  └─────────────┘  └─────────────┘                ││
│  └─────────────────────────────────────────────────────────────────────┘│
│           │                                                             │
│           ▼                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                      PostgreSQL                                     ││
│  │  sync_log | sync_log_details | statement_change_log | exports      ││
│  └─────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────┘
```

### Logging Integration Points

```
Pull Service (F2) ──────► Audit Service ──────► sync_log
       │                        │
       ▼                        ▼
 For each control ────► sync_log_details

Push Service (F4) ──────► Audit Service ──────► sync_log
       │                        │
       ▼                        ▼
 For each statement ───► sync_log_details
                               │
                               ▼
                        statement_change_log
```

---

## 3. Technical Stack

### Backend Components

| Component | Technology | Justification |
|-----------|------------|---------------|
| Query Building | squirrel | SQL builder for dynamic queries |
| CSV Export | encoding/csv | Go stdlib |
| PDF Export | jung-kurt/gofpdf | Pure Go PDF generation |
| Background Jobs | Go routines + channels | Simple, in-process |

### Frontend Components

| Component | Technology | Justification |
|-----------|------------|---------------|
| Data Table | TanStack Table | Virtualized, sortable tables |
| Date Picker | react-day-picker | Date range selection |
| File Download | file-saver | Cross-browser downloads |

---

## 4. Data Design

### Database Schema

```sql
-- migrations/20260127_005_create_audit_tables.up.sql

-- Main sync log (already partially created, completing here)
CREATE TABLE IF NOT EXISTS sync_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    system_id UUID NOT NULL REFERENCES systems(id),
    system_name VARCHAR(255) NOT NULL,  -- Denormalized for queries
    operation VARCHAR(10) NOT NULL CHECK (operation IN ('pull', 'push')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('started', 'completed', 'failed', 'partial')),

    -- Counts
    records_total INTEGER DEFAULT 0,
    records_succeeded INTEGER DEFAULT 0,
    records_failed INTEGER DEFAULT 0,
    records_skipped INTEGER DEFAULT 0,

    -- Timing
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    duration_ms INTEGER,

    -- Error info
    error_message TEXT,
    error_details JSONB,

    -- User info
    user_id UUID REFERENCES users(id),
    user_email VARCHAR(255) NOT NULL,  -- Denormalized

    -- Immutable
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_sync_log_system_time ON sync_log(system_id, started_at DESC);
CREATE INDEX idx_sync_log_user_time ON sync_log(user_id, started_at DESC);
CREATE INDEX idx_sync_log_status ON sync_log(status) WHERE status IN ('failed', 'partial');
CREATE INDEX idx_sync_log_operation ON sync_log(operation, started_at DESC);
CREATE INDEX idx_sync_log_time ON sync_log(started_at DESC);

-- Detailed per-item results
CREATE TABLE sync_log_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sync_log_id UUID NOT NULL REFERENCES sync_log(id) ON DELETE CASCADE,

    -- Item identification
    statement_id UUID REFERENCES implementation_statements(id),
    control_number VARCHAR(20) NOT NULL,  -- Denormalized

    -- Result
    result VARCHAR(20) NOT NULL CHECK (result IN ('success', 'failed', 'skipped', 'conflict')),
    error_code VARCHAR(50),
    error_message TEXT,

    -- Change tracking (for push)
    content_hash_before VARCHAR(64),
    content_hash_after VARCHAR(64),

    -- ServiceNow response
    servicenow_sys_id VARCHAR(32),
    servicenow_response JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sync_details_log ON sync_log_details(sync_log_id);
CREATE INDEX idx_sync_details_statement ON sync_log_details(statement_id);
CREATE INDEX idx_sync_details_control ON sync_log_details(control_number);

-- Statement change history (comprehensive change log)
CREATE TABLE statement_change_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    statement_id UUID NOT NULL REFERENCES implementation_statements(id) ON DELETE CASCADE,
    control_number VARCHAR(20) NOT NULL,  -- Denormalized

    change_type VARCHAR(20) NOT NULL CHECK (change_type IN ('created', 'updated', 'deleted', 'pulled', 'pushed')),
    source VARCHAR(20) NOT NULL CHECK (source IN ('local_edit', 'pull', 'push', 'system')),

    -- Content snapshots (optional, for important changes)
    content_before TEXT,
    content_after TEXT,
    content_hash_before VARCHAR(64),
    content_hash_after VARCHAR(64),

    -- Attribution
    changed_by UUID REFERENCES users(id),
    changed_by_email VARCHAR(255),
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Link to sync operation
    sync_log_id UUID REFERENCES sync_log(id)
);

CREATE INDEX idx_statement_changes_stmt ON statement_change_log(statement_id, changed_at DESC);
CREATE INDEX idx_statement_changes_time ON statement_change_log(changed_at DESC);

-- Export tracking
CREATE TABLE audit_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    format VARCHAR(10) NOT NULL CHECK (format IN ('csv', 'pdf')),

    -- Query parameters
    filters JSONB NOT NULL,

    -- Result
    file_path VARCHAR(255),
    file_size_bytes BIGINT,
    record_count INTEGER,

    -- Timing
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,

    -- Errors
    error_message TEXT,

    -- User
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_exports_user ON audit_exports(created_by, created_at DESC);
CREATE INDEX idx_exports_status ON audit_exports(status) WHERE status = 'processing';
```

### Data Retention Policy

```sql
-- Retention: 7 years (per compliance requirements)
-- Archive strategy: Move to partitioned archive tables after 2 years

-- Future: Partition by month for large-scale deployments
-- CREATE TABLE sync_log_2026_01 PARTITION OF sync_log
--     FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
```

---

## 5. API Design

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/audit/sync | Query sync logs |
| GET | /api/v1/audit/sync/:id | Get sync log details |
| GET | /api/v1/audit/statements/:id/history | Get statement history |
| POST | /api/v1/audit/export | Start export job |
| GET | /api/v1/audit/export/:id | Get export status |
| GET | /api/v1/audit/export/:id/download | Download export file |

### Query Sync Logs

```go
// GET /api/v1/audit/sync

type SyncLogQuery struct {
    // Filters
    SystemID  *uuid.UUID `query:"system_id"`
    UserID    *uuid.UUID `query:"user_id"`
    Operation *string    `query:"operation"` // pull, push
    Status    *string    `query:"status"`    // success, failed, partial
    StartDate *time.Time `query:"start_date"`
    EndDate   *time.Time `query:"end_date"`
    Search    *string    `query:"search"`    // Search error messages

    // Pagination
    Page    int `query:"page" default:"1"`
    PerPage int `query:"per_page" default:"20" max:"100"`

    // Sorting
    SortBy    string `query:"sort_by" default:"started_at"`
    SortOrder string `query:"sort_order" default:"desc"`
}

type SyncLogResponse struct {
    Data []SyncLogEntry `json:"data"`
    Meta PaginationMeta `json:"meta"`
}

type SyncLogEntry struct {
    ID              uuid.UUID  `json:"id"`
    SystemName      string     `json:"system_name"`
    Operation       string     `json:"operation"`
    Status          string     `json:"status"`
    RecordsTotal    int        `json:"records_total"`
    RecordsSucceeded int       `json:"records_succeeded"`
    RecordsFailed   int        `json:"records_failed"`
    StartedAt       time.Time  `json:"started_at"`
    CompletedAt     *time.Time `json:"completed_at"`
    DurationMs      *int       `json:"duration_ms"`
    UserEmail       string     `json:"user_email"`
    ErrorMessage    *string    `json:"error_message,omitempty"`
}
```

### Get Sync Log Details

```go
// GET /api/v1/audit/sync/:id

type SyncLogDetailResponse struct {
    // Summary
    ID            uuid.UUID   `json:"id"`
    SystemName    string      `json:"system_name"`
    Operation     string      `json:"operation"`
    Status        string      `json:"status"`
    StartedAt     time.Time   `json:"started_at"`
    CompletedAt   *time.Time  `json:"completed_at"`
    DurationMs    *int        `json:"duration_ms"`
    UserEmail     string      `json:"user_email"`

    // Counts
    RecordsTotal     int `json:"records_total"`
    RecordsSucceeded int `json:"records_succeeded"`
    RecordsFailed    int `json:"records_failed"`
    RecordsSkipped   int `json:"records_skipped"`

    // Item details
    Items []SyncLogDetailItem `json:"items"`

    // Error info
    ErrorMessage *string `json:"error_message,omitempty"`
}

type SyncLogDetailItem struct {
    ControlNumber string  `json:"control_number"`
    Result        string  `json:"result"`
    ErrorMessage  *string `json:"error_message,omitempty"`
}
```

### Export

```go
// POST /api/v1/audit/export

type ExportRequest struct {
    Format    string         `json:"format" validate:"required,oneof=csv pdf"`
    Filters   SyncLogQuery   `json:"filters"`
}

type ExportResponse struct {
    ExportID  uuid.UUID `json:"export_id"`
    Status    string    `json:"status"`
    EstimatedRecords int `json:"estimated_records"`
}

// GET /api/v1/audit/export/:id

type ExportStatusResponse struct {
    ExportID     uuid.UUID  `json:"export_id"`
    Status       string     `json:"status"`
    RecordCount  *int       `json:"record_count,omitempty"`
    FileSizeBytes *int64    `json:"file_size_bytes,omitempty"`
    DownloadURL  *string    `json:"download_url,omitempty"`
    ExpiresAt    *time.Time `json:"expires_at,omitempty"`
    ErrorMessage *string    `json:"error_message,omitempty"`
}
```

---

## 6. Component Architecture

### Backend Package Structure

```
internal/
├── domain/
│   └── audit/
│       ├── service.go         # Query and export logic
│       ├── repository.go      # Database access
│       ├── logger.go          # Audit logging functions
│       ├── exporter.go        # Export generation
│       ├── csv_writer.go      # CSV formatting
│       ├── pdf_writer.go      # PDF formatting
│       └── service_test.go
└── api/
    └── handlers/
        └── audit/
            ├── handler.go     # HTTP handlers
            └── schemas.go     # Request/response types
```

### Audit Logger Integration

```go
// internal/domain/audit/logger.go

type AuditLogger struct {
    repo AuditRepository
}

// Called by Pull/Push services
func (l *AuditLogger) LogSyncStart(ctx context.Context, params LogStartParams) (*SyncLog, error) {
    entry := &SyncLog{
        ID:          uuid.New(),
        SystemID:    params.SystemID,
        SystemName:  params.SystemName,
        Operation:   params.Operation,
        Status:      "started",
        UserID:      params.UserID,
        UserEmail:   params.UserEmail,
        StartedAt:   time.Now(),
    }
    return l.repo.CreateSyncLog(ctx, entry)
}

func (l *AuditLogger) LogSyncComplete(ctx context.Context, logID uuid.UUID, results SyncResults) error {
    return l.repo.UpdateSyncLog(ctx, logID, map[string]interface{}{
        "status":            results.Status,
        "records_total":     results.Total,
        "records_succeeded": results.Succeeded,
        "records_failed":    results.Failed,
        "records_skipped":   results.Skipped,
        "completed_at":      time.Now(),
        "duration_ms":       results.DurationMs,
        "error_message":     results.ErrorMessage,
    })
}

func (l *AuditLogger) LogSyncDetail(ctx context.Context, logID uuid.UUID, detail SyncDetailParams) error {
    entry := &SyncLogDetail{
        ID:            uuid.New(),
        SyncLogID:     logID,
        StatementID:   detail.StatementID,
        ControlNumber: detail.ControlNumber,
        Result:        detail.Result,
        ErrorMessage:  detail.ErrorMessage,
    }
    return l.repo.CreateSyncLogDetail(ctx, entry)
}

func (l *AuditLogger) LogStatementChange(ctx context.Context, params StatementChangeParams) error {
    entry := &StatementChangeLog{
        ID:              uuid.New(),
        StatementID:     params.StatementID,
        ControlNumber:   params.ControlNumber,
        ChangeType:      params.ChangeType,
        Source:          params.Source,
        ContentBefore:   params.ContentBefore,
        ContentAfter:    params.ContentAfter,
        ChangedBy:       params.UserID,
        ChangedByEmail:  params.UserEmail,
        ChangedAt:       time.Now(),
        SyncLogID:       params.SyncLogID,
    }
    return l.repo.CreateStatementChangeLog(ctx, entry)
}
```

### Query Builder

```go
// internal/domain/audit/repository.go

func (r *repository) QuerySyncLogs(ctx context.Context, q SyncLogQuery) ([]SyncLogEntry, int, error) {
    builder := sq.Select(
        "id", "system_name", "operation", "status",
        "records_total", "records_succeeded", "records_failed",
        "started_at", "completed_at", "duration_ms",
        "user_email", "error_message",
    ).From("sync_log")

    // Apply filters
    if q.SystemID != nil {
        builder = builder.Where(sq.Eq{"system_id": *q.SystemID})
    }
    if q.UserID != nil {
        builder = builder.Where(sq.Eq{"user_id": *q.UserID})
    }
    if q.Operation != nil {
        builder = builder.Where(sq.Eq{"operation": *q.Operation})
    }
    if q.Status != nil {
        builder = builder.Where(sq.Eq{"status": *q.Status})
    }
    if q.StartDate != nil {
        builder = builder.Where(sq.GtOrEq{"started_at": *q.StartDate})
    }
    if q.EndDate != nil {
        builder = builder.Where(sq.LtOrEq{"started_at": *q.EndDate})
    }
    if q.Search != nil {
        builder = builder.Where(sq.ILike{"error_message": "%" + *q.Search + "%"})
    }

    // Count total
    countBuilder := builder.RemoveLimit().RemoveOffset()
    countBuilder = countBuilder.Column("COUNT(*)")
    // ... execute count

    // Apply pagination and sorting
    builder = builder.
        OrderBy(fmt.Sprintf("%s %s", q.SortBy, q.SortOrder)).
        Limit(uint64(q.PerPage)).
        Offset(uint64((q.Page - 1) * q.PerPage))

    // Execute query
    query, args, _ := builder.ToSql()
    rows, err := r.db.Query(ctx, query, args...)
    // ... scan rows

    return entries, total, nil
}
```

### Export Generator

```go
// internal/domain/audit/exporter.go

type Exporter struct {
    repo     AuditRepository
    storage  FileStorage
    logger   zerolog.Logger
}

func (e *Exporter) StartExport(ctx context.Context, userID uuid.UUID, format string, filters SyncLogQuery) (*AuditExport, error) {
    // Create export record
    export := &AuditExport{
        ID:        uuid.New(),
        Status:    "pending",
        Format:    format,
        Filters:   filters,
        CreatedBy: userID,
    }

    if err := e.repo.CreateExport(ctx, export); err != nil {
        return nil, err
    }

    // Start background job
    go e.executeExport(context.Background(), export)

    return export, nil
}

func (e *Exporter) executeExport(ctx context.Context, export *AuditExport) {
    e.repo.UpdateExportStatus(ctx, export.ID, "processing")

    // Query data (streaming for large datasets)
    entries, total, err := e.repo.QuerySyncLogsStream(ctx, export.Filters)
    if err != nil {
        e.repo.UpdateExportError(ctx, export.ID, err.Error())
        return
    }

    // Generate file
    var filePath string
    var fileSize int64

    switch export.Format {
    case "csv":
        filePath, fileSize, err = e.generateCSV(ctx, entries)
    case "pdf":
        filePath, fileSize, err = e.generatePDF(ctx, entries)
    }

    if err != nil {
        e.repo.UpdateExportError(ctx, export.ID, err.Error())
        return
    }

    // Update export record
    e.repo.UpdateExportComplete(ctx, export.ID, filePath, fileSize, total)
}

func (e *Exporter) generateCSV(ctx context.Context, entries <-chan SyncLogEntry) (string, int64, error) {
    filePath := fmt.Sprintf("/tmp/exports/audit_%s.csv", uuid.New().String())
    file, _ := os.Create(filePath)
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Header
    writer.Write([]string{
        "ID", "System", "Operation", "Status", "Records Total",
        "Records Succeeded", "Records Failed", "Started At",
        "Completed At", "Duration (ms)", "User", "Error",
    })

    // Data rows
    for entry := range entries {
        writer.Write([]string{
            entry.ID.String(),
            entry.SystemName,
            entry.Operation,
            entry.Status,
            strconv.Itoa(entry.RecordsTotal),
            strconv.Itoa(entry.RecordsSucceeded),
            strconv.Itoa(entry.RecordsFailed),
            entry.StartedAt.Format(time.RFC3339),
            formatNullableTime(entry.CompletedAt),
            formatNullableInt(entry.DurationMs),
            entry.UserEmail,
            formatNullableString(entry.ErrorMessage),
        })
    }

    info, _ := file.Stat()
    return filePath, info.Size(), nil
}
```

---

## 7. State Management

### Frontend State

```typescript
// src/features/audit/hooks/useAuditLog.ts

interface AuditFilters {
  systemId?: string;
  userId?: string;
  operation?: 'pull' | 'push';
  status?: 'completed' | 'failed' | 'partial';
  startDate?: Date;
  endDate?: Date;
  search?: string;
}

export function useAuditLog(filters: AuditFilters) {
  return useQuery({
    queryKey: ['audit', 'sync', filters],
    queryFn: () => auditApi.getSyncLogs(filters),
    keepPreviousData: true,
  });
}

export function useSyncLogDetail(id: string) {
  return useQuery({
    queryKey: ['audit', 'sync', id],
    queryFn: () => auditApi.getSyncLogDetail(id),
    enabled: !!id,
  });
}

export function useExport() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: ExportParams) => auditApi.startExport(params),
    onSuccess: (data) => {
      // Poll for completion
      queryClient.invalidateQueries(['audit', 'export', data.export_id]);
    },
  });
}
```

---

## 8. Security Considerations

### Immutability
- Audit tables are append-only (no UPDATE/DELETE in application)
- Database user has INSERT only, no UPDATE/DELETE
- Consider row-level security for production

### Access Control
- All users can view audit logs (read-only)
- Export requires `reviewer` or higher role
- Admin can see all users' activities

### Data Privacy
- User emails stored (for attribution)
- Statement content optionally stored (configurable)
- Consider anonymization for old records

---

## 9. Performance & Scalability

### Performance Targets

| Operation | Target | Approach |
|-----------|--------|----------|
| Query 1000 logs | < 2s | Indexed queries, pagination |
| Get log details | < 500ms | Primary key lookup |
| Export 10K records | < 30s | Streaming, background job |

### Optimization Strategies

1. **Efficient Indexing**
   - Composite indexes for common filter combinations
   - Partial indexes for active statuses

2. **Streaming for Exports**
   ```go
   // Use channels for memory-efficient export
   func (r *repository) QuerySyncLogsStream(ctx context.Context, q Query) (<-chan SyncLogEntry, error) {
       ch := make(chan SyncLogEntry, 100)
       go func() {
           defer close(ch)
           rows, _ := r.db.Query(ctx, query, args...)
           for rows.Next() {
               var entry SyncLogEntry
               rows.Scan(&entry)
               ch <- entry
           }
       }()
       return ch, nil
   }
   ```

3. **Query Result Caching**
   - Short TTL (30s) for aggregate counts
   - TanStack Query cache for frontend

---

## 10. Testing Strategy

### Unit Tests

```go
// internal/domain/audit/service_test.go

func TestAuditLogger_LogSyncStart(t *testing.T) {
    // Test entry creation
    // Verify all fields populated
}

func TestAuditLogger_LogSyncComplete(t *testing.T) {
    // Test status update
    // Verify duration calculation
}

func TestExporter_GenerateCSV(t *testing.T) {
    // Test CSV formatting
    // Verify all columns present
}
```

### Integration Tests

```go
func TestAuditRepository_QuerySyncLogs(t *testing.T) {
    // Test filter combinations
    // Test pagination
    // Test sorting
}
```

### E2E Tests

```typescript
test('user can view audit log', async ({ page }) => {
  await page.goto('/audit');
  await expect(page.locator('table')).toBeVisible();
  await expect(page.locator('tbody tr')).toHaveCount(20);
});

test('user can filter audit log', async ({ page }) => {
  await page.goto('/audit');
  await page.selectOption('#operation-filter', 'push');
  await expect(page.locator('tbody tr')).toContainText('push');
});

test('user can export audit log', async ({ page }) => {
  await page.goto('/audit');
  await page.click('button:has-text("Export")');
  await page.selectOption('#format', 'csv');
  await page.click('button:has-text("Generate Export")');

  // Wait for download
  const download = await page.waitForEvent('download');
  expect(download.suggestedFilename()).toMatch(/audit.*\.csv/);
});
```

---

## 11. Deployment & DevOps

### Configuration

```bash
# Audit settings
AUDIT_RETENTION_DAYS=2555  # 7 years
AUDIT_EXPORT_MAX_RECORDS=10000
AUDIT_EXPORT_EXPIRY_HOURS=24
AUDIT_EXPORT_PATH=/var/exports

# Content logging (optional)
AUDIT_LOG_CONTENT=false  # Set true to log statement content changes
```

### Monitoring

- Track query performance (slow queries > 2s)
- Monitor export queue depth
- Alert on high failure rates in sync logs

---

## 12. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Large table performance | Medium | Medium | Indexing, partitioning |
| Export memory issues | Medium | Low | Streaming, limits |
| Storage growth | High | Low | Retention policy, archival |
| Data privacy concerns | Low | Medium | Configurable content logging |

---

## 13. Development Phases

### Phase 1: Core Logging (2 days)
- Database schema
- Audit logger implementation
- Integration with Pull/Push services

### Phase 2: Query API (2 days)
- Query endpoint with filters
- Detail endpoint
- Statement history endpoint

### Phase 3: Frontend (2 days)
- Audit log table
- Filter panel
- Expandable details

### Phase 4: Export (2 days)
- Export generation (CSV, PDF)
- Background job processing
- Download handling

### Phase 5: Testing (1 day)
- Integration tests
- E2E tests

**Estimated Total: 9 days**

---

## 14. Open Technical Questions

1. **Content Storage:** Should we store full statement content for all changes or just hashes?
2. **Archival:** When should we archive old logs to cold storage?
3. **Real-time:** Should we support real-time audit log streaming (WebSocket)?
4. **SIEM Integration:** What format for future SIEM export (CEF, LEEF, JSON)?
