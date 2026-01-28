# Task List: Control Package Pull (F2)

**Feature ID:** F2
**Related PRD:** 002_FPRD|Control_Package_Pull.md
**Related TDD:** 002_FTDD|Control_Package_Pull.md
**Related TID:** 002_FTID|Control_Package_Pull.md
**Depends On:** F1 (ServiceNow Connection)

---

## Relevant Files

### Backend - Database
- `backend/migrations/20260127_002_create_pull_tables.sql` - Systems, controls, statements, pull_jobs tables
- `backend/internal/infrastructure/database/queries/system.sql` - System sqlc queries
- `backend/internal/infrastructure/database/queries/control.sql` - Control sqlc queries
- `backend/internal/infrastructure/database/queries/statement.sql` - Statement sqlc queries

### Backend - Domain (System)
- `backend/internal/domain/system/models.go` - System domain models
- `backend/internal/domain/system/service.go` - System business logic
- `backend/internal/domain/system/service_test.go` - System service tests
- `backend/internal/domain/system/repository.go` - System repository interface

### Backend - Domain (Control)
- `backend/internal/domain/control/models.go` - Control domain models
- `backend/internal/domain/control/service.go` - Control business logic
- `backend/internal/domain/control/service_test.go` - Control service tests
- `backend/internal/domain/control/repository.go` - Control repository interface

### Backend - Domain (Statement)
- `backend/internal/domain/statement/models.go` - Statement domain models
- `backend/internal/domain/statement/service.go` - Statement business logic
- `backend/internal/domain/statement/service_test.go` - Statement service tests
- `backend/internal/domain/statement/repository.go` - Statement repository interface

### Backend - Jobs
- `backend/internal/jobs/pull/executor.go` - Pull job execution logic
- `backend/internal/jobs/pull/executor_test.go` - Pull executor tests
- `backend/internal/jobs/pull/progress.go` - Progress tracking for pull jobs

### Backend - Infrastructure
- `backend/internal/infrastructure/servicenow/client.go` - Add pull methods (modify)
- `backend/internal/infrastructure/servicenow/pagination.go` - Pagination handler

### Backend - API
- `backend/internal/api/handlers/sync/handler.go` - Sync HTTP handlers
- `backend/internal/api/handlers/sync/handler_test.go` - Sync handler tests
- `backend/internal/api/handlers/sync/schemas.go` - Sync request/response DTOs

### Frontend - Systems Feature
- `frontend/src/features/systems/components/SystemList.tsx` - System listing component
- `frontend/src/features/systems/components/SystemCard.tsx` - Individual system card
- `frontend/src/features/systems/components/SystemSelector.tsx` - Multi-select for pull
- `frontend/src/features/systems/hooks/useSystems.ts` - System query hooks
- `frontend/src/features/systems/api/systemsApi.ts` - Systems API client
- `frontend/src/features/systems/types.ts` - System TypeScript types

### Frontend - Pull Feature
- `frontend/src/features/pull/components/PullWizard.tsx` - Multi-step pull flow
- `frontend/src/features/pull/components/PullProgress.tsx` - Progress indicator
- `frontend/src/features/pull/components/PullResults.tsx` - Results summary
- `frontend/src/features/pull/hooks/usePull.ts` - Pull mutation hooks
- `frontend/src/features/pull/hooks/usePullStatus.ts` - Status polling hook
- `frontend/src/features/pull/api/pullApi.ts` - Pull API client
- `frontend/src/features/pull/types.ts` - Pull TypeScript types

### Frontend - Pages
- `frontend/src/pages/PullPage.tsx` - Pull workflow page

### Notes

- Pull operations run as background jobs to avoid HTTP timeout
- Frontend polls job status every 2 seconds while running
- ServiceNow pagination uses sysparm_offset and sysparm_limit
- Run backend tests: `cd backend && go test ./...`
- Run frontend tests: `cd frontend && pnpm test`

---

## Tasks

- [ ] 1.0 Create Database Migration for Pull Tables
  - [ ] 1.1 Create migration file `20260127_002_create_pull_tables.sql`
  - [ ] 1.2 Create `systems` table with sn_sys_id, name, description, acronym, owner, status
  - [ ] 1.3 Add sync metadata columns to systems (sn_updated_on, last_pull_at, last_push_at)
  - [ ] 1.4 Create unique constraint on systems.sn_sys_id
  - [ ] 1.5 Create `controls` table with system_id FK, control_id, control_name, control_family
  - [ ] 1.6 Add implementation_status and responsible_role to controls
  - [ ] 1.7 Create unique constraint on (system_id, sn_sys_id) for controls
  - [ ] 1.8 Create `statements` table with control_id FK, statement_type, content fields
  - [ ] 1.9 Add local editing columns (local_content, is_modified, modified_at, sync_status)
  - [ ] 1.10 Create unique constraint on (control_id, sn_sys_id) for statements
  - [ ] 1.11 Create `pull_jobs` table with status, system_ids array, progress JSONB
  - [ ] 1.12 Add indexes on foreign keys and common query columns
  - [ ] 1.13 Run migration and verify all tables created

