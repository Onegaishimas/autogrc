# Task List: Sync Audit Trail (F5)

**Feature ID:** F5
**Related PRD:** 005_FPRD|Sync_Audit_Trail.md
**Related TDD:** 005_FTDD|Sync_Audit_Trail.md
**Related TID:** 005_FTID|Sync_Audit_Trail.md
**Depends On:** F1, F2, F3, F4 (Receives events from all sync operations)

---

## Relevant Files

### Backend - Database
- `backend/migrations/20260127_005_create_audit_tables.sql` - Partitioned audit tables
- `backend/internal/infrastructure/database/queries/audit.sql` - Audit sqlc queries

### Backend - Domain
- `backend/internal/domain/audit/models.go` - Audit event domain models
- `backend/internal/domain/audit/service.go` - Audit business logic
- `backend/internal/domain/audit/service_test.go` - Audit service tests
- `backend/internal/domain/audit/repository.go` - Audit repository interface
- `backend/internal/domain/audit/query_builder.go` - Dynamic query builder
- `backend/internal/domain/audit/exporter.go` - Export functionality

### Backend - Infrastructure (Export)
- `backend/internal/infrastructure/export/csv.go` - CSV export implementation
- `backend/internal/infrastructure/export/pdf.go` - PDF export implementation

### Backend - API
- `backend/internal/api/handlers/audit/handler.go` - Audit HTTP handlers
- `backend/internal/api/handlers/audit/handler_test.go` - Handler tests
- `backend/internal/api/handlers/audit/schemas.go` - Audit DTOs

### Frontend - Audit Feature
- `frontend/src/features/audit/components/AuditLog.tsx` - Main audit log view
- `frontend/src/features/audit/components/AuditFilters.tsx` - Filter controls
- `frontend/src/features/audit/components/AuditTable.tsx` - Events table
- `frontend/src/features/audit/components/AuditDetail.tsx` - Event detail modal
- `frontend/src/features/audit/components/ExportDialog.tsx` - Export options
- `frontend/src/features/audit/components/DateRangePicker.tsx` - Date range selector
- `frontend/src/features/audit/hooks/useAuditEvents.ts` - Query audit events
- `frontend/src/features/audit/hooks/useExport.ts` - Export mutation
- `frontend/src/features/audit/api/auditApi.ts` - API client
- `frontend/src/features/audit/types.ts` - TypeScript types

### Frontend - Pages
- `frontend/src/pages/AuditPage.tsx` - Audit log page

### Frontend - Tests
- `frontend/src/features/audit/components/AuditFilters.test.tsx` - Filter tests
- `frontend/src/features/audit/components/AuditTable.test.tsx` - Table tests

### Notes

- Audit tables use PostgreSQL partitioning for 7-year retention
- Events are append-only (no UPDATE or DELETE)
- squirrel library for dynamic query building
- gofpdf for PDF generation
- Run backend tests: `cd backend && go test ./...`
- Run frontend tests: `cd frontend && pnpm test`

---

## Tasks

- [ ] 1.0 Create Partitioned Audit Table Schema
  - [ ] 1.1 Create migration file `20260127_005_create_audit_tables.sql`
  - [ ] 1.2 Create `audit_events` table with partitioning by created_at
  - [ ] 1.3 Define columns: id, event_type, entity_type, entity_id, user_id, user_email
  - [ ] 1.4 Add action, status, details (JSONB), ip_address (INET), user_agent
  - [ ] 1.5 Set composite primary key (id, created_at) for partitioning
  - [ ] 1.6 Create monthly partitions for current year (2026_01 through 2026_12)
  - [ ] 1.7 Create default partition for future dates
  - [ ] 1.8 Create index on event_type
  - [ ] 1.9 Create index on (entity_type, entity_id)
  - [ ] 1.10 Create index on user_id (partial, WHERE user_id IS NOT NULL)
  - [ ] 1.11 Create index on created_at DESC
  - [ ] 1.12 Create GIN index on details JSONB
  - [ ] 1.13 Create composite index on (event_type, entity_type, created_at DESC)
  - [ ] 1.14 Run migration and verify partitions created

