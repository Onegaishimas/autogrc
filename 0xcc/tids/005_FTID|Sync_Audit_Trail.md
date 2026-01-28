# Technical Implementation Document: Sync Audit Trail (F5)

**Document Version:** 1.0
**Date:** 2026-01-27
**Feature ID:** F5
**Related PRD:** 005_FPRD|Sync_Audit_Trail.md
**Related TDD:** 005_FTDD|Sync_Audit_Trail.md

---

## 1. Implementation Overview

### Summary
This TID provides implementation guidance for the comprehensive audit trail system that logs all synchronization operations, supports querying and filtering, and enables export for compliance reporting.

### Key Implementation Principles
- **Append-Only Storage:** Audit records are immutable once written
- **Structured Events:** All events follow consistent schema with typed fields
- **Efficient Querying:** Indexed columns for common filter patterns
- **7-Year Retention:** Partitioned tables for long-term storage management

### Integration Points
- **F1-F4 Dependencies:** Receives audit events from all sync operations
- **Database:** PostgreSQL with table partitioning for retention
- **Export:** CSV and PDF generation for compliance reports

---

## 2. File Structure and Organization

### Backend Files to Create

```
backend/
├── internal/
│   ├── api/
│   │   └── handlers/
│   │       └── audit/
│   │           ├── handler.go            # HTTP handlers (create)
│   │           ├── handler_test.go       # Handler tests (create)
│   │           └── schemas.go            # Request/response DTOs (create)
│   ├── domain/
│   │   └── audit/
│   │       ├── models.go                 # Audit domain models (create)
│   │       ├── service.go                # Audit business logic (create)
│   │       ├── service_test.go           # Service tests (create)
│   │       ├── repository.go             # DB interface (create)
│   │       ├── query_builder.go          # Dynamic query builder (create)
│   │       └── exporter.go               # Export functionality (create)
│   └── infrastructure/
│       ├── database/
│       │   └── queries/
│       │       └── audit.sql             # sqlc queries (create)
│       └── export/
│           ├── csv.go                    # CSV export (create)
│           └── pdf.go                    # PDF export (create)
└── migrations/
    └── 20260127_005_create_audit_tables.sql  # Audit schema (create)
```

### Frontend Files to Create

```
frontend/src/
├── features/
│   └── audit/
│       ├── components/
│       │   ├── AuditLog.tsx              # Main audit log view (create)
│       │   ├── AuditFilters.tsx          # Filter controls (create)
│       │   ├── AuditTable.tsx            # Events table (create)
│       │   ├── AuditDetail.tsx           # Event detail modal (create)
│       │   ├── ExportDialog.tsx          # Export options dialog (create)
│       │   └── DateRangePicker.tsx       # Date range selector (create)
│       ├── hooks/
│       │   ├── useAuditEvents.ts         # Query audit events (create)
│       │   └── useExport.ts              # Export mutation hook (create)
│       ├── api/
│       │   └── auditApi.ts               # API client (create)
│       └── types.ts                      # TypeScript types (create)
└── pages/
    └── AuditPage.tsx                     # Audit log page (create)
```

---

## 3. Component Implementation Hints

### Audit Event Model

```go
// internal/domain/audit/models.go
type EventType string

const (
    EventTypePull              EventType = "pull"
    EventTypePush              EventType = "push"
    EventTypeEdit              EventType = "edit"
    EventTypeConflictDetected  EventType = "conflict_detected"
    EventTypeConflictResolved  EventType = "conflict_resolved"
    EventTypeConnectionTest    EventType = "connection_test"
    EventTypeConnectionConfig  EventType = "connection_config"
)

type Event struct {
    ID          uuid.UUID              `db:"id"`
    EventType   EventType              `db:"event_type"`
    EntityType  string                 `db:"entity_type"`  // system, control, statement
    EntityID    string                 `db:"entity_id"`
    UserID      *uuid.UUID             `db:"user_id"`
    UserEmail   *string                `db:"user_email"`   // Denormalized for search
    Action      string                 `db:"action"`
    Status      string                 `db:"status"`       // success, failure
    Details     map[string]interface{} `db:"details"`      // JSONB
    IPAddress   *string                `db:"ip_address"`
    UserAgent   *string                `db:"user_agent"`
    CreatedAt   time.Time              `db:"created_at"`
}

type QueryFilters struct {
    EventTypes  []EventType
    EntityTypes []string
    EntityID    *string
    UserID      *uuid.UUID
    Status      *string
    StartDate   *time.Time
    EndDate     *time.Time
    SearchText  *string
    Page        int
    PerPage     int
}
```

