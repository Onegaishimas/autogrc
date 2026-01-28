# Task List: Control Package Push (F4)

**Feature ID:** F4
**Related PRD:** 004_FPRD|Control_Package_Push.md
**Related TDD:** 004_FTDD|Control_Package_Push.md
**Related TID:** 004_FTID|Control_Package_Push.md
**Depends On:** F1 (Connection), F2 (Pull), F3 (Editor)

---

## Relevant Files

### Backend - Database
- `backend/migrations/20260127_004_create_push_tables.sql` - Push jobs table
- `backend/internal/infrastructure/database/queries/statement.sql` - Add push queries (modify)
- `backend/internal/infrastructure/database/queries/push_job.sql` - Push job queries

### Backend - Domain
- `backend/internal/domain/push/models.go` - Push domain models
- `backend/internal/domain/push/service.go` - Push business logic
- `backend/internal/domain/push/service_test.go` - Push service tests
- `backend/internal/domain/push/conflict.go` - Conflict detection logic
- `backend/internal/domain/push/executor.go` - Push execution logic

### Backend - Infrastructure
- `backend/internal/infrastructure/servicenow/client.go` - Add PUT methods (modify)

### Backend - API
- `backend/internal/api/handlers/push/handler.go` - Push HTTP handlers
- `backend/internal/api/handlers/push/handler_test.go` - Handler tests
- `backend/internal/api/handlers/push/schemas.go` - Push DTOs

### Frontend - Push Feature
- `frontend/src/features/push/components/PushWorkflow.tsx` - Main push flow
- `frontend/src/features/push/components/ModifiedList.tsx` - Modified statements list
- `frontend/src/features/push/components/ConflictResolution.tsx` - Conflict diff viewer
- `frontend/src/features/push/components/PushPreview.tsx` - Pre-push preview
- `frontend/src/features/push/components/PushProgress.tsx` - Push progress
- `frontend/src/features/push/components/PushResults.tsx` - Results summary
- `frontend/src/features/push/hooks/useModifiedStatements.ts` - Query modified
- `frontend/src/features/push/hooks/usePush.ts` - Push mutation hooks
- `frontend/src/features/push/hooks/usePushStatus.ts` - Status polling
- `frontend/src/features/push/hooks/useConflicts.ts` - Conflict detection
- `frontend/src/features/push/api/pushApi.ts` - API client
- `frontend/src/features/push/types.ts` - TypeScript types

### Frontend - Pages
- `frontend/src/pages/PushPage.tsx` - Push workflow page

### Frontend - Tests
- `frontend/src/features/push/components/ModifiedList.test.tsx` - List tests
- `frontend/src/features/push/components/ConflictResolution.test.tsx` - Conflict tests

### Notes

- Conflict detection compares sys_updated_on timestamps
- Push operations run concurrently with semaphore limit (5)
- Monaco diff editor for conflict visualization
- Run frontend tests: `cd frontend && pnpm test`

---

## Tasks

- [ ] 1.0 Create Push Job Database Schema
  - [ ] 1.1 Create migration file `20260127_004_create_push_tables.sql`
  - [ ] 1.2 Create `push_status` enum type (pending, running, completed, failed)
  - [ ] 1.3 Create `push_jobs` table with status, statement_ids array, results JSONB
  - [ ] 1.4 Add started_at, completed_at timestamp columns
  - [ ] 1.5 Add created_by foreign key to users
  - [ ] 1.6 Create index on status for pending/running jobs
  - [ ] 1.7 Create index on created_by for user filtering
  - [ ] 1.8 Run migration and verify schema

- [ ] 2.0 Add Statement Push Queries
  - [ ] 2.1 Add GetModifiedStatements query to statement.sql
  - [ ] 2.2 Join with controls and systems for display info
  - [ ] 2.3 Add GetConflictedStatements query
  - [ ] 2.4 Add MarkStatementPushed query
  - [ ] 2.5 Implement content promotion (local → canonical)
  - [ ] 2.6 Add ClearConflict query
  - [ ] 2.7 Add OverwriteLocal query for keep_remote resolution
  - [ ] 2.8 Run sqlc generate

- [ ] 3.0 Update ServiceNow Client for Push
  - [ ] 3.1 Add UpdateStatement method to client.go
  - [ ] 3.2 Implement PUT request to sn_compliance_statement table
  - [ ] 3.3 Handle 404 (statement deleted) error
  - [ ] 3.4 Handle 401 (auth expired) error
  - [ ] 3.5 Parse updated sys_updated_on from response
  - [ ] 3.6 Add tests for update scenarios