- [ ] 2.0 Implement Audit Repository
  - [ ] 2.1 Create `backend/internal/domain/audit/models.go`
  - [ ] 2.2 Define EventType constants (pull, push, edit, conflict_detected, etc.)
  - [ ] 2.3 Define Event struct with all fields
  - [ ] 2.4 Define QueryFilters struct for search parameters
  - [ ] 2.5 Create `backend/internal/domain/audit/repository.go`
  - [ ] 2.6 Define Repository interface (Insert, Query, QueryAll, GetByID)
  - [ ] 2.7 Create `audit.sql` with InsertAuditEvent query
  - [ ] 2.8 Run sqlc generate for insert query
  - [ ] 2.9 Create `backend/internal/domain/audit/query_builder.go`
  - [ ] 2.10 Import squirrel query builder library
  - [ ] 2.11 Implement buildQuery method with dynamic filters
  - [ ] 2.12 Add event_type filter (IN clause)
  - [ ] 2.13 Add entity_type filter
  - [ ] 2.14 Add entity_id filter
  - [ ] 2.15 Add user_id filter
  - [ ] 2.16 Add status filter
  - [ ] 2.17 Add date range filters (start_date, end_date)
  - [ ] 2.18 Add search text filter (user_email and details ILIKE)
  - [ ] 2.19 Add pagination (LIMIT, OFFSET)
  - [ ] 2.20 Implement Query method using query builder
  - [ ] 2.21 Implement count query for pagination metadata
  - [ ] 2.22 Add repository tests with various filter combinations

- [ ] 3.0 Implement Audit Service
  - [ ] 3.1 Create `backend/internal/domain/audit/service.go`
  - [ ] 3.2 Define Service struct with repository and exporter dependencies
  - [ ] 3.3 Implement Record method for capturing events
  - [ ] 3.4 Enrich events with request context (IP, user agent, user ID)
  - [ ] 3.5 Generate UUID and timestamp for new events
  - [ ] 3.6 Implement Query method with filter parsing
  - [ ] 3.7 Implement GetStats method for dashboard stats
  - [ ] 3.8 Calculate events by type, status, time periods
  - [ ] 3.9 Implement Export method with format selection
  - [ ] 3.10 Create `service_test.go`
  - [ ] 3.11 Test Record method
  - [ ] 3.12 Test Query with filters
  - [ ] 3.13 Test GetStats calculations

- [ ] 4.0 Implement CSV Exporter
  - [ ] 4.1 Create `backend/internal/infrastructure/export/csv.go`
  - [ ] 4.2 Define CSVExporter struct
  - [ ] 4.3 Implement Export method accepting events slice
  - [ ] 4.4 Write header row with column names
  - [ ] 4.5 Iterate events and write data rows
  - [ ] 4.6 Format timestamps as RFC3339
  - [ ] 4.7 Marshal details JSONB as string
  - [ ] 4.8 Handle nil values gracefully
  - [ ] 4.9 Return bytes buffer
  - [ ] 4.10 Add tests for CSV output format

- [ ] 5.0 Implement PDF Exporter
  - [ ] 5.1 Install gofpdf library
  - [ ] 5.2 Create `backend/internal/infrastructure/export/pdf.go`
  - [ ] 5.3 Define PDFExporter struct
  - [ ] 5.4 Implement Export method
  - [ ] 5.5 Configure landscape A4 layout
  - [ ] 5.6 Add report title and generation timestamp
  - [ ] 5.7 Add total events count
  - [ ] 5.8 Create table header with column titles
  - [ ] 5.9 Set appropriate column widths
  - [ ] 5.10 Iterate events and add table rows
  - [ ] 5.11 Handle page breaks with header repeat
  - [ ] 5.12 Truncate long content to fit columns
  - [ ] 5.13 Return bytes buffer
  - [ ] 5.14 Add tests for PDF generation