### Audit Service Interface

```go
// internal/domain/audit/service.go
type Service struct {
    repo     Repository
    exporter Exporter
    logger   zerolog.Logger
}

// Record is called by other services to log events
func (s *Service) Record(ctx context.Context, event Event) error {
    // Enrich event with request context
    if reqCtx, ok := ctx.Value(RequestContextKey).(*RequestContext); ok {
        event.IPAddress = &reqCtx.IPAddress
        event.UserAgent = &reqCtx.UserAgent
        if reqCtx.User != nil {
            event.UserID = &reqCtx.User.ID
            event.UserEmail = &reqCtx.User.Email
        }
    }

    event.ID = uuid.New()
    event.CreatedAt = time.Now()

    if err := s.repo.Insert(ctx, &event); err != nil {
        s.logger.Error().Err(err).Str("event_type", string(event.EventType)).Msg("failed to record audit event")
        return fmt.Errorf("failed to record audit event: %w", err)
    }

    return nil
}

// Query returns filtered and paginated events
func (s *Service) Query(ctx context.Context, filters QueryFilters) (*QueryResult, error) {
    return s.repo.Query(ctx, filters)
}

// Export generates a file with filtered events
func (s *Service) Export(ctx context.Context, filters QueryFilters, format string) ([]byte, error) {
    // Fetch all matching events (no pagination for export)
    filters.Page = 0
    filters.PerPage = 0 // 0 means no limit

    events, err := s.repo.QueryAll(ctx, filters)
    if err != nil {
        return nil, fmt.Errorf("failed to query events: %w", err)
    }

    switch format {
    case "csv":
        return s.exporter.ToCSV(events)
    case "pdf":
        return s.exporter.ToPDF(events)
    default:
        return nil, fmt.Errorf("unsupported export format: %s", format)
    }
}
```

---

## 4. Database Implementation Approach

### Partitioned Audit Table Schema

```sql
-- migrations/20260127_005_create_audit_tables.sql

-- Enable partitioning extension if needed
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Audit events table with monthly partitions
CREATE TABLE audit_events (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    user_id UUID,
    user_email VARCHAR(255),
    action VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'success',
    details JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Create partitions for current and next year
-- (Production: automate partition creation via pg_partman or cron)
CREATE TABLE audit_events_2026_01 PARTITION OF audit_events
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE audit_events_2026_02 PARTITION OF audit_events
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
-- ... continue for remaining months

-- Default partition for future dates
CREATE TABLE audit_events_default PARTITION OF audit_events DEFAULT;

-- Indexes for common query patterns
CREATE INDEX idx_audit_event_type ON audit_events(event_type);
CREATE INDEX idx_audit_entity ON audit_events(entity_type, entity_id);
CREATE INDEX idx_audit_user ON audit_events(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_audit_created ON audit_events(created_at DESC);
CREATE INDEX idx_audit_status ON audit_events(status);

-- Full-text search on details
CREATE INDEX idx_audit_details_gin ON audit_events USING GIN (details);

-- Composite index for common filter combination
CREATE INDEX idx_audit_common_filters ON audit_events(event_type, entity_type, created_at DESC);
```

### sqlc Query Patterns

```sql
-- name: InsertAuditEvent :exec
INSERT INTO audit_events (
    id, event_type, entity_type, entity_id, user_id, user_email,
    action, status, details, ip_address, user_agent, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: GetAuditEventByID :one
SELECT * FROM audit_events WHERE id = $1 AND created_at = $2;

-- Note: Complex queries with dynamic filters use query builder (see below)
```

### Dynamic Query Builder