- [ ] 2.0 Implement ServiceNow Pagination Handler
  - [ ] 2.1 Create `backend/internal/infrastructure/servicenow/pagination.go`
  - [ ] 2.2 Define `PaginatedResponse` struct for parsing ServiceNow responses
  - [ ] 2.3 Implement `FetchAllPages` method with offset/limit pagination
  - [ ] 2.4 Add configurable page size (default 100)
  - [ ] 2.5 Handle X-Total-Count header for progress tracking
  - [ ] 2.6 Add rate limiting awareness and exponential backoff
  - [ ] 2.7 Update `client.go` with methods: FetchSystems, FetchControls, FetchStatements
  - [ ] 2.8 Add tests for pagination with multiple pages
  - [ ] 2.9 Add tests for empty results handling

- [ ] 3.0 Implement System Domain Layer
  - [ ] 3.1 Create `backend/internal/domain/system/models.go`
  - [ ] 3.2 Define `System` struct with all fields
  - [ ] 3.3 Create `backend/internal/domain/system/repository.go`
  - [ ] 3.4 Define Repository interface (GetByID, GetBySNSysID, List, Upsert)
  - [ ] 3.5 Create `system.sql` with sqlc queries
  - [ ] 3.6 Implement UpsertSystem query with ON CONFLICT handling
  - [ ] 3.7 Implement ListSystems with pagination
  - [ ] 3.8 Run sqlc generate
  - [ ] 3.9 Create `backend/internal/domain/system/service.go`
  - [ ] 3.10 Implement Service with repository dependency
  - [ ] 3.11 Implement DiscoverSystems method calling ServiceNow
  - [ ] 3.12 Implement ListLocalSystems method
  - [ ] 3.13 Create `service_test.go` with mocked dependencies
  - [ ] 3.14 Add tests for system discovery and listing

- [ ] 4.0 Implement Control Domain Layer
  - [ ] 4.1 Create `backend/internal/domain/control/models.go`
  - [ ] 4.2 Define `Control` struct with system relationship
  - [ ] 4.3 Add NIST 800-53 family mapping helper function
  - [ ] 4.4 Create `backend/internal/domain/control/repository.go`
  - [ ] 4.5 Define Repository interface (GetByID, ListBySystem, Upsert, UpsertBatch)
  - [ ] 4.6 Create `control.sql` with sqlc queries
  - [ ] 4.7 Implement UpsertControl with ON CONFLICT
  - [ ] 4.8 Run sqlc generate
  - [ ] 4.9 Create `backend/internal/domain/control/service.go`
  - [ ] 4.10 Implement batch upsert for efficiency
  - [ ] 4.11 Create `service_test.go`
  - [ ] 4.12 Add tests for control family extraction

- [ ] 5.0 Implement Statement Domain Layer
  - [ ] 5.1 Create `backend/internal/domain/statement/models.go`
  - [ ] 5.2 Define `Statement` struct with sync_status enum
  - [ ] 5.3 Define SyncStatus constants (synced, modified, conflict)
  - [ ] 5.4 Create `backend/internal/domain/statement/repository.go`
  - [ ] 5.5 Define Repository interface with conflict-aware upsert
  - [ ] 5.6 Create `statement.sql` with sqlc queries
  - [ ] 5.7 Implement UpsertStatement preserving local modifications
  - [ ] 5.8 Implement conflict detection in upsert query
  - [ ] 5.9 Run sqlc generate
  - [ ] 5.10 Create `backend/internal/domain/statement/service.go`
  - [ ] 5.11 Implement conflict detection logic
  - [ ] 5.12 Create `service_test.go`
  - [ ] 5.13 Add tests for conflict scenarios

- [ ] 6.0 Implement Pull Job Executor
  - [ ] 6.1 Create `backend/internal/jobs/pull/progress.go`
  - [ ] 6.2 Define `PullProgress` struct with counters
  - [ ] 6.3 Create `backend/internal/jobs/pull/executor.go`
  - [ ] 6.4 Define `PullJob` struct with status enum
  - [ ] 6.5 Define `PullExecutor` struct with all dependencies
  - [ ] 6.6 Implement `Execute` method processing systems sequentially
  - [ ] 6.7 Implement `pullSystem` method fetching controls and statements
  - [ ] 6.8 Add data transformation from ServiceNow response to domain models
  - [ ] 6.9 Implement progress tracking with throttled updates
  - [ ] 6.10 Handle partial failures (continue with remaining systems)
  - [ ] 6.11 Add context cancellation support
  - [ ] 6.12 Create `executor_test.go`
  - [ ] 6.13 Add tests for successful pull
  - [ ] 6.14 Add tests for partial failure scenarios

