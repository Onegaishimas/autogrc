# Feature PRD: Sync Audit Trail (F5)

**Document Version:** 1.0
**Date:** 2026-01-27
**Feature ID:** F5
**Phase:** 1 (Core Sync MVP)
**Priority:** Must Have

---

## 1. Feature Overview

### Feature Name
Sync Audit Trail

### Brief Description
Comprehensive logging and display of all synchronization operations between ControlCRUD and ServiceNow GRC, providing accountability, troubleshooting, and compliance evidence.

### Problem Statement
Organizations require audit trails for compliance documentation activities. Without logging, there's no accountability for who made changes, when sync operations occurred, or why failures happened. This creates compliance risk and makes troubleshooting difficult.

### Feature Goals
1. Log all pull and push operations with timestamps and users
2. Record success/failure status and error details
3. Track changes to individual statements
4. Provide searchable audit history
5. Support compliance reporting requirements

### User Value Proposition
*"As a Compliance Manager, I can view a complete history of all sync operations so that I can demonstrate accountability and investigate any issues."*

### Connection to Project Objectives
- **Compliance requirement:** Audit trails are essential for security control documentation
- **Troubleshooting:** Helps diagnose sync failures
- **Accountability:** Shows who changed what and when

---

## 2. User Stories & Scenarios

### Primary User Stories

**US-5.1: View Sync History**
> As a Compliance Manager, I want to view the history of sync operations so that I can monitor system usage and compliance.

**Acceptance Criteria:**
- [ ] Can view list of all sync operations
- [ ] See operation type (pull/push)
- [ ] See timestamp and user
- [ ] See status (success/failure/partial)
- [ ] See summary (records processed)
- [ ] Pagination for large history

**US-5.2: View Operation Details**
> As a Compliance Manager, I want to drill into a specific sync operation so that I can see exactly what happened.

**Acceptance Criteria:**
- [ ] Can expand operation to see details
- [ ] See list of affected statements
- [ ] See success/failure per statement
- [ ] See error messages for failures
- [ ] See duration of operation

**US-5.3: Filter and Search Audit Log**
> As a Compliance Manager, I want to filter and search the audit log so that I can find specific operations or patterns.

**Acceptance Criteria:**
- [ ] Filter by date range
- [ ] Filter by operation type (pull/push)
- [ ] Filter by user
- [ ] Filter by status (success/failure)
- [ ] Filter by system
- [ ] Search by control number

**US-5.4: Export Audit Log**
> As a Compliance Manager, I want to export the audit log so that I can include it in compliance reports or share with auditors.

**Acceptance Criteria:**
- [ ] Export to CSV format
- [ ] Export to PDF format
- [ ] Include selected date range
- [ ] Include all detail fields
- [ ] Filename includes export date

### Secondary User Scenarios

**US-5.5: View Statement Change History**
> As a Control Author, I want to see the change history for a specific statement so that I can understand how it evolved.

**US-5.6: Receive Failure Notifications**
> As an Administrator, I want to be alerted when sync operations fail so that I can take corrective action.

### Edge Cases and Error Scenarios

| Scenario | Expected Behavior |
|----------|-------------------|
| Very large audit log (>10k entries) | Pagination, date range filtering |
| Export times out | Background export with notification |
| User deleted | Show "Deleted User" with ID |
| System deleted | Show "Deleted System" with ID |

---

## 3. Functional Requirements

### FR-5.1: Sync Operation Logging
1. System SHALL log every pull operation start and completion
2. System SHALL log every push operation start and completion
3. System SHALL record user ID for each operation
4. System SHALL record system ID for each operation
5. System SHALL record timestamp for start and completion
6. System SHALL record status (success, failure, partial)
7. System SHALL record count of records processed
8. System SHALL record count of errors
9. System SHALL record error messages

### FR-5.2: Statement-Level Logging
1. System SHALL log each statement affected by a sync
2. System SHALL record statement ID and control number
3. System SHALL record operation result per statement
4. System SHALL record error details for failures
5. System SHALL record previous and new content hash (for change tracking)

### FR-5.3: Audit Log Display
1. System SHALL display sync history in reverse chronological order
2. System SHALL paginate results (default 20 per page)
3. System SHALL support filtering by all logged attributes
4. System SHALL support date range selection
5. System SHALL support full-text search on error messages
6. System SHALL provide expandable detail view