```go
// internal/domain/audit/query_builder.go
import (
    sq "github.com/Masterminds/squirrel"
)

func (r *Repository) buildQuery(filters QueryFilters) (string, []interface{}, error) {
    builder := sq.Select(
        "id", "event_type", "entity_type", "entity_id",
        "user_id", "user_email", "action", "status",
        "details", "ip_address", "user_agent", "created_at",
    ).
        From("audit_events").
        OrderBy("created_at DESC").
        PlaceholderFormat(sq.Dollar)

    // Apply filters
    if len(filters.EventTypes) > 0 {
        builder = builder.Where(sq.Eq{"event_type": filters.EventTypes})
    }
    if len(filters.EntityTypes) > 0 {
        builder = builder.Where(sq.Eq{"entity_type": filters.EntityTypes})
    }
    if filters.EntityID != nil {
        builder = builder.Where(sq.Eq{"entity_id": *filters.EntityID})
    }
    if filters.UserID != nil {
        builder = builder.Where(sq.Eq{"user_id": *filters.UserID})
    }
    if filters.Status != nil {
        builder = builder.Where(sq.Eq{"status": *filters.Status})
    }
    if filters.StartDate != nil {
        builder = builder.Where(sq.GtOrEq{"created_at": *filters.StartDate})
    }
    if filters.EndDate != nil {
        builder = builder.Where(sq.LtOrEq{"created_at": *filters.EndDate})
    }
    if filters.SearchText != nil && *filters.SearchText != "" {
        // Search in user_email and details JSONB
        builder = builder.Where(sq.Or{
            sq.ILike{"user_email": "%" + *filters.SearchText + "%"},
            sq.Expr("details::text ILIKE ?", "%"+*filters.SearchText+"%"),
        })
    }

    // Pagination
    if filters.PerPage > 0 {
        builder = builder.Limit(uint64(filters.PerPage))
        if filters.Page > 0 {
            builder = builder.Offset(uint64(filters.Page * filters.PerPage))
        }
    }

    return builder.ToSql()
}

func (r *Repository) Query(ctx context.Context, filters QueryFilters) (*QueryResult, error) {
    sql, args, err := r.buildQuery(filters)
    if err != nil {
        return nil, fmt.Errorf("failed to build query: %w", err)
    }

    rows, err := r.db.Query(ctx, sql, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to execute query: %w", err)
    }
    defer rows.Close()

    var events []*Event
    for rows.Next() {
        var e Event
        if err := rows.Scan(/* fields */); err != nil {
            return nil, fmt.Errorf("failed to scan row: %w", err)
        }
        events = append(events, &e)
    }

    // Get total count for pagination
    countSQL, countArgs, _ := r.buildCountQuery(filters)
    var totalCount int
    r.db.QueryRow(ctx, countSQL, countArgs...).Scan(&totalCount)

    return &QueryResult{
        Events:     events,
        TotalCount: totalCount,
        Page:       filters.Page,
        PerPage:    filters.PerPage,
    }, nil
}
```

---

## 5. API Implementation Strategy

### Endpoints

```go
// internal/api/handlers/audit/handler.go
func (h *Handler) RegisterRoutes(r chi.Router) {
    r.Route("/api/v1/audit", func(r chi.Router) {
        r.Get("/", h.QueryEvents)
        r.Get("/{id}", h.GetEvent)
        r.Get("/export", h.ExportEvents)
        r.Get("/stats", h.GetStats)
    })
}
```

### Request/Response Schemas

```go
// internal/api/handlers/audit/schemas.go

type QueryEventsRequest struct {
    EventTypes  []string `query:"event_type"`
    EntityTypes []string `query:"entity_type"`
    EntityID    string   `query:"entity_id"`
    UserID      string   `query:"user_id"`
    Status      string   `query:"status"`
    StartDate   string   `query:"start_date"`
    EndDate     string   `query:"end_date"`
    Search      string   `query:"search"`
    Page        int      `query:"page"`
    PerPage     int      `query:"per_page"`
}

type QueryEventsResponse struct {
    Events     []EventResponse `json:"events"`
    Pagination PaginationMeta  `json:"pagination"`
}

type EventResponse struct {
    ID         uuid.UUID              `json:"id"`
    EventType  string                 `json:"event_type"`
    EntityType string                 `json:"entity_type"`
    EntityID   string                 `json:"entity_id"`
    UserEmail  *string                `json:"user_email,omitempty"`
    Action     string                 `json:"action"`
    Status     string                 `json:"status"`
    Details    map[string]interface{} `json:"details"`
    CreatedAt  time.Time              `json:"created_at"`
}

type PaginationMeta struct {
    Page       int `json:"page"`
    PerPage    int `json:"per_page"`
    TotalCount int `json:"total_count"`
    TotalPages int `json:"total_pages"`
}

type ExportEventsRequest struct {
    Format string `query:"format" validate:"required,oneof=csv pdf"`
    // Same filters as QueryEventsRequest
}

type StatsResponse struct {
    TotalEvents     int            `json:"total_events"`
    EventsByType    map[string]int `json:"events_by_type"`
    EventsByStatus  map[string]int `json:"events_by_status"`
    EventsToday     int            `json:"events_today"`
    EventsThisWeek  int            `json:"events_this_week"`
    EventsThisMonth int            `json:"events_this_month"`
}
```

