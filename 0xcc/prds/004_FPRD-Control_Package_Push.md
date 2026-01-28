# Feature PRD: Control Package Push (F4)

**Document Version:** 1.0
**Date:** 2026-01-27
**Feature ID:** F4
**Phase:** 1 (Core Sync MVP)
**Priority:** Must Have

---

## 1. Feature Overview

### Feature Name
Control Package Push

### Brief Description
Enable users to push (export) modified implementation statements from ControlCRUD back to ServiceNow GRC, completing the bidirectional sync loop.

### Problem Statement
After editing implementation statements in ControlCRUD, users need a reliable way to synchronize changes back to ServiceNow GRC. Manual copy/paste is error-prone and time-consuming. Users need confidence that their work will be accurately reflected in the system of record.

### Feature Goals
1. Push modified statements to ServiceNow GRC
2. Detect and handle conflicts (ServiceNow changed since pull)
3. Provide clear success/failure feedback
4. Update evidence references if modified
5. Maintain audit trail of push operations

### User Value Proposition
*"As a Control Author, I can push my edited statements to ServiceNow with one click, knowing changes will be accurately saved and I'll be warned of any conflicts."*

### Connection to Project Objectives
- **Bidirectional sync:** This is the "push" half of the core sync loop
- **Single source of truth:** ServiceNow GRC remains the authoritative record
- **Efficiency:** Eliminates manual copy/paste back to ServiceNow

---

## 2. User Stories & Scenarios

### Primary User Stories

**US-4.1: Push Modified Statements**
> As a Control Author, I want to push my modified statements to ServiceNow so that my work is saved to the system of record.

**Acceptance Criteria:**
- [ ] Can initiate push with single action
- [ ] Only modified statements are pushed
- [ ] Progress indicator during push
- [ ] Success message with summary (X statements pushed)
- [ ] ServiceNow records updated with new content
- [ ] Local modified flags cleared on success

**US-4.2: Review Changes Before Push**
> As a Control Author, I want to review what will be pushed before confirming so that I don't accidentally push incomplete work.

**Acceptance Criteria:**
- [ ] Pre-push summary shows count of modified statements
- [ ] Can expand to see list of modified controls
- [ ] Can view diff (before/after) for each statement
- [ ] Can exclude specific statements from push
- [ ] Confirm/cancel before push executes

**US-4.3: Handle Conflicts**
> As a Control Author, I want to be warned if ServiceNow data changed since my last pull so that I don't overwrite others' work.

**Acceptance Criteria:**
- [ ] Detect if ServiceNow record was modified since pull
- [ ] Show conflict details (who changed, when)
- [ ] Options: Overwrite, Skip, View Diff
- [ ] Can resolve conflicts per-statement
- [ ] Clear explanation of consequences

**US-4.4: View Push Results**
> As a Control Author, I want to see the results of my push so that I know which statements succeeded or failed.

**Acceptance Criteria:**
- [ ] Summary: X succeeded, Y failed, Z skipped
- [ ] Details for each statement
- [ ] Error messages for failures
- [ ] Retry option for failed items
- [ ] Results persist until dismissed

### Secondary User Scenarios

**US-4.5: Push Single Statement**
> As a Control Author, I want to push a single statement immediately after editing so that I can save work incrementally.

**US-4.6: Cancel In-Progress Push**
> As a Control Author, I want to cancel a push in progress if I realize I made a mistake.

### Edge Cases and Error Scenarios

| Scenario | Expected Behavior |
|----------|-------------------|
| No modified statements | Inform user, no action needed |
| ServiceNow connection lost | Clear error, allow retry |
| Partial push failure | Report which failed, allow retry |
| Statement deleted in ServiceNow | Warn, offer to recreate or skip |
| Concurrent push (same user) | Prevent duplicate push |
| Rate limit exceeded | Queue and retry with backoff |
| Push timeout | Allow resume from last success |

---

## 3. Functional Requirements