### FR-5.4: Export Functionality
1. System SHALL export to CSV with all fields
2. System SHALL export to PDF with formatted layout
3. System SHALL apply current filters to export
4. System SHALL limit export to 10,000 records
5. System SHALL include export metadata (date, filters, user)

### FR-5.5: Retention
1. System SHALL retain audit logs for minimum 7 years
2. System SHALL support configurable retention period
3. System SHALL NOT automatically delete audit logs
4. System SHALL provide admin-only archive capability

---

## 4. User Experience Requirements

### UI Components

**Audit Log Page:**
```
┌─────────────────────────────────────────────────────────────┐
│ Sync Audit Trail                              [Export ▼]    │
├─────────────────────────────────────────────────────────────┤
│ Filters: [Date Range] [Type ▼] [User ▼] [Status ▼] [Search] │
├─────────────────────────────────────────────────────────────┤
│ ▶ 2026-01-27 10:35 | Push | user@ex.com | System A | ✓ 5/5  │
│ ▶ 2026-01-27 09:00 | Pull | user@ex.com | System A | ✓ 312  │
│ ▼ 2026-01-26 16:20 | Push | other@ex.com| System B | ⚠ 3/5  │
│   ├─ AC-1: Success                                          │
│   ├─ AC-2: Success                                          │
│   ├─ AC-3: Failed - Rate limit exceeded                     │
│   ├─ AC-4: Success                                          │
│   └─ AC-5: Skipped - Conflict                               │
├─────────────────────────────────────────────────────────────┤
│ Showing 1-20 of 156 | [< Prev] [1] [2] [3] ... [Next >]     │
└─────────────────────────────────────────────────────────────┘
```

**Export Modal:**
- Format selection (CSV/PDF)
- Date range confirmation
- Preview record count
- Download progress
- Download button

### Interaction Patterns
- Click row to expand details
- Keyboard navigation through log
- Filter changes refresh list
- Export opens modal with options
- Clear filters button

### Accessibility Requirements
- Table semantics for screen readers
- Expandable rows keyboard accessible
- Filter controls labeled
- Export status announced

---

## 5. Data Requirements

### Data Model

```
Table: sync_log (complete schema)
├── id: UUID (PK)
├── system_id: UUID (FK systems)
├── operation: ENUM('pull', 'push')
├── status: ENUM('started', 'completed', 'failed', 'partial')
├── records_total: INTEGER
├── records_succeeded: INTEGER
├── records_failed: INTEGER
├── records_skipped: INTEGER
├── error_message: TEXT -- Overall error if failed
├── started_at: TIMESTAMP
├── completed_at: TIMESTAMP
├── duration_ms: INTEGER
├── created_by: UUID (FK users)
├── metadata: JSONB -- Additional context
└── INDEX on (created_by, started_at)
└── INDEX on (system_id, started_at)
└── INDEX on (status)

Table: sync_log_details
├── id: UUID (PK)
├── sync_log_id: UUID (FK sync_log)
├── statement_id: UUID (FK implementation_statements)
├── control_number: VARCHAR(20) -- Denormalized for display
├── operation_result: ENUM('success', 'failed', 'skipped', 'conflict')
├── error_code: VARCHAR(50)
├── error_message: TEXT
├── content_hash_before: VARCHAR(64) -- SHA-256 of content
├── content_hash_after: VARCHAR(64)
├── servicenow_response: JSONB
└── created_at: TIMESTAMP
└── INDEX on (sync_log_id)
└── INDEX on (statement_id)

Table: statement_change_log
├── id: UUID (PK)
├── statement_id: UUID (FK implementation_statements)
├── change_type: ENUM('created', 'updated', 'deleted', 'pulled', 'pushed')
├── content_before: TEXT
├── content_after: TEXT
├── changed_by: UUID (FK users)
├── changed_at: TIMESTAMP
├── source: ENUM('local_edit', 'pull', 'push')
└── INDEX on (statement_id, changed_at)
```

### Data Validation
- Timestamps in UTC
- User IDs must reference valid users (or NULL for system operations)
- Content hashes are SHA-256 hex strings

---

## 6. Technical Constraints

### From ADR
- **Backend:** Go with structured logging (zerolog)
- **Database:** PostgreSQL with proper indexing
- **Export:** Server-side generation for large datasets
- **Retention:** 7-year minimum for compliance

### Performance Considerations
- Index on common filter columns
- Pagination required for large datasets
- Async export for large date ranges
- Consider partitioning by date for very large logs

---

## 7. API/Integration Specifications

### Internal API Endpoints