---

## 6. Frontend Implementation Approach

### Audit Log Main Component

```typescript
// features/audit/components/AuditLog.tsx
export function AuditLog() {
  const [filters, setFilters] = useState<AuditFilters>({
    page: 0,
    perPage: 50,
  });
  const [selectedEvent, setSelectedEvent] = useState<string | null>(null);

  const { data, isLoading } = useAuditEvents(filters);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Audit Trail</h1>
        <ExportButton filters={filters} />
      </div>

      <AuditFilters
        filters={filters}
        onChange={setFilters}
      />

      <AuditTable
        events={data?.events ?? []}
        isLoading={isLoading}
        onSelectEvent={setSelectedEvent}
      />

      <Pagination
        page={filters.page}
        perPage={filters.perPage}
        totalCount={data?.pagination.totalCount ?? 0}
        onPageChange={(page) => setFilters({ ...filters, page })}
      />

      {selectedEvent && (
        <AuditDetail
          eventId={selectedEvent}
          onClose={() => setSelectedEvent(null)}
        />
      )}
    </div>
  );
}
```

### Filter Component

```typescript
// features/audit/components/AuditFilters.tsx
const EVENT_TYPES = [
  { value: 'pull', label: 'Pull' },
  { value: 'push', label: 'Push' },
  { value: 'edit', label: 'Edit' },
  { value: 'conflict_detected', label: 'Conflict Detected' },
  { value: 'conflict_resolved', label: 'Conflict Resolved' },
  { value: 'connection_test', label: 'Connection Test' },
  { value: 'connection_config', label: 'Connection Config' },
];

const ENTITY_TYPES = [
  { value: 'system', label: 'System' },
  { value: 'control', label: 'Control' },
  { value: 'statement', label: 'Statement' },
  { value: 'connection', label: 'Connection' },
];

interface AuditFiltersProps {
  filters: AuditFilters;
  onChange: (filters: AuditFilters) => void;
}

export function AuditFilters({ filters, onChange }: AuditFiltersProps) {
  return (
    <Card>
      <CardContent className="pt-4">
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {/* Event Type Multi-Select */}
          <div className="space-y-2">
            <Label>Event Type</Label>
            <MultiSelect
              options={EVENT_TYPES}
              selected={filters.eventTypes ?? []}
              onChange={(values) => onChange({ ...filters, eventTypes: values, page: 0 })}
              placeholder="All types"
            />
          </div>

          {/* Entity Type Multi-Select */}
          <div className="space-y-2">
            <Label>Entity Type</Label>
            <MultiSelect
              options={ENTITY_TYPES}
              selected={filters.entityTypes ?? []}
              onChange={(values) => onChange({ ...filters, entityTypes: values, page: 0 })}
              placeholder="All entities"
            />
          </div>

          {/* Date Range */}
          <div className="space-y-2">
            <Label>Date Range</Label>
            <DateRangePicker
              from={filters.startDate}
              to={filters.endDate}
              onChange={(range) => onChange({
                ...filters,
                startDate: range.from,
                endDate: range.to,
                page: 0,
              })}
            />
          </div>

          {/* Search */}
          <div className="space-y-2">
            <Label>Search</Label>
            <div className="relative">
              <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search events..."
                value={filters.search ?? ''}
                onChange={(e) => onChange({ ...filters, search: e.target.value, page: 0 })}
                className="pl-8"
              />
            </div>
          </div>

          {/* Status */}
          <div className="space-y-2">
            <Label>Status</Label>
            <Select
              value={filters.status ?? 'all'}
              onValueChange={(v) => onChange({
                ...filters,
                status: v === 'all' ? undefined : v,
                page: 0,
              })}
            >
              <SelectTrigger>
                <SelectValue placeholder="All statuses" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All</SelectItem>
                <SelectItem value="success">Success</SelectItem>
                <SelectItem value="failure">Failure</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        {/* Active filters summary */}
        <ActiveFilters filters={filters} onClear={() => onChange({ page: 0, perPage: 50 })} />
      </CardContent>
    </Card>
  );
}
```

### Audit Table Component

