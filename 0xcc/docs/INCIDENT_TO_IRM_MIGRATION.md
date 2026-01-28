# Incident Table to IRM Controls Migration Guide

## Overview

The current implementation uses ServiceNow's `incident` table as a **demo/development stand-in** because the ServiceNow developer instance (`dev187038.service-now.com`) does not have the IRM (Integrated Risk Management) application installed.

**Note:** ServiceNow's GRC (Governance, Risk, and Compliance) module has been rebranded as **IRM (Integrated Risk Management)**. This document uses IRM terminology throughout. When searching ServiceNow documentation or APIs, you may find references to both "GRC" and "IRM" - they refer to the same product family.

When access to a ServiceNow instance with IRM is available, this document provides all the details needed to switch from the incident table to the actual IRM tables.

---

## Current Demo Configuration

### ServiceNow Instance

- **URL:** `https://dev187038.service-now.com/`
- **Credentials:** `admin` / `y$j6-MzkM7MY`
- **IRM Status:** NOT INSTALLED (returns 404 on IRM table queries)

### Tables Currently Used

| Purpose | Demo Table | Target IRM Table |
|---------|------------|------------------|
| Policy Statements | `incident` | `sn_compliance_policy_statement` |

---

## Files That Need Modification

### 1. Backend - ServiceNow Client

**File:** `backend/internal/infrastructure/servicenow/policy_statement.go`

**Current Code (lines 31-33):**

```go
// policyStatementTable is the ServiceNow table to query for policy statements.
// DEMO: "incident" - Change to "sn_compliance_policy_statement" for IRM
policyStatementTable = "incident"
```

**Target Code:**

```go
policyStatementTable = "sn_compliance_policy_statement"
```

---

### 2. Backend - ServiceNow Models

**File:** `backend/internal/infrastructure/servicenow/models.go`

**Current `PolicyStatementRecord` struct:**

```go
type PolicyStatementRecord struct {
    SysID            string `json:"sys_id"`
    Number           string `json:"number"`
    Name             string `json:"name"`              // IRM: populated | DEMO: empty
    ShortDescription string `json:"short_description"` // Both: populated
    Description      string `json:"description"`       // Both: populated
    State            string `json:"state"`             // IRM: "draft","active" | DEMO: "1","2","3"
    Category         string `json:"category"`          // Both: populated (different values)
    ControlFamily    string `json:"u_control_family"`  // IRM: real value | DEMO: empty
    Priority         string `json:"priority"`          // DEMO ONLY: remove for IRM
    Active           string `json:"active"`            // Both: "true"/"false" or "1"/"0"
    SysCreatedOn     string `json:"sys_created_on"`    // Both: timestamp
    SysUpdatedOn     string `json:"sys_updated_on"`    // Both: timestamp
}
```

**Changes for IRM:**

- Remove `Priority` field (incident-specific)
- Verify `u_control_family` is the correct custom field name in your IRM instance
- Add any additional IRM-specific fields needed

---

### 3. Backend - Domain Service Transformation

**File:** `backend/internal/domain/controls/service.go`

**Current `transformPolicyStatement` function has two demo fallbacks:**

```go
// DEMO FALLBACK #1: Name
name := record.Name
if name == "" {
    name = record.ShortDescription // DEMO ONLY: Remove for IRM
}

// DEMO FALLBACK #2: ControlFamily
controlFamily := record.ControlFamily
if controlFamily == "" && record.Priority != "" {
    controlFamily = "Priority " + record.Priority // DEMO ONLY: Remove for IRM
}
```

**For IRM:** Remove both fallback blocks - IRM records have proper `Name` and `ControlFamily` values.

---

### 4. Frontend - State Label Mapping

**File:** `frontend/src/features/controls/components/ControlCard.tsx`

**Current `getStateLabel` function handles incident numeric states:**

```typescript
// DEMO: Incident numeric state mapping - Remove this block for IRM
const stateNum = parseInt(state, 10);
if (!isNaN(stateNum)) {
  const incidentStateMap: Record<number, string> = {
    1: 'New',
    2: 'In Progress',
    3: 'On Hold',
    6: 'Resolved',
    7: 'Closed',
  };
  return incidentStateMap[stateNum] || `State ${state}`;
}
```

**For IRM:** Remove the numeric state mapping block. IRM uses string states like "draft", "active", "retired".

---

## Field Mapping Reference

### Incident → Policy Statement Field Mapping