- [ ] 4.0 Implement Conflict Detection
  - [ ] 4.1 Create `backend/internal/domain/push/conflict.go`
  - [ ] 4.2 Define ConflictResult struct with local/remote versions
  - [ ] 4.3 Define ConflictChecker struct with dependencies
  - [ ] 4.4 Implement CheckConflicts method fetching remote timestamps
  - [ ] 4.5 Compare sys_updated_on timestamps
  - [ ] 4.6 Return conflict results with content for diff display
  - [ ] 4.7 Add tests for conflict detection scenarios
  - [ ] 4.8 Test no conflict when remote unchanged
  - [ ] 4.9 Test conflict when remote is newer

- [ ] 5.0 Implement Push Executor
  - [ ] 5.1 Create `backend/internal/domain/push/executor.go`
  - [ ] 5.2 Define PushResult struct with success/error
  - [ ] 5.3 Define PushExecutor struct with dependencies
  - [ ] 5.4 Implement ExecutePush method with concurrent pushes
  - [ ] 5.5 Use semaphore channel for concurrency limit (5)
  - [ ] 5.6 Implement pushStatement helper method
  - [ ] 5.7 Update local state on successful push (MarkStatementPushed)
  - [ ] 5.8 Record audit event for each push (F5 integration)
  - [ ] 5.9 Handle individual statement failures (continue with others)
  - [ ] 5.10 Add tests for executor
  - [ ] 5.11 Test concurrent push behavior
  - [ ] 5.12 Test partial failure scenarios

- [ ] 6.0 Implement Push Service
  - [ ] 6.1 Create `backend/internal/domain/push/service.go`
  - [ ] 6.2 Implement GetModifiedStatements method
  - [ ] 6.3 Implement CheckConflicts method
  - [ ] 6.4 Implement ResolveConflict method (keep_local or keep_remote)
  - [ ] 6.5 Validate statement is in conflict state before resolving
  - [ ] 6.6 Record conflict resolution in audit log
  - [ ] 6.7 Implement StartPush method creating push job
  - [ ] 6.8 Implement GetPushStatus method
  - [ ] 6.9 Implement GetPushResults method
  - [ ] 6.10 Create `service_test.go`
  - [ ] 6.11 Test ResolveConflict scenarios

- [ ] 7.0 Implement Push API Handlers
  - [ ] 7.1 Create `backend/internal/api/handlers/push/schemas.go`
  - [ ] 7.2 Define ModifiedStatementsResponse with content preview
  - [ ] 7.3 Define CheckConflictsRequest/Response
  - [ ] 7.4 Define ResolveConflictRequest (keep_local/keep_remote)
  - [ ] 7.5 Define StartPushRequest/Response
  - [ ] 7.6 Define PushStatusResponse with progress counters
  - [ ] 7.7 Define PushResultsResponse with per-statement results
  - [ ] 7.8 Create `backend/internal/api/handlers/push/handler.go`
  - [ ] 7.9 Implement GetModifiedStatements (GET /api/v1/push/modified)
  - [ ] 7.10 Implement CheckConflicts (POST /api/v1/push/check-conflicts)
  - [ ] 7.11 Implement ResolveConflict (POST /api/v1/push/resolve-conflict/{id})
  - [ ] 7.12 Implement StartPush (POST /api/v1/push)
  - [ ] 7.13 Implement GetPushStatus (GET /api/v1/push/{jobId})
  - [ ] 7.14 Implement GetPushResults (GET /api/v1/push/{jobId}/results)
  - [ ] 7.15 Create `handler_test.go`
  - [ ] 7.16 Test all endpoints

- [ ] 8.0 Implement Frontend API Layer
  - [ ] 8.1 Create `frontend/src/features/push/types.ts`
  - [ ] 8.2 Define ModifiedStatement interface
  - [ ] 8.3 Define ConflictInfo interface
  - [ ] 8.4 Define PushStatus and PushResult interfaces
  - [ ] 8.5 Create `frontend/src/features/push/api/pushApi.ts`
  - [ ] 8.6 Implement getModifiedStatements API function
  - [ ] 8.7 Implement checkConflicts API function
  - [ ] 8.8 Implement resolveConflict API function
  - [ ] 8.9 Implement startPush API function
  - [ ] 8.10 Implement getPushStatus API function
  - [ ] 8.11 Implement getPushResults API function
  - [ ] 8.12 Create query and mutation hooks
  - [ ] 8.13 Implement useModifiedStatements query
  - [ ] 8.14 Implement useConflicts mutation
  - [ ] 8.15 Implement useResolveConflict mutation
  - [ ] 8.16 Implement useStartPush mutation
  - [ ] 8.17 Implement usePushStatus polling hook

