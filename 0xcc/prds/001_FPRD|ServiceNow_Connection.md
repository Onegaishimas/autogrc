# Feature PRD: ServiceNow GRC Connection (F1)

**Document Version:** 1.0
**Date:** 2026-01-27
**Feature ID:** F1
**Phase:** 1 (Core Sync MVP)
**Priority:** Must Have

---

## 1. Feature Overview

### Feature Name
ServiceNow GRC Connection

### Brief Description
Establish and manage the connection between ControlCRUD and a ServiceNow GRC instance, enabling secure API communication for subsequent pull/push operations.

### Problem Statement
Users need a reliable, secure connection to their ServiceNow GRC instance before they can synchronize control implementation statements. Without proper connection management, users cannot verify connectivity, troubleshoot issues, or ensure their credentials are valid.

### Feature Goals
1. Enable administrators to configure ServiceNow instance connection parameters
2. Provide connection testing to validate credentials and permissions
3. Display connection health status throughout the application
4. Securely store and manage ServiceNow credentials

### User Value Proposition
*"As an administrator, I can configure and verify the ServiceNow connection once, giving all users confidence that sync operations will work reliably."*

### Connection to Project Objectives
- **Foundation for F2-F4:** All sync features depend on this connection
- **Reduces friction:** Users don't re-enter credentials for each operation
- **Improves reliability:** Proactive connection health monitoring prevents sync failures

---

## 2. User Stories & Scenarios

### Primary User Stories

**US-1.1: Configure ServiceNow Connection**
> As an Administrator, I want to configure the ServiceNow instance URL and credentials so that ControlCRUD can communicate with our GRC system.

**Acceptance Criteria:**
- [ ] Can enter ServiceNow instance URL (e.g., `https://dev12345.service-now.com`)
- [ ] Can enter authentication credentials (username/password or OAuth)
- [ ] URL is validated for proper format
- [ ] Credentials are stored securely (encrypted at rest)
- [ ] Configuration persists across application restarts

**US-1.2: Test Connection**
> As an Administrator, I want to test the ServiceNow connection so that I can verify the configuration is correct before users attempt sync operations.

**Acceptance Criteria:**
- [ ] Can initiate a connection test with one click
- [ ] Test verifies API accessibility (can reach ServiceNow)
- [ ] Test verifies authentication (credentials accepted)
- [ ] Test verifies permissions (can read GRC tables)
- [ ] Clear success/failure message with details
- [ ] Test completes within 10 seconds

**US-1.3: View Connection Status**
> As a User, I want to see the current connection status so that I know if sync operations will work.

**Acceptance Criteria:**
- [ ] Connection status visible in application header/sidebar
- [ ] Status indicators: Connected, Disconnected, Error, Unknown
- [ ] Last successful connection timestamp displayed
- [ ] Click status to see detailed connection information

### Secondary User Scenarios

**US-1.4: Update Connection Configuration**
> As an Administrator, I want to update the ServiceNow connection settings when our instance URL or credentials change.

**US-1.5: Connection Error Recovery**
> As a User, I want to understand why the connection failed so that I can take corrective action or report the issue.

### Edge Cases and Error Scenarios

| Scenario | Expected Behavior |
|----------|-------------------|
| Invalid URL format | Validation error before save |
| Incorrect credentials | Clear error message with 401 details |
| ServiceNow instance unreachable | Timeout error with retry suggestion |
| Insufficient permissions | Error listing required GRC table permissions |
| Rate limit exceeded | Inform user, suggest retry delay |
| SSL/TLS certificate issues | Clear error about certificate validation |

---

## 3. Functional Requirements

### FR-1.1: Connection Configuration
1. System SHALL accept ServiceNow instance URL input
2. System SHALL validate URL format (HTTPS required, valid domain)
3. System SHALL accept authentication method selection (Basic Auth or OAuth 2.0)
4. System SHALL accept username and password for Basic Auth
5. System SHALL accept OAuth client credentials for OAuth 2.0
6. System SHALL encrypt credentials before storage (AES-256)
7. System SHALL persist configuration in database

### FR-1.2: Connection Testing
1. System SHALL provide a "Test Connection" action
2. System SHALL attempt API call to ServiceNow Table API
3. System SHALL verify read access to `sn_compliance_policy` table
4. System SHALL verify read access to `sn_compliance_control` table
5. System SHALL return success with API version and instance info
6. System SHALL return failure with specific error code and message
7. System SHALL timeout after 10 seconds

### FR-1.3: Connection Status
1. System SHALL maintain current connection status in application state
2. System SHALL display status indicator (icon + text)
3. System SHALL show last successful connection timestamp
4. System SHALL refresh status on application load
5. System SHALL update status after any sync operation

### FR-1.4: Credential Management
1. System SHALL NOT log credentials in plain text
2. System SHALL NOT expose credentials in API responses
3. System SHALL mask password field in UI after entry
4. System SHALL support credential rotation without data loss

---

## 4. User Experience Requirements

### UI Components

**Connection Settings Page:**
- Form with ServiceNow URL input
- Authentication method selector (tabs or radio)
- Credential input fields (masked password)
- "Test Connection" button with loading state
- "Save Configuration" button
- Connection test result display area

**Status Indicator (Global):**
- Small icon in header/sidebar
- Color coded: Green (connected), Yellow (warning), Red (error), Gray (unknown)
- Tooltip with last check time
- Click to expand details

