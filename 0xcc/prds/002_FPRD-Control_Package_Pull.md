# Feature PRD: Control Package Pull (F2)

**Document Version:** 1.0
**Date:** 2026-01-27
**Feature ID:** F2
**Phase:** 1 (Core Sync MVP)
**Priority:** Must Have

---

## 1. Feature Overview

### Feature Name
Control Package Pull

### Brief Description
Enable users to pull (import) control packages from ServiceNow GRC into ControlCRUD, including controls, implementation statements, and evidence references for a selected information system.

### Problem Statement
Control authors need to work with current data from ServiceNow GRC without manually exporting/importing files. They need a reliable way to fetch the latest control package state so they can edit implementation statements with accurate baseline data.

### Feature Goals
1. Allow users to select an information system from ServiceNow
2. Pull all controls and their baselines for the selected system
3. Import existing implementation statements
4. Import evidence reference metadata
5. Track sync state to detect changes

### User Value Proposition
*"As a Control Author, I can pull the current control package from ServiceNow with one click, so I always start editing with accurate, up-to-date data."*

### Connection to Project Objectives
- **Bidirectional sync:** This is the "pull" half of the core sync loop
- **Data accuracy:** Ensures authors work with current ServiceNow state
- **Foundation for editing:** F3 (Statement Editor) depends on pulled data

---

## 2. User Stories & Scenarios

### Primary User Stories

**US-2.1: Select Information System**
> As a Control Author, I want to select an information system from ServiceNow so that I can work on its control package.

**Acceptance Criteria:**
- [ ] Can browse list of information systems from ServiceNow
- [ ] Can search/filter systems by name
- [ ] See system metadata (name, owner, categorization level)
- [ ] Select a system to work with
- [ ] Selected system persists for the session

**US-2.2: Pull Control Package**
> As a Control Author, I want to pull the control package for my selected system so that I have current data to work with.

**Acceptance Criteria:**
- [ ] Can initiate pull with single action
- [ ] Pull retrieves all controls in the system's baseline
- [ ] Pull retrieves existing implementation statements
- [ ] Pull retrieves evidence references (metadata, not files)
- [ ] Progress indicator during pull
- [ ] Success message with summary (X controls, Y statements)
- [ ] Pull completes within 2 minutes for 300 controls

**US-2.3: View Pull Status**
> As a Control Author, I want to see when data was last pulled so that I know how current my working data is.

**Acceptance Criteria:**
- [ ] Last pull timestamp displayed
- [ ] Pull source (ServiceNow instance) displayed
- [ ] Count of controls/statements in local data
- [ ] Warning if local data is stale (>24 hours)

### Secondary User Scenarios

**US-2.4: Refresh Pull**
> As a Control Author, I want to re-pull data from ServiceNow to refresh my local copy with any changes made in ServiceNow.

**US-2.5: View Pull Differences**
> As a Control Author, I want to see what changed since my last pull so that I understand what's new from ServiceNow.

### Edge Cases and Error Scenarios

| Scenario | Expected Behavior |
|----------|-------------------|
| No systems in ServiceNow | Display "No systems found" message |
| System has no controls | Pull succeeds with 0 controls message |
| Partial pull failure | Show which controls failed, allow retry |
| Network timeout mid-pull | Resume capability or clear failure state |
| Local changes exist | Warn before overwriting, offer merge/skip |
| ServiceNow data changed | Detect conflicts, show diff |

---

## 3. Functional Requirements

### FR-2.1: System Selection
1. System SHALL query ServiceNow for list of information systems
2. System SHALL display systems with name, owner, impact level
3. System SHALL support search by system name
4. System SHALL support pagination for large system lists
5. System SHALL cache system list for 5 minutes

### FR-2.2: Control Package Pull
1. System SHALL pull all controls in the system's control baseline
2. System SHALL pull control metadata (family, number, title, description)
3. System SHALL pull implementation statements for each control
4. System SHALL pull evidence references (links, not file contents)
5. System SHALL pull control responsibility (common, hybrid, system-specific)
6. System SHALL store pulled data in local database
7. System SHALL record pull timestamp and source
8. System SHALL handle pagination for large result sets