- [ ] 9.0 Implement Modified List Component
  - [ ] 9.1 Create `frontend/src/features/push/components/ModifiedList.tsx`
  - [ ] 9.2 Display list of modified statements as cards
  - [ ] 9.3 Show control ID, name, system name
  - [ ] 9.4 Show modified timestamp with formatDistanceToNow
  - [ ] 9.5 Show conflict badge for conflicted statements
  - [ ] 9.6 Implement checkbox selection
  - [ ] 9.7 Disable selection for conflicted statements
  - [ ] 9.8 Add "Select All (Non-Conflicted)" button
  - [ ] 9.9 Show selection count
  - [ ] 9.10 Create `ModifiedList.test.tsx`
  - [ ] 9.11 Test selection behavior
  - [ ] 9.12 Test conflict badge display

- [ ] 10.0 Implement Conflict Resolution Component
  - [ ] 10.1 Create `frontend/src/features/push/components/ConflictResolution.tsx`
  - [ ] 10.2 Implement as Dialog component
  - [ ] 10.3 Show control ID in title
  - [ ] 10.4 Display local and remote timestamps
  - [ ] 10.5 Integrate Monaco DiffEditor for side-by-side comparison
  - [ ] 10.6 Set editor to read-only mode
  - [ ] 10.7 Add "Keep Local Version" button
  - [ ] 10.8 Add "Keep ServiceNow Version" button
  - [ ] 10.9 Call resolveConflict mutation on choice
  - [ ] 10.10 Create `ConflictResolution.test.tsx`
  - [ ] 10.11 Test diff display
  - [ ] 10.12 Test resolution button behavior

- [ ] 11.0 Implement Push Preview Component
  - [ ] 11.1 Create `frontend/src/features/push/components/PushPreview.tsx`
  - [ ] 11.2 Display summary of statements to be pushed
  - [ ] 11.3 Show total count
  - [ ] 11.4 List statement control IDs
  - [ ] 11.5 Add confirmation warning text
  - [ ] 11.6 Add "Confirm Push" button
  - [ ] 11.7 Add "Back" button

- [ ] 12.0 Implement Push Progress Component
  - [ ] 12.1 Create `frontend/src/features/push/components/PushProgress.tsx`
  - [ ] 12.2 Display progress bar (completed/total)
  - [ ] 12.3 Show succeeded and failed counts
  - [ ] 12.4 Display status icon (running/completed/failed)
  - [ ] 12.5 Poll status every 2 seconds while running
  - [ ] 12.6 Stop polling when completed or failed
  - [ ] 12.7 Call onComplete callback when done

- [ ] 13.0 Implement Push Results Component
  - [ ] 13.1 Create `frontend/src/features/push/components/PushResults.tsx`
  - [ ] 13.2 Display summary (X succeeded, Y failed)
  - [ ] 13.3 Show success icon for completed statements
  - [ ] 13.4 Show error icon and message for failed statements
  - [ ] 13.5 Add user-friendly error messages
  - [ ] 13.6 Add "Done" button to reset workflow

- [ ] 14.0 Implement Push Workflow Component
  - [ ] 14.1 Create `frontend/src/features/push/components/PushWorkflow.tsx`
  - [ ] 14.2 Define step state machine (select, conflicts, preview, progress, results)
  - [ ] 14.3 Track selected statement IDs in state
  - [ ] 14.4 Track current job ID in state
  - [ ] 14.5 Add Stepper component for visual progress
  - [ ] 14.6 Render ModifiedList on select step
  - [ ] 14.7 Render conflict resolution on conflicts step
  - [ ] 14.8 Render PushPreview on preview step
  - [ ] 14.9 Render PushProgress on progress step
  - [ ] 14.10 Render PushResults on results step
  - [ ] 14.11 Implement step navigation (back/continue)
  - [ ] 14.12 Reset state on "Done"

- [ ] 15.0 Implement Conflict Resolution Step
  - [ ] 15.1 Create ConflictStep component within PushWorkflow
  - [ ] 15.2 Check for conflicts when entering step
  - [ ] 15.3 Display list of conflicts to resolve
  - [ ] 15.4 Open ConflictResolution dialog for each conflict
  - [ ] 15.5 Track resolved conflicts
  - [ ] 15.6 Proceed to preview when all conflicts resolved
  - [ ] 15.7 Skip step if no conflicts

- [ ] 16.0 Create Push Page
  - [ ] 16.1 Create `frontend/src/pages/PushPage.tsx`
  - [ ] 16.2 Render PushWorkflow component
  - [ ] 16.3 Add page header
  - [ ] 16.4 Add route definition

- [ ] 17.0 End-to-End Integration
  - [ ] 17.1 Modify statements in editor (F3)
  - [ ] 17.2 Navigate to push page
  - [ ] 17.3 Verify modified statements appear in list
  - [ ] 17.4 Select statements and check for conflicts
  - [ ] 17.5 Test conflict resolution workflow
  - [ ] 17.6 Test push execution
  - [ ] 17.7 Verify statements updated in ServiceNow (if available)
  - [ ] 17.8 Verify local state cleared (is_modified = false)
  - [ ] 17.9 Test partial failure handling