- [ ] 6.0 Implement Audit API Handlers
  - [ ] 6.1 Create `backend/internal/api/handlers/audit/schemas.go`
  - [ ] 6.2 Define QueryEventsRequest with filter parameters
  - [ ] 6.3 Define QueryEventsResponse with events and pagination
  - [ ] 6.4 Define EventResponse with all event fields
  - [ ] 6.5 Define PaginationMeta (page, per_page, total_count, total_pages)
  - [ ] 6.6 Define ExportEventsRequest with format parameter
  - [ ] 6.7 Define StatsResponse with aggregated counts
  - [ ] 6.8 Create `backend/internal/api/handlers/audit/handler.go`
  - [ ] 6.9 Implement QueryEvents handler (GET /api/v1/audit)
  - [ ] 6.10 Parse query parameters into QueryFilters
  - [ ] 6.11 Implement GetEvent handler (GET /api/v1/audit/{id})
  - [ ] 6.12 Implement ExportEvents handler (GET /api/v1/audit/export)
  - [ ] 6.13 Set Content-Type and Content-Disposition headers for download
  - [ ] 6.14 Implement GetStats handler (GET /api/v1/audit/stats)
  - [ ] 6.15 Add rate limiting on export endpoint
  - [ ] 6.16 Create `handler_test.go`
  - [ ] 6.17 Test query with various filters
  - [ ] 6.18 Test export file generation

- [ ] 7.0 Integrate Audit Recording in Other Features
  - [ ] 7.1 Add audit service dependency to Connection Service (F1)
  - [ ] 7.2 Record connection_config events in SaveConfig
  - [ ] 7.3 Record connection_test events in TestConnection
  - [ ] 7.4 Add audit service dependency to Pull Executor (F2)
  - [ ] 7.5 Record pull events with system/controls/statements counts
  - [ ] 7.6 Add audit service dependency to Statement Service (F3)
  - [ ] 7.7 Record edit events with content length and session_id
  - [ ] 7.8 Add audit service dependency to Push Executor (F4)
  - [ ] 7.9 Record push events with success/failure status
  - [ ] 7.10 Record conflict_detected events
  - [ ] 7.11 Record conflict_resolved events with resolution choice
  - [ ] 7.12 Verify audit events generated for all operations