```typescript
// features/audit/components/AuditTable.tsx
const columns: ColumnDef<AuditEvent>[] = [
  {
    accessorKey: 'createdAt',
    header: 'Time',
    cell: ({ row }) => format(new Date(row.original.createdAt), 'PPpp'),
  },
  {
    accessorKey: 'eventType',
    header: 'Event',
    cell: ({ row }) => (
      <Badge variant={getEventVariant(row.original.eventType)}>
        {formatEventType(row.original.eventType)}
      </Badge>
    ),
  },
  {
    accessorKey: 'entityType',
    header: 'Entity',
    cell: ({ row }) => (
      <span className="font-medium">
        {row.original.entityType}: {truncate(row.original.entityId, 20)}
      </span>
    ),
  },
  {
    accessorKey: 'action',
    header: 'Action',
  },
  {
    accessorKey: 'userEmail',
    header: 'User',
    cell: ({ row }) => row.original.userEmail ?? 'System',
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({ row }) => (
      <Badge variant={row.original.status === 'success' ? 'default' : 'destructive'}>
        {row.original.status}
      </Badge>
    ),
  },
  {
    id: 'actions',
    cell: ({ row }) => (
      <Button variant="ghost" size="sm" onClick={() => onSelect(row.original.id)}>
        <Eye className="h-4 w-4" />
      </Button>
    ),
  },
];

export function AuditTable({ events, isLoading, onSelectEvent }: AuditTableProps) {
  return (
    <DataTable
      columns={columns}
      data={events}
      isLoading={isLoading}
    />
  );
}
```

### Event Detail Modal

```typescript
// features/audit/components/AuditDetail.tsx
export function AuditDetail({ eventId, onClose }: AuditDetailProps) {
  const { data: event, isLoading } = useAuditEvent(eventId);

  if (isLoading) return <Skeleton className="h-96" />;
  if (!event) return null;

  return (
    <Dialog open onOpenChange={onClose}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Audit Event Details</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <DetailItem label="Event ID" value={event.id} />
            <DetailItem label="Event Type" value={formatEventType(event.eventType)} />
            <DetailItem label="Entity" value={`${event.entityType}: ${event.entityId}`} />
            <DetailItem label="Action" value={event.action} />
            <DetailItem label="Status" value={event.status} />
            <DetailItem label="User" value={event.userEmail ?? 'System'} />
            <DetailItem label="Time" value={format(new Date(event.createdAt), 'PPpp')} />
            <DetailItem label="IP Address" value={event.ipAddress ?? 'N/A'} />
          </div>

          {Object.keys(event.details).length > 0 && (
            <div className="space-y-2">
              <Label>Details</Label>
              <pre className="p-4 bg-muted rounded-md text-sm overflow-auto max-h-64">
                {JSON.stringify(event.details, null, 2)}
              </pre>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
```

---

## 7. Export Implementation

### CSV Exporter

```go
// internal/infrastructure/export/csv.go
import (
    "encoding/csv"
    "bytes"
)

type CSVExporter struct{}

func (e *CSVExporter) Export(events []*audit.Event) ([]byte, error) {
    var buf bytes.Buffer
    writer := csv.NewWriter(&buf)

    // Header
    header := []string{
        "Event ID", "Timestamp", "Event Type", "Entity Type", "Entity ID",
        "Action", "Status", "User Email", "IP Address", "Details",
    }
    if err := writer.Write(header); err != nil {
        return nil, fmt.Errorf("failed to write header: %w", err)
    }

    // Data rows
    for _, event := range events {
        row := []string{
            event.ID.String(),
            event.CreatedAt.Format(time.RFC3339),
            string(event.EventType),
            event.EntityType,
            event.EntityID,
            event.Action,
            event.Status,
            safeString(event.UserEmail),
            safeString(event.IPAddress),
            marshalDetails(event.Details),
        }
        if err := writer.Write(row); err != nil {
            return nil, fmt.Errorf("failed to write row: %w", err)
        }
    }

    writer.Flush()
    if err := writer.Error(); err != nil {
        return nil, fmt.Errorf("csv write error: %w", err)
    }

    return buf.Bytes(), nil
}

func safeString(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}

func marshalDetails(d map[string]interface{}) string {
    if len(d) == 0 {
        return ""
    }
    b, _ := json.Marshal(d)
    return string(b)
}
```

### PDF Exporter