| Incident Field | Policy Statement Field | Notes |
|----------------|------------------------|-------|
| `sys_id` | `sys_id` | Same - unique identifier |
| `number` | `number` | Same - human-readable ID (INC vs PS prefix) |
| `short_description` | `name` | Incidents don't have `name`, use `short_description` |
| `short_description` | `short_description` | Same field, different usage |
| `description` | `description` | Same |
| `state` (1,2,3,6,7) | `state` (draft,active,etc) | Different values |
| `category` | `category` | May have different values |
| `priority` | `u_control_family` | WORKAROUND - not equivalent |
| `active` | `active` | Same |
| `sys_created_on` | `sys_created_on` | Same |
| `sys_updated_on` | `sys_updated_on` | Same |

### IRM-Specific Fields to Add

When switching to IRM, consider adding these fields:

| Field | Description |
|-------|-------------|
| `policy` | Reference to parent policy record |
| `owner` | Assigned statement owner |
| `effective_date` | When statement becomes effective |
| `review_date` | Next review date |
| `control_objectives` | Related control objectives |
| `compliance_requirements` | Mapped compliance requirements |

---

## ServiceNow IRM Table Reference

### Primary IRM Tables

| Table | API Name | Description |
|-------|----------|-------------|
| Policy | `sn_compliance_policy` | Compliance policies |
| Policy Statement | `sn_compliance_policy_statement` | Individual policy statements |
| Control | `sn_compliance_control` | Compliance controls |
| Control Objective | `sn_compliance_control_objective` | Control objectives |
| Citation | `sn_compliance_citation` | Regulatory citations |

### Useful IRM Query Parameters

```text
# Get active policy statements
sysparm_query=active=true

# Get policy statements for a specific policy
sysparm_query=policy=<policy_sys_id>

# Get policy statements by control family
sysparm_query=u_control_family=AC

# Order by number
sysparm_orderby=number
```

---

## Migration Checklist

When IRM access becomes available:

### Pre-Migration

- [ ] Verify IRM is installed: Query `sn_compliance_policy_statement` table
- [ ] Document actual field names (may have custom prefix like `u_`)
- [ ] Document actual state values used
- [ ] Note any custom fields specific to your implementation

### Backend Changes

- [ ] Update table name constant in `policy_statement.go`
- [ ] Update `PolicyStatementRecord` struct fields in `models.go`
- [ ] Remove incident-specific fallback logic in `service.go`
- [ ] Update unit tests if any mock incident-specific behavior

### Frontend Changes

- [ ] Update `getStateLabel` with IRM state values in `ControlCard.tsx`
- [ ] Update `getStateClass` with IRM state classes in `ControlCard.tsx`
- [ ] Verify field display (Name should now be populated)

### Testing

- [ ] Test API returns data from IRM table
- [ ] Verify pagination works with IRM data volume
- [ ] Test search functionality with IRM fields
- [ ] Verify state badges display correctly

### Configuration (Future Enhancement)

Consider making the table name configurable via environment variable:

```go
// config.go
type ServiceNowConfig struct {
    InstanceURL          string
    PolicyStatementTable string // Default: "sn_compliance_policy_statement"
}

// policy_statement.go
tableName := c.config.PolicyStatementTable
if tableName == "" {
    tableName = "sn_compliance_policy_statement"
}
endpoint := fmt.Sprintf("%s/api/now/table/%s", c.config.InstanceURL, tableName)
```

---

## Quick Reference: Code Locations

| What | File | Line(s) |
|------|------|---------|
| Table name constant | `backend/internal/infrastructure/servicenow/policy_statement.go` | 31-33 |
| Record struct | `backend/internal/infrastructure/servicenow/models.go` | 108-121 |
| Name fallback | `backend/internal/domain/controls/service.go` | 131-134 |
| Priority→Family mapping | `backend/internal/domain/controls/service.go` | 141-144 |
| State labels | `frontend/src/features/controls/components/ControlCard.tsx` | 56-75 |
| State classes | `frontend/src/features/controls/components/ControlCard.tsx` | 39-51 |

---

## Notes

- The demo uses incidents because they're available on all ServiceNow instances
- Incidents have numeric states (1,2,3,6,7) while IRM likely uses string states
- The `Priority` field is only used as a visual placeholder for `ControlFamily`
- When real IRM is available, the `Name` field will be populated (no fallback needed)
- Test the IRM table access first:

```bash
curl -u admin:password "https://instance.service-now.com/api/now/table/sn_compliance_policy_statement?sysparm_limit=1"
```

---

**Document Created:** 2026-01-28
**Last Updated:** 2026-01-28
**Author:** Claude Code Assistant