### Interaction Patterns
- Form validation on blur and submit
- Async test with loading spinner
- Toast notification for save success/failure
- Modal confirmation for credential changes

### Accessibility Requirements
- Form labels properly associated with inputs
- Error messages announced to screen readers
- Keyboard navigation for all actions
- Color not sole indicator of status (icon + text)

---

## 5. Data Requirements

### Data Model

```
Table: servicenow_connections
├── id: UUID (PK)
├── instance_url: VARCHAR(255) NOT NULL
├── auth_method: ENUM('basic', 'oauth') NOT NULL
├── username: VARCHAR(255) -- encrypted
├── password_encrypted: BYTEA -- AES-256 encrypted
├── oauth_client_id: VARCHAR(255) -- encrypted
├── oauth_client_secret_encrypted: BYTEA
├── is_active: BOOLEAN DEFAULT true
├── last_test_at: TIMESTAMP
├── last_test_status: ENUM('success', 'failure', 'pending')
├── last_test_message: TEXT
├── created_at: TIMESTAMP
├── updated_at: TIMESTAMP
├── created_by: UUID (FK users)
└── updated_by: UUID (FK users)
```

### Data Validation
- URL: Valid HTTPS URL, ServiceNow domain pattern
- Username: Non-empty, max 255 characters
- Password: Non-empty (not validated for complexity)
- OAuth credentials: Valid UUID format for client_id

---

## 6. Technical Constraints

### From ADR
- **Backend:** Go with Chi/Echo router
- **Database:** PostgreSQL with encrypted credential storage
- **Auth:** SSO required; only admins can configure connection
- **API:** RESTful endpoints under `/api/v1/`

### ServiceNow API Requirements
- Table API access required
- Minimum required scopes: `read` on GRC tables
- Rate limits: Respect ServiceNow API rate limiting
- Authentication: Basic Auth or OAuth 2.0 supported

---

## 7. API/Integration Specifications

### Internal API Endpoints

**GET /api/v1/connection/status**
```json
Response: {
  "data": {
    "status": "connected",
    "instance_url": "https://dev12345.service-now.com",
    "last_test_at": "2026-01-27T10:30:00Z",
    "last_test_status": "success"
  }
}
```

**POST /api/v1/connection/config**
```json
Request: {
  "instance_url": "https://dev12345.service-now.com",
  "auth_method": "basic",
  "username": "admin",
  "password": "********"
}
Response: {
  "data": {
    "id": "uuid",
    "instance_url": "https://dev12345.service-now.com",
    "auth_method": "basic",
    "created_at": "2026-01-27T10:30:00Z"
  }
}
```

**POST /api/v1/connection/test**
```json
Response (Success): {
  "data": {
    "status": "success",
    "message": "Connection successful",
    "instance_info": {
      "version": "Tokyo",
      "build": "glide-tokyo-12-2022"
    },
    "permissions": {
      "sn_compliance_policy": "read",
      "sn_compliance_control": "read"
    }
  }
}
Response (Failure): {
  "error": {
    "code": "CONNECTION_FAILED",
    "message": "Authentication failed: Invalid credentials",
    "details": { "http_status": 401 }
  }
}
```

### ServiceNow API Integration
- Use Table API: `/api/now/table/{tableName}`
- Test endpoint: `GET /api/now/table/sys_properties?sysparm_limit=1`
- Required headers: `Authorization`, `Content-Type: application/json`

---

## 8. Non-Functional Requirements

| Requirement | Target |
|-------------|--------|
| Connection test response time | < 10 seconds |
| Credential encryption | AES-256-GCM |
| Configuration save time | < 1 second |
| Status refresh interval | On-demand + on app load |

---

## 9. Feature Boundaries (Non-Goals)

**Not Included:**
- Multiple ServiceNow instance support (single instance only for MVP)
- ServiceNow user provisioning or management
- Automatic credential rotation
- Connection pooling or load balancing
- Proxy configuration for ServiceNow access

**Future Enhancements:**
- Multi-instance support for enterprise deployments
- Certificate-based authentication
- Connection health monitoring with alerts

---

## 10. Dependencies

| Dependency | Type | Status |
|------------|------|--------|
| PostgreSQL database | Infrastructure | Required |
| Enterprise SSO | Authentication | Required for admin access |
| ServiceNow developer instance | External | User-provided |

---

## 11. Success Criteria

| Metric | Target |
|--------|--------|
| Connection test accuracy | 100% (correct pass/fail) |
| Test completion time | < 10 seconds |
| Credential security | No plain-text exposure |
| Admin satisfaction | Can configure in < 5 minutes |

---

## 12. Testing Requirements

### Unit Tests
- URL validation logic
- Credential encryption/decryption
- Status state management

### Integration Tests
- API endpoint responses
- Database persistence
- ServiceNow API mocking

### E2E Tests
- Full configuration flow
- Connection test with mock ServiceNow

---

## 13. Implementation Considerations

### Complexity Assessment
**Low-Medium:** Standard CRUD with external API integration

### Recommended Approach
1. Database schema and migrations first
2. Backend API endpoints with mock responses
3. ServiceNow client implementation
4. Frontend settings page
5. Integration testing with ServiceNow dev instance

### Potential Challenges
- ServiceNow API versioning differences
- OAuth token refresh handling
- Secure credential storage patterns

---

## 14. Open Questions

- [ ] Should we support multiple auth methods simultaneously?
- [ ] What is the session timeout for ServiceNow OAuth tokens?
- [ ] Are there specific ServiceNow API versions we must support?