```go
// internal/infrastructure/export/pdf.go
import (
    "github.com/jung-kurt/gofpdf"
)

type PDFExporter struct{}

func (e *PDFExporter) Export(events []*audit.Event) ([]byte, error) {
    pdf := gofpdf.New("L", "mm", "A4", "") // Landscape for more columns
    pdf.SetFont("Arial", "", 8)

    pdf.AddPage()

    // Title
    pdf.SetFont("Arial", "B", 14)
    pdf.Cell(0, 10, "Audit Trail Report")
    pdf.Ln(12)

    // Report metadata
    pdf.SetFont("Arial", "", 10)
    pdf.Cell(0, 6, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")))
    pdf.Ln(6)
    pdf.Cell(0, 6, fmt.Sprintf("Total Events: %d", len(events)))
    pdf.Ln(12)

    // Table header
    pdf.SetFont("Arial", "B", 8)
    headers := []string{"Timestamp", "Type", "Entity", "Action", "Status", "User"}
    widths := []float64{40, 30, 50, 60, 20, 50}
    for i, h := range headers {
        pdf.CellFormat(widths[i], 7, h, "1", 0, "C", false, 0, "")
    }
    pdf.Ln(-1)

    // Table data
    pdf.SetFont("Arial", "", 7)
    for _, event := range events {
        row := []string{
            event.CreatedAt.Format("2006-01-02 15:04"),
            string(event.EventType),
            fmt.Sprintf("%s: %s", event.EntityType, truncate(event.EntityID, 20)),
            truncate(event.Action, 40),
            event.Status,
            safeString(event.UserEmail),
        }
        for i, cell := range row {
            pdf.CellFormat(widths[i], 6, cell, "1", 0, "L", false, 0, "")
        }
        pdf.Ln(-1)

        // Page break if needed
        if pdf.GetY() > 180 {
            pdf.AddPage()
            // Re-add header on new page
            pdf.SetFont("Arial", "B", 8)
            for i, h := range headers {
                pdf.CellFormat(widths[i], 7, h, "1", 0, "C", false, 0, "")
            }
            pdf.Ln(-1)
            pdf.SetFont("Arial", "", 7)
        }
    }

    var buf bytes.Buffer
    if err := pdf.Output(&buf); err != nil {
        return nil, fmt.Errorf("failed to generate PDF: %w", err)
    }

    return buf.Bytes(), nil
}
```

---

## 8. Testing Implementation Approach

### Repository Tests

```go
// internal/domain/audit/repository_test.go
func TestRepository_Query(t *testing.T) {
    ctx := context.Background()
    repo := setupTestRepo(t)

    // Seed test data
    events := []*Event{
        {EventType: EventTypePull, EntityType: "system", Status: "success"},
        {EventType: EventTypePush, EntityType: "statement", Status: "success"},
        {EventType: EventTypePush, EntityType: "statement", Status: "failure"},
    }
    for _, e := range events {
        repo.Insert(ctx, e)
    }

    tests := []struct {
        name    string
        filters QueryFilters
        wantLen int
    }{
        {
            name:    "filter by event type",
            filters: QueryFilters{EventTypes: []EventType{EventTypePush}},
            wantLen: 2,
        },
        {
            name:    "filter by status",
            filters: QueryFilters{Status: ptr("failure")},
            wantLen: 1,
        },
        {
            name:    "combined filters",
            filters: QueryFilters{
                EventTypes: []EventType{EventTypePush},
                Status:     ptr("success"),
            },
            wantLen: 1,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := repo.Query(ctx, tt.filters)
            assert.NoError(t, err)
            assert.Len(t, result.Events, tt.wantLen)
        })
    }
}
```

### Frontend Tests

```typescript
// features/audit/components/AuditFilters.test.tsx
describe('AuditFilters', () => {
  it('calls onChange when event type filter changes', async () => {
    const onChange = jest.fn();
    render(<AuditFilters filters={{}} onChange={onChange} />);

    await userEvent.click(screen.getByLabelText('Event Type'));
    await userEvent.click(screen.getByText('Pull'));

    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({ eventTypes: ['pull'], page: 0 })
    );
  });

  it('resets page to 0 when any filter changes', async () => {
    const onChange = jest.fn();
    render(<AuditFilters filters={{ page: 5 }} onChange={onChange} />);

    await userEvent.type(screen.getByPlaceholderText('Search events...'), 'test');

    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({ page: 0 })
    );
  });
});
```