**GET /api/v1/audit/sync**
```json
Query: ?start_date=2026-01-01&end_date=2026-01-31&operation=push&status=failed&page=1&per_page=20

Response: {
  "data": [
    {
      "id": "uuid",
      "system_name": "System A",
      "operation": "push",
      "status": "partial",
      "records_total": 5,
      "records_succeeded": 3,
      "records_failed": 2,
      "started_at": "2026-01-27T10:30:00Z",
      "completed_at": "2026-01-27T10:30:45Z",
      "duration_ms": 45000,
      "user_email": "user@example.com"
    }
  ],
  "meta": {
    "total": 156,
    "page": 1,
    "per_page": 20
  }
}
```

**GET /api/v1/audit/sync/{id}**
```json
Response: {
  "data": {
    "id": "uuid",
    "system_name": "System A",
    "operation": "push",
    "status": "partial",
    "started_at": "2026-01-27T10:30:00Z",
    "completed_at": "2026-01-27T10:30:45Z",
    "details": [
      {
        "control_number": "AC-1",
        "result": "success"
      },
      {
        "control_number": "AC-2",
        "result": "failed",
        "error": "Rate limit exceeded"
      }
    ]
  }
}
```

**GET /api/v1/audit/statements/{id}/history**
```json
Response: {
  "data": [
    {
      "change_type": "pushed",
      "changed_at": "2026-01-27T10:30:00Z",
      "changed_by": "user@example.com",
      "has_content_change": true
    },
    {
      "change_type": "updated",
      "changed_at": "2026-01-27T09:00:00Z",
      "changed_by": "user@example.com",
      "has_content_change": true
    }
  ]
}
```

**POST /api/v1/audit/export**
```json
Request: {
  "format": "csv",
  "start_date": "2026-01-01",
  "end_date": "2026-01-31",
  "filters": {
    "operation": "push",
    "status": "failed"
  }
}
Response: {
  "data": {
    "export_id": "uuid",
    "status": "processing",
    "estimated_records": 500
  }
}
```

**GET /api/v1/audit/export/{id}**
```json
Response: {
  "data": {
    "status": "completed",
    "download_url": "/api/v1/audit/export/uuid/download",
    "expires_at": "2026-01-28T10:30:00Z",
    "record_count": 500,
    "file_size_bytes": 125000
  }
}
```

---

## 8. Non-Functional Requirements

| Requirement | Target |
|-------------|--------|
| Audit log query | < 2 seconds for 1000 records |
| Export generation | < 30 seconds for 10,000 records |
| Log retention | 7 years minimum |
| Audit log integrity | Immutable (no edits/deletes) |
| Search response | < 3 seconds |

---

## 9. Feature Boundaries (Non-Goals)

**Not Included:**
- Real-time audit streaming
- Audit log modification or deletion
- Integration with external SIEM systems
- Automated alerting on failures
- Audit log analytics/dashboards

**Future Enhancements:**
- SIEM integration (Splunk, etc.)
- Automated failure notifications
- Trend analysis and reporting
- Compliance report templates

---

## 10. Dependencies

| Dependency | Type | Status |
|------------|------|--------|
| F2: Control Package Pull | Feature | Logs pull operations |
| F4: Control Package Push | Feature | Logs push operations |
| PostgreSQL database | Infrastructure | Required |

---

## 11. Success Criteria

| Metric | Target |
|--------|--------|
| Log completeness | 100% of operations logged |
| Query performance | < 2 seconds for typical queries |
| Export reliability | 100% successful exports |
| Auditor satisfaction | Meets compliance requirements |

---

## 12. Testing Requirements

### Unit Tests
- Log entry creation
- Filter logic
- Export formatting

### Integration Tests
- Database queries with filters
- Pagination
- Export file generation

### E2E Tests
- Full audit log viewing workflow
- Export download
- Statement history view

---

## 13. Implementation Considerations

### Complexity Assessment
**Medium:** CRUD operations, filtering, export generation

### Recommended Approach
1. Database schema and migrations
2. Logging middleware/hooks for sync operations
3. Audit log API endpoints
4. Audit log UI with filtering
5. Export generation service
6. Statement history view

### Potential Challenges
- Efficient queries on large audit logs
- Export memory management for large datasets
- Ensuring no audit entries are lost
- Date/time zone handling

---

## 14. Open Questions

- [ ] Should we log read operations (viewing statements)?
- [ ] What compliance standards should we specifically support?
- [ ] Should exports include statement content or just metadata?
- [ ] How should we handle audit log backup/archival?