### FR-4.1: Push Initiation
1. System SHALL provide "Push to ServiceNow" action
2. System SHALL identify all locally modified statements
3. System SHALL display pre-push summary
4. System SHALL require confirmation before push
5. System SHALL allow selective statement exclusion

### FR-4.2: Conflict Detection
1. System SHALL compare local sys_updated_on with ServiceNow
2. System SHALL fetch current ServiceNow version before push
3. System SHALL identify conflicts before attempting update
4. System SHALL present conflict resolution options
5. System SHALL record conflict resolution choice in audit log

### FR-4.3: Push Execution
1. System SHALL push statements to ServiceNow Table API
2. System SHALL handle pagination for large push sets
3. System SHALL update local sync state on success
4. System SHALL clear local modified flag on success
5. System SHALL record push timestamp and user
6. System SHALL handle API errors gracefully

### FR-4.4: Push Results
1. System SHALL report success/failure per statement
2. System SHALL provide error details for failures
3. System SHALL offer retry for failed items
4. System SHALL log all push operations to sync_log

### FR-4.5: Evidence Reference Push
1. System SHALL push modified evidence references
2. System SHALL update evidence links in ServiceNow
3. System SHALL NOT upload evidence files (metadata only)

---

## 4. User Experience Requirements

### UI Components

**Push Summary Modal:**
- Count of statements to push
- Expandable list of affected controls
- Checkbox to include/exclude items
- Conflict warnings highlighted
- "Push" and "Cancel" buttons

**Push Progress View:**
- Progress bar with percentage
- Current item being processed
- Running count (5/20 complete)
- Cancel button
- Expandable log

**Push Results Modal:**
- Success/failure summary
- Expandable details per statement
- Error messages with explanations
- "Retry Failed" button
- "Close" button

**Conflict Resolution:**
- Side-by-side diff view
- Local version vs ServiceNow version
- Action buttons: Overwrite, Skip, Pull Latest
- Apply to all option for bulk conflicts

### Interaction Patterns
- Push initiated from toolbar/header button
- Pre-push review as modal
- Progress as overlay or drawer
- Results as dismissible modal
- Conflict resolution inline in results

### Accessibility Requirements
- Progress announced to screen readers
- Conflict resolution keyboard navigable
- Error messages clearly associated
- Focus management through workflow

---

## 5. Data Requirements

### Data Model

Uses models from F2, extends sync_log:

```
Table: sync_log (push-specific fields)
├── operation: 'push'
├── statements_attempted: INTEGER
├── statements_succeeded: INTEGER
├── statements_failed: INTEGER
├── statements_skipped: INTEGER
├── conflict_count: INTEGER
├── conflict_resolution: JSONB -- Array of {statement_id, resolution}

Table: push_results (transient, per-push)
├── id: UUID (PK)
├── sync_log_id: UUID (FK sync_log)
├── statement_id: UUID (FK implementation_statements)
├── status: ENUM('success', 'failed', 'skipped', 'conflict')
├── error_code: VARCHAR(50)
├── error_message: TEXT
├── servicenow_response: JSONB
└── created_at: TIMESTAMP
```

### Data Validation
- Content matches local database
- Version numbers for optimistic locking
- ServiceNow sys_id present for update

---

## 6. Technical Constraints

### From ADR
- **Backend:** Go with retry logic for API calls
- **API:** ServiceNow Table API for updates
- **Transactions:** Per-statement updates (no batch transaction)
- **Rate limiting:** Respect ServiceNow limits

### ServiceNow API Requirements
- PUT to `/api/now/table/{table}/{sys_id}`
- Request body with updated fields
- Response includes updated sys_updated_on
- Handle 409 Conflict response

---

## 7. API/Integration Specifications

### Internal API Endpoints

**GET /api/v1/systems/{id}/push/preview**
```json
Response: {
  "data": {
    "statements_to_push": 5,
    "statements": [
      {
        "id": "uuid",
        "control_number": "AC-1",
        "has_conflict": false,
        "local_modified_at": "2026-01-27T10:30:00Z"
      }
    ],
    "conflicts": []
  }
}
```