---

## 9. Integration with Other Features

### Recording Events from F1-F4

```go
// Example: Recording pull event in pull executor
func (e *PullExecutor) pullSystem(ctx context.Context, job *PullJob, systemID uuid.UUID) error {
    startTime := time.Now()

    // ... pull logic ...

    // Record success
    e.auditService.Record(ctx, audit.Event{
        EventType:  audit.EventTypePull,
        EntityType: "system",
        EntityID:   systemID.String(),
        Action:     "pull_controls",
        Status:     "success",
        Details: map[string]interface{}{
            "controls_count":   controlsCount,
            "statements_count": statementsCount,
            "duration_ms":      time.Since(startTime).Milliseconds(),
        },
    })

    return nil
}

// Example: Recording edit event in statement service
func (s *Service) SaveContent(ctx context.Context, req SaveContentRequest) (*Statement, error) {
    // ... save logic ...

    s.auditService.Record(ctx, audit.Event{
        EventType:  audit.EventTypeEdit,
        EntityType: "statement",
        EntityID:   req.ID.String(),
        Action:     "update_content",
        Status:     "success",
        Details: map[string]interface{}{
            "content_length": len(req.Content),
            "session_id":     req.SessionID,
        },
    })

    return stmt, nil
}
```

---

## 10. Performance Implementation Hints

### Query Performance

```sql
-- Ensure partition pruning by always including date range
-- BAD: SELECT * FROM audit_events WHERE event_type = 'pull'
-- GOOD: SELECT * FROM audit_events WHERE event_type = 'pull' AND created_at > NOW() - INTERVAL '30 days'

-- Use EXPLAIN ANALYZE to verify partition pruning
EXPLAIN ANALYZE
SELECT * FROM audit_events
WHERE event_type = 'pull'
  AND created_at BETWEEN '2026-01-01' AND '2026-01-31';
```

### Async Event Recording

```go
// Use channel-based async recording to prevent blocking main operations
type AsyncAuditService struct {
    events chan Event
    repo   Repository
}

func NewAsyncAuditService(repo Repository, bufferSize int) *AsyncAuditService {
    s := &AsyncAuditService{
        events: make(chan Event, bufferSize),
        repo:   repo,
    }
    go s.processEvents()
    return s
}

func (s *AsyncAuditService) Record(ctx context.Context, event Event) error {
    select {
    case s.events <- event:
        return nil
    default:
        // Buffer full - log warning but don't fail the main operation
        log.Warn().Msg("audit event buffer full, event dropped")
        return nil
    }
}

func (s *AsyncAuditService) processEvents() {
    for event := range s.events {
        if err := s.repo.Insert(context.Background(), &event); err != nil {
            log.Error().Err(err).Msg("failed to persist audit event")
        }
    }
}
```

---

## 11. Retention and Archival

### Partition Management

```sql
-- Create function to manage partitions
CREATE OR REPLACE FUNCTION manage_audit_partitions() RETURNS void AS $$
DECLARE
    partition_date DATE;
    partition_name TEXT;
BEGIN
    -- Create next month's partition if it doesn't exist
    partition_date := date_trunc('month', NOW() + INTERVAL '1 month');
    partition_name := 'audit_events_' || to_char(partition_date, 'YYYY_MM');

    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF audit_events FOR VALUES FROM (%L) TO (%L)',
        partition_name,
        partition_date,
        partition_date + INTERVAL '1 month'
    );

    -- Drop partitions older than 7 years
    FOR partition_name IN
        SELECT tablename FROM pg_tables
        WHERE tablename LIKE 'audit_events_%'
          AND tablename < 'audit_events_' || to_char(NOW() - INTERVAL '7 years', 'YYYY_MM')
    LOOP
        EXECUTE format('DROP TABLE IF EXISTS %I', partition_name);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Schedule via pg_cron or external scheduler
-- SELECT cron.schedule('manage_audit_partitions', '0 0 1 * *', 'SELECT manage_audit_partitions()');
```

---

## 12. Security Checklist

- [ ] Audit events are append-only (no UPDATE or DELETE allowed)
- [ ] Sensitive data (passwords, tokens) never stored in details
- [ ] User cannot modify audit records via API
- [ ] Export requires appropriate permissions
- [ ] Rate limit export endpoint to prevent DoS
- [ ] IP address captured for all events
- [ ] Failed operations are also audited
- [ ] Partition management automated and monitored