### FR-2.3: Sync State Tracking
1. System SHALL record last pull timestamp per system
2. System SHALL store ServiceNow record sys_id for each item
3. System SHALL store ServiceNow sys_updated_on for change detection
4. System SHALL flag items modified locally since last pull
5. System SHALL detect conflicts on subsequent pulls

### FR-2.4: Data Mapping
1. System SHALL map ServiceNow control records to local schema
2. System SHALL preserve ServiceNow sys_id as external reference
3. System SHALL handle missing optional fields gracefully
4. System SHALL normalize text encoding (UTF-8)

---

## 4. User Experience Requirements

### UI Components

**System Selector:**
- Dropdown or modal with system list
- Search input with autocomplete
- System cards showing key metadata
- "Load System" button

**Pull Interface:**
- "Pull from ServiceNow" button (prominent)
- Progress bar during pull
- Expandable log showing pull progress
- Summary card after completion

**Sync Status Display:**
- Last sync timestamp
- Sync health indicator
- "View Changes" link if differences detected

### Interaction Patterns
- Pull is async with progress updates
- Cancel option for long-running pulls
- Toast notification on completion
- Modal for conflict resolution

### Accessibility Requirements
- Progress announced to screen readers
- Keyboard accessible system selection
- Clear error messages with remediation steps

---

## 5. Data Requirements

### Data Model

```
Table: systems
├── id: UUID (PK)
├── servicenow_sys_id: VARCHAR(32) UNIQUE
├── name: VARCHAR(255) NOT NULL
├── description: TEXT
├── owner: VARCHAR(255)
├── impact_level: ENUM('low', 'moderate', 'high')
├── last_pull_at: TIMESTAMP
├── last_pull_status: ENUM('success', 'partial', 'failed')
├── created_at: TIMESTAMP
└── updated_at: TIMESTAMP

Table: controls
├── id: UUID (PK)
├── system_id: UUID (FK systems)
├── servicenow_sys_id: VARCHAR(32)
├── control_family: VARCHAR(10) -- e.g., 'AC', 'AU'
├── control_number: VARCHAR(20) -- e.g., 'AC-1', 'AC-2(1)'
├── title: VARCHAR(500)
├── description: TEXT
├── responsibility: ENUM('common', 'hybrid', 'system-specific')
├── baseline_impact: ENUM('low', 'moderate', 'high')
├── servicenow_updated_at: TIMESTAMP
├── created_at: TIMESTAMP
└── updated_at: TIMESTAMP

Table: implementation_statements
├── id: UUID (PK)
├── control_id: UUID (FK controls)
├── servicenow_sys_id: VARCHAR(32)
├── content: TEXT
├── status: ENUM('draft', 'review', 'approved')
├── servicenow_updated_at: TIMESTAMP
├── local_modified_at: TIMESTAMP -- NULL if unchanged
├── created_at: TIMESTAMP
└── updated_at: TIMESTAMP

Table: evidence_references
├── id: UUID (PK)
├── statement_id: UUID (FK implementation_statements)
├── servicenow_sys_id: VARCHAR(32)
├── name: VARCHAR(255)
├── description: TEXT
├── reference_url: VARCHAR(2048)
├── evidence_type: VARCHAR(50)
├── created_at: TIMESTAMP
└── updated_at: TIMESTAMP

Table: sync_log
├── id: UUID (PK)
├── system_id: UUID (FK systems)
├── operation: ENUM('pull', 'push')
├── status: ENUM('started', 'completed', 'failed')
├── records_processed: INTEGER
├── records_failed: INTEGER
├── error_message: TEXT
├── started_at: TIMESTAMP
├── completed_at: TIMESTAMP
└── created_by: UUID (FK users)
```

### Data Validation
- ServiceNow sys_id: 32-character alphanumeric
- Control number: Matches 800-53 pattern (e.g., AC-1, AC-2(1))
- Content: Maximum 50,000 characters
- URLs: Valid URL format

---

## 6. Technical Constraints

### From ADR
- **Backend:** Go with Chi/Echo router
- **Database:** PostgreSQL with JSONB for flexible metadata
- **API:** RESTful endpoints, pagination required
- **Performance:** Pull 300 controls in < 2 minutes

### ServiceNow API Requirements
- Table API for GRC tables
- Pagination via `sysparm_offset` and `sysparm_limit`
- Query filtering via `sysparm_query`
- Field selection via `sysparm_fields`