**POST /api/v1/systems/{id}/push**
```json
Request: {
  "statement_ids": ["uuid1", "uuid2"],
  "conflict_resolution": {
    "uuid3": "overwrite"
  }
}
Response: {
  "data": {
    "push_id": "uuid",
    "status": "started"
  }
}
```

**GET /api/v1/systems/{id}/push/{push_id}/status**
```json
Response: {
  "data": {
    "status": "completed",
    "progress": 100,
    "succeeded": 4,
    "failed": 1,
    "skipped": 0,
    "results": [
      {
        "statement_id": "uuid",
        "control_number": "AC-1",
        "status": "success"
      },
      {
        "statement_id": "uuid2",
        "control_number": "AC-2",
        "status": "failed",
        "error": "Rate limit exceeded"
      }
    ]
  }
}
```

**GET /api/v1/statements/{id}/diff**
```json
Response: {
  "data": {
    "local": {
      "content_html": "<p>Local version...</p>",
      "modified_at": "2026-01-27T10:30:00Z",
      "modified_by": "user@example.com"
    },
    "servicenow": {
      "content_html": "<p>ServiceNow version...</p>",
      "modified_at": "2026-01-27T09:00:00Z",
      "modified_by": "other@example.com"
    },
    "has_conflict": true
  }
}
```

### ServiceNow API Calls
```
PUT /api/now/table/sn_compliance_control/{sys_id}
Content-Type: application/json
{
  "description": "Updated implementation statement..."
}

Response:
{
  "result": {
    "sys_id": "abc123",
    "sys_updated_on": "2026-01-27 10:35:00"
  }
}
```

---

## 8. Non-Functional Requirements

| Requirement | Target |
|-------------|--------|
| Push single statement | < 5 seconds |
| Push 50 statements | < 2 minutes |
| Conflict detection | 100% accurate |
| Push reliability | > 99% (excluding conflicts) |
| Retry success rate | > 95% on transient errors |

---

## 9. Feature Boundaries (Non-Goals)

**Not Included:**
- Batch transaction (atomic all-or-nothing)
- Automatic conflict resolution
- Push scheduling
- Push to multiple systems simultaneously
- Rollback capability after push

**Future Enhancements:**
- Automatic periodic push
- Conflict resolution policies
- Push approval workflow

---

## 10. Dependencies

| Dependency | Type | Status |
|------------|------|--------|
| F1: ServiceNow Connection | Feature | Required |
| F2: Control Package Pull | Feature | Required for data |
| F3: Statement Editor | Feature | Required for modifications |

---

## 11. Success Criteria

| Metric | Target |
|--------|--------|
| Push success rate | > 99% (excluding conflicts) |
| Conflict detection accuracy | 100% |
| Data integrity | No data loss or corruption |
| User confidence | Users trust push results |

---

## 12. Testing Requirements

### Unit Tests
- Conflict detection logic
- Push request formatting
- Error handling

### Integration Tests
- ServiceNow API mocking
- Push result processing
- Retry logic

### E2E Tests
- Full push workflow
- Conflict resolution flow
- Error handling scenarios

---

## 13. Implementation Considerations

### Complexity Assessment
**Medium-High:** Conflict detection, error handling, state management

### Recommended Approach
1. Push preview endpoint
2. Conflict detection service
3. Push execution service with retry
4. Push status tracking
5. Frontend push modal and progress
6. Conflict resolution UI
7. Results display and retry

### Potential Challenges
- ServiceNow API rate limiting
- Conflict resolution UX complexity
- Ensuring data consistency
- Handling partial failures

---

## 14. Open Questions

- [ ] How should we handle statements that were deleted in ServiceNow?
- [ ] Should we support "push and pull" (refresh after push)?
- [ ] What is the maximum number of statements to push in one operation?
- [ ] Should failed pushes block the UI or run in background?