- [ ] 7.0 Implement Sync API Handlers
  - [ ] 7.1 Create `backend/internal/api/handlers/sync/schemas.go`
  - [ ] 7.2 Define DiscoverSystemsResponse with isImported flag
  - [ ] 7.3 Define StartPullRequest with system_ids validation
  - [ ] 7.4 Define PullStatusResponse with progress details
  - [ ] 7.5 Create `backend/internal/api/handlers/sync/handler.go`
  - [ ] 7.6 Implement `DiscoverSystems` handler (GET /api/v1/sync/systems/discover)
  - [ ] 7.7 Implement `ListSystems` handler (GET /api/v1/sync/systems)
  - [ ] 7.8 Implement `StartPull` handler creating job record (POST /api/v1/sync/pull)
  - [ ] 7.9 Implement `GetPullStatus` handler (GET /api/v1/sync/pull/{jobId})
  - [ ] 7.10 Implement `CancelPull` handler (DELETE /api/v1/sync/pull/{jobId})
  - [ ] 7.11 Create `handler_test.go`
  - [ ] 7.12 Add tests for all endpoints
  - [ ] 7.13 Add validation error tests

- [ ] 8.0 Implement Background Job Processor
  - [ ] 8.1 Create `backend/internal/jobs/processor.go`
  - [ ] 8.2 Implement job processor polling for pending jobs
  - [ ] 8.3 Add graceful shutdown with context cancellation
  - [ ] 8.4 Start processor in `cmd/server/main.go`
  - [ ] 8.5 Add job status update on completion/failure
  - [ ] 8.6 Add tests for processor lifecycle

- [ ] 9.0 Implement Frontend API Layer
  - [ ] 9.1 Create `frontend/src/features/systems/types.ts`
  - [ ] 9.2 Create `frontend/src/features/pull/types.ts`
  - [ ] 9.3 Create `frontend/src/features/systems/api/systemsApi.ts`
  - [ ] 9.4 Implement discoverSystems API function
  - [ ] 9.5 Implement listSystems API function
  - [ ] 9.6 Create `frontend/src/features/pull/api/pullApi.ts`
  - [ ] 9.7 Implement startPull API function
  - [ ] 9.8 Implement getPullStatus API function
  - [ ] 9.9 Implement cancelPull API function
  - [ ] 9.10 Create `frontend/src/features/systems/hooks/useSystems.ts`
  - [ ] 9.11 Implement useDiscoverSystems query hook
  - [ ] 9.12 Implement useLocalSystems query hook
  - [ ] 9.13 Create `frontend/src/features/pull/hooks/usePull.ts`
  - [ ] 9.14 Implement useStartPull mutation hook
  - [ ] 9.15 Create `frontend/src/features/pull/hooks/usePullStatus.ts`
  - [ ] 9.16 Implement polling hook with 2-second interval while running

- [ ] 10.0 Implement System Selection Components
  - [ ] 10.1 Create `frontend/src/features/systems/components/SystemCard.tsx`
  - [ ] 10.2 Display system name, acronym, description
  - [ ] 10.3 Show "Already imported" badge for existing systems
  - [ ] 10.4 Add checkbox/selection state
  - [ ] 10.5 Create `frontend/src/features/systems/components/SystemList.tsx`
  - [ ] 10.6 Display grid of SystemCards
  - [ ] 10.7 Add loading skeleton
  - [ ] 10.8 Create `frontend/src/features/systems/components/SystemSelector.tsx`
  - [ ] 10.9 Implement multi-select with max limit (10 systems)
  - [ ] 10.10 Add "Select All" functionality
  - [ ] 10.11 Display selection count

- [ ] 11.0 Implement Pull Workflow Components
  - [ ] 11.1 Create `frontend/src/features/pull/components/PullProgress.tsx`
  - [ ] 11.2 Display progress bar for systems completed
  - [ ] 11.3 Show controls/statements counts
  - [ ] 11.4 Display status icon (running/completed/failed)
  - [ ] 11.5 Handle error display with alert
  - [ ] 11.6 Create `frontend/src/features/pull/components/PullResults.tsx`
  - [ ] 11.7 Display summary of imported items
  - [ ] 11.8 Show any errors that occurred
  - [ ] 11.9 Add "Done" button to return to start
  - [ ] 11.10 Create `frontend/src/features/pull/components/PullWizard.tsx`
  - [ ] 11.11 Implement step state machine (select, confirm, progress, results)
  - [ ] 11.12 Add Stepper component for visual progress
  - [ ] 11.13 Implement navigation between steps
  - [ ] 11.14 Create `frontend/src/pages/PullPage.tsx`
  - [ ] 11.15 Add route for pull page

- [ ] 12.0 Frontend Testing
  - [ ] 12.1 Create tests for SystemCard component
  - [ ] 12.2 Create tests for SystemSelector multi-select behavior
  - [ ] 12.3 Create tests for PullProgress polling
  - [ ] 12.4 Create tests for PullWizard step transitions
  - [ ] 12.5 Run all frontend tests

- [ ] 13.0 End-to-End Integration
  - [ ] 13.1 Verify system discovery from ServiceNow works
  - [ ] 13.2 Test pull job creation and status polling
  - [ ] 13.3 Verify controls and statements imported correctly
  - [ ] 13.4 Test sync_status set correctly for new statements
  - [ ] 13.5 Test partial failure handling
  - [ ] 13.6 Verify progress updates in UI during pull