### ServiceNow GRC Tables
- `sn_grc_profile` - System profiles
- `sn_compliance_policy` - Policies
- `sn_compliance_control` - Controls
- `sn_compliance_control_test` - Control assessments
- Related tables for evidence and statements

---

## 7. API/Integration Specifications

### Internal API Endpoints

**GET /api/v1/systems**
```json
Response: {
  "data": [
    {
      "id": "uuid",
      "servicenow_sys_id": "abc123",
      "name": "System A",
      "owner": "John Doe",
      "impact_level": "moderate",
      "last_pull_at": "2026-01-27T10:30:00Z"
    }
  ],
  "meta": { "total": 50, "page": 1, "per_page": 20 }
}
```

**POST /api/v1/systems/{id}/pull**
```json
Response: {
  "data": {
    "pull_id": "uuid",
    "status": "started",
    "started_at": "2026-01-27T10:30:00Z"
  }
}
```

**GET /api/v1/systems/{id}/pull/{pull_id}/status**
```json
Response: {
  "data": {
    "status": "completed",
    "progress": 100,
    "controls_pulled": 312,
    "statements_pulled": 287,
    "evidence_pulled": 145,
    "errors": []
  }
}
```

**GET /api/v1/systems/{id}/controls**
```json
Response: {
  "data": [
    {
      "id": "uuid",
      "control_number": "AC-1",
      "title": "Access Control Policy and Procedures",
      "responsibility": "common",
      "statement": {
        "id": "uuid",
        "content": "The organization...",
        "status": "approved",
        "local_modified": false
      }
    }
  ],
  "meta": { "total": 312, "page": 1, "per_page": 50 }
}
```

### ServiceNow API Calls
```
GET /api/now/table/sn_grc_profile
  ?sysparm_fields=sys_id,name,owner,description
  &sysparm_limit=100

GET /api/now/table/sn_compliance_control
  ?sysparm_query=profile={sys_id}
  &sysparm_fields=sys_id,number,name,description,sys_updated_on
  &sysparm_limit=100
  &sysparm_offset=0
```

---

## 8. Non-Functional Requirements

| Requirement | Target |
|-------------|--------|
| Pull 300 controls | < 2 minutes |
| System list load | < 3 seconds |
| Pull progress updates | Every 2 seconds |
| Concurrent pull operations | 1 per user |
| Data freshness indicator | Show if > 24 hours old |

---

## 9. Feature Boundaries (Non-Goals)

**Not Included:**
- Real-time sync (webhook-based updates)
- Selective control pull (all or nothing per system)
- Pull scheduling or automation
- Evidence file download (metadata only)
- Control assessment/test data pull

**Future Enhancements:**
- Incremental pull (only changed items)
- Background sync with notifications
- Selective control family pull

---

## 10. Dependencies

| Dependency | Type | Status |
|------------|------|--------|
| F1: ServiceNow Connection | Feature | Required before pull |
| PostgreSQL database | Infrastructure | Required |
| ServiceNow GRC tables | External | Must have data |

---

## 11. Success Criteria

| Metric | Target |
|--------|--------|
| Pull success rate | > 99% |
| Pull time (300 controls) | < 2 minutes |
| Data accuracy | 100% match with ServiceNow |
| User can start editing | < 3 minutes from system select |

---

## 12. Testing Requirements

### Unit Tests
- Data mapping logic
- Pagination handling
- Conflict detection

### Integration Tests
- ServiceNow API mocking
- Database persistence
- Pull status tracking

### E2E Tests
- Full pull workflow
- Error handling scenarios

---

## 13. Implementation Considerations

### Complexity Assessment
**Medium:** Multiple API calls, pagination, data transformation

### Recommended Approach
1. ServiceNow API client methods for each table
2. Database schema and migrations
3. Pull orchestration service
4. Status tracking and progress updates
5. Frontend system selector and pull UI

### Potential Challenges
- ServiceNow API rate limits
- Large dataset pagination
- Timeout handling for slow pulls
- Mapping ServiceNow schema variations

---

## 14. Open Questions

- [ ] What is the exact ServiceNow GRC table schema for implementation statements?
- [ ] How are inherited controls represented in ServiceNow?
- [ ] What is the maximum number of controls per system we should support?