- [ ] 8.0 Implement Async Event Recording (Optional Optimization)
  - [ ] 8.1 Create AsyncAuditService wrapper
  - [ ] 8.2 Implement buffered channel for events
  - [ ] 8.3 Start background goroutine for processing
  - [ ] 8.4 Handle buffer overflow gracefully (log warning, don't fail)
  - [ ] 8.5 Add graceful shutdown to drain buffer
  - [ ] 8.6 Add tests for async behavior

- [ ] 9.0 Implement Frontend API Layer
  - [ ] 9.1 Create `frontend/src/features/audit/types.ts`
  - [ ] 9.2 Define AuditEvent interface
  - [ ] 9.3 Define AuditFilters interface
  - [ ] 9.4 Define AuditStats interface
  - [ ] 9.5 Create `frontend/src/features/audit/api/auditApi.ts`
  - [ ] 9.6 Implement queryEvents API function with filters
  - [ ] 9.7 Implement getEvent API function
  - [ ] 9.8 Implement exportEvents API function (returns blob)
  - [ ] 9.9 Implement getStats API function
  - [ ] 9.10 Create `frontend/src/features/audit/hooks/useAuditEvents.ts`
  - [ ] 9.11 Implement useAuditEvents query with keepPreviousData
  - [ ] 9.12 Implement useAuditEvent query for single event
  - [ ] 9.13 Implement useAuditStats query
  - [ ] 9.14 Create `frontend/src/features/audit/hooks/useExport.ts`
  - [ ] 9.15 Implement useExport mutation with file download

- [ ] 10.0 Implement Audit Filters Component
  - [ ] 10.1 Create `frontend/src/features/audit/components/AuditFilters.tsx`
  - [ ] 10.2 Define EVENT_TYPES array with label/value pairs
  - [ ] 10.3 Define ENTITY_TYPES array
  - [ ] 10.4 Add event type multi-select
  - [ ] 10.5 Add entity type multi-select
  - [ ] 10.6 Add date range picker (from/to)
  - [ ] 10.7 Add search input with icon
  - [ ] 10.8 Add status select (all/success/failure)
  - [ ] 10.9 Reset page to 0 when filters change
  - [ ] 10.10 Add "Clear Filters" button
  - [ ] 10.11 Display active filters summary
  - [ ] 10.12 Create `AuditFilters.test.tsx`
  - [ ] 10.13 Test filter onChange behavior
  - [ ] 10.14 Test page reset on filter change

- [ ] 11.0 Implement Audit Table Component
  - [ ] 11.1 Create `frontend/src/features/audit/components/AuditTable.tsx`
  - [ ] 11.2 Define columns with ColumnDef type
  - [ ] 11.3 Add Time column with date formatting
  - [ ] 11.4 Add Event column with type badge
  - [ ] 11.5 Add Entity column with type:id format
  - [ ] 11.6 Add Action column
  - [ ] 11.7 Add User column (email or "System")
  - [ ] 11.8 Add Status column with colored badge
  - [ ] 11.9 Add actions column with view button
  - [ ] 11.10 Implement DataTable with columns and data
  - [ ] 11.11 Add loading skeleton
  - [ ] 11.12 Add empty state message
  - [ ] 11.13 Create `AuditTable.test.tsx`

- [ ] 12.0 Implement Audit Detail Modal
  - [ ] 12.1 Create `frontend/src/features/audit/components/AuditDetail.tsx`
  - [ ] 12.2 Fetch event details with useAuditEvent hook
  - [ ] 12.3 Display as Dialog component
  - [ ] 12.4 Show all event fields in grid layout
  - [ ] 12.5 Format timestamp with full date/time
  - [ ] 12.6 Display details JSONB as formatted pre block
  - [ ] 12.7 Add loading state
  - [ ] 12.8 Add close button

- [ ] 13.0 Implement Export Dialog Component
  - [ ] 13.1 Create `frontend/src/features/audit/components/ExportDialog.tsx`
  - [ ] 13.2 Display as Dialog with format selection
  - [ ] 13.3 Add CSV and PDF radio options
  - [ ] 13.4 Show current filter summary
  - [ ] 13.5 Add "Export" button with loading state
  - [ ] 13.6 Trigger file download on success
  - [ ] 13.7 Handle export errors

- [ ] 14.0 Implement Main Audit Log Component
  - [ ] 14.1 Create `frontend/src/features/audit/components/AuditLog.tsx`
  - [ ] 14.2 Initialize filters state with defaults
  - [ ] 14.3 Track selected event for detail modal
  - [ ] 14.4 Use useAuditEvents hook with filters
  - [ ] 14.5 Render page header with title and export button
  - [ ] 14.6 Render AuditFilters component
  - [ ] 14.7 Render AuditTable component
  - [ ] 14.8 Render pagination controls
  - [ ] 14.9 Render AuditDetail modal when event selected
  - [ ] 14.10 Render ExportDialog when open

- [ ] 15.0 Create Audit Page
  - [ ] 15.1 Create `frontend/src/pages/AuditPage.tsx`
  - [ ] 15.2 Render AuditLog component
  - [ ] 15.3 Add route definition

- [ ] 16.0 Implement Partition Management (Production)
  - [ ] 16.1 Create partition management SQL function
  - [ ] 16.2 Implement create next month partition logic
  - [ ] 16.3 Implement drop partitions older than 7 years
  - [ ] 16.4 Document pg_cron or external scheduler setup
  - [ ] 16.5 Add monitoring for partition health

- [ ] 17.0 End-to-End Integration
  - [ ] 17.1 Perform connection test (F1) and verify audit event
  - [ ] 17.2 Perform pull operation (F2) and verify audit event
  - [ ] 17.3 Edit statement (F3) and verify audit event
  - [ ] 17.4 Push statement (F4) and verify audit event
  - [ ] 17.5 Navigate to audit page
  - [ ] 17.6 Test filter by event type
  - [ ] 17.7 Test filter by date range
  - [ ] 17.8 Test search functionality
  - [ ] 17.9 Test export to CSV
  - [ ] 17.10 Test export to PDF
  - [ ] 17.11 Verify event details display correctly
