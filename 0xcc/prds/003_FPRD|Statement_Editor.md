# Feature PRD: Statement Editor (F3)

**Document Version:** 1.0
**Date:** 2026-01-27
**Feature ID:** F3
**Phase:** 1 (Core Sync MVP)
**Priority:** Must Have

---

## 1. Feature Overview

### Feature Name
Statement Editor

### Brief Description
A focused, rich text editing environment for authoring and editing NIST 800-53 control implementation statements, with side-by-side view of control requirements.

### Problem Statement
Control authors currently edit implementation statements in ServiceNow's constrained interface, which lacks context, formatting tools, and efficient navigation. They need a purpose-built editing environment that surfaces relevant information alongside the editor.

### Feature Goals
1. Provide rich text editing with formatting support
2. Display control requirements alongside the editor
3. Enable efficient navigation between controls
4. Auto-save work to prevent data loss
5. Track changes for eventual push to ServiceNow

### User Value Proposition
*"As a Control Author, I can edit implementation statements in a focused environment with control requirements visible, so I can write complete, accurate statements efficiently."*

### Connection to Project Objectives
- **Central authoring experience:** This is the primary work area for users
- **Quality improvement:** Better tools lead to better documentation
- **Efficiency:** Reduces time spent per statement

---

## 2. User Stories & Scenarios

### Primary User Stories

**US-3.1: Edit Implementation Statement**
> As a Control Author, I want to edit an implementation statement using a rich text editor so that I can format content professionally.

**Acceptance Criteria:**
- [ ] Can edit statement content in WYSIWYG editor
- [ ] Supports bold, italic, underline formatting
- [ ] Supports bullet and numbered lists
- [ ] Supports headings (H1-H3)
- [ ] Supports tables for structured content
- [ ] Content auto-saves every 30 seconds
- [ ] Manual save option available

**US-3.2: View Control Requirements**
> As a Control Author, I want to see the control requirements while editing so that I can ensure my statement addresses all requirements.

**Acceptance Criteria:**
- [ ] Control title and number visible
- [ ] Control description/requirement text displayed
- [ ] Control enhancements listed (if applicable)
- [ ] Organization-defined parameters highlighted
- [ ] Panel resizable or collapsible

**US-3.3: Navigate Between Controls**
> As a Control Author, I want to quickly navigate between controls so that I can work through multiple statements efficiently.

**Acceptance Criteria:**
- [ ] Control list/browser visible
- [ ] Can filter by control family (AC, AU, etc.)
- [ ] Can filter by status (draft, review, approved)
- [ ] Can search by control number or keyword
- [ ] Current control highlighted in list
- [ ] Keyboard shortcuts for next/previous

**US-3.4: Track Edit Status**
> As a Control Author, I want to see which statements I've modified so that I know what needs to be pushed to ServiceNow.

**Acceptance Criteria:**
- [ ] Modified indicator on edited statements
- [ ] Last edited timestamp visible
- [ ] "Unsaved changes" warning before navigation
- [ ] Changed fields highlighted (if applicable)

### Secondary User Scenarios

**US-3.5: View Statement History**
> As a Control Author, I want to see previous versions of a statement so that I can review or revert changes.

**US-3.6: Copy Content Between Statements**
> As a Control Author, I want to copy content from one statement to another so that I can reuse similar implementation descriptions.

### Edge Cases and Error Scenarios

| Scenario | Expected Behavior |
|----------|-------------------|
| Network disconnect during edit | Local save continues, warn on reconnect |
| Browser closed with unsaved changes | Restore from auto-save on return |
| Concurrent edit (same statement) | Warn on save, offer merge/overwrite |
| Very long statement (>50k chars) | Accept but warn about length |
| Paste from Word/external source | Clean formatting, preserve structure |

---

## 3. Functional Requirements

### FR-3.1: Rich Text Editor
1. System SHALL provide WYSIWYG editing interface
2. System SHALL support text formatting (bold, italic, underline, strikethrough)
3. System SHALL support paragraph styles (normal, heading 1-3)
4. System SHALL support lists (bullet, numbered, nested)
5. System SHALL support tables (insert, resize, add/remove rows/columns)
6. System SHALL support undo/redo (minimum 50 levels)
7. System SHALL handle paste from external sources (clean HTML)
8. System SHALL NOT support images (text-only for statements)

### FR-3.2: Auto-Save
1. System SHALL auto-save content every 30 seconds
2. System SHALL auto-save on editor blur (focus loss)
3. System SHALL indicate save status (saving, saved, error)
4. System SHALL store auto-save in local database
5. System SHALL recover unsaved content on session restore

### FR-3.3: Control Context Panel
1. System SHALL display control number and title
2. System SHALL display control description/requirement
3. System SHALL display related enhancements
4. System SHALL highlight organization-defined parameters (ODPs)
5. System SHALL display control responsibility type
6. System SHALL be resizable (drag to resize)
7. System SHALL be collapsible (toggle visibility)

### FR-3.4: Control Navigation
1. System SHALL display list of controls for current system
2. System SHALL support filtering by control family
3. System SHALL support filtering by statement status
4. System SHALL support search by control number or text
5. System SHALL indicate modification status per control
6. System SHALL support keyboard navigation (↑↓ arrows)
7. System SHALL support keyboard shortcuts (Ctrl+S save, Ctrl+← → nav)

### FR-3.5: Change Tracking
1. System SHALL mark statements as modified when changed
2. System SHALL record modification timestamp
3. System SHALL record modifying user
4. System SHALL preserve original content for diff comparison
5. System SHALL clear modified flag after successful push

---

## 4. User Experience Requirements

### UI Layout

**Split-Pane Design:**
```
┌─────────────────────────────────────────────────────────────┐
│ Header: System Name | Control Count | Sync Status           │
├───────────┬─────────────────────────┬───────────────────────┤
│           │                         │                       │
│ Control   │                         │  Control              │
│ Navigator │    Statement Editor     │  Requirements         │
│           │                         │  Panel                │
│ - AC-1 ●  │  [Rich Text Editor]     │                       │
│ - AC-2    │                         │  Title: AC-1          │
│ - AC-3 ●  │                         │  Description: ...     │
│ - ...     │                         │  ODPs: [param1]       │
│           │                         │                       │
├───────────┴─────────────────────────┴───────────────────────┤
│ Footer: Auto-saved 30s ago | Modified: 5 | Ready to push    │
└─────────────────────────────────────────────────────────────┘
```

### UI Components

**Control Navigator (Left Panel):**
- Scrollable list of controls
- Family grouping (collapsible)
- Status icons (●= modified, ✓= approved)
- Search input at top
- Filter dropdowns

**Statement Editor (Center):**
- TipTap-based rich text editor
- Formatting toolbar (persistent or floating)
- Character/word count
- Save status indicator

**Requirements Panel (Right):**
- Control metadata card
- Scrollable requirement text
- ODP highlighting with distinct style
- Collapse toggle

### Interaction Patterns
- Click control to load in editor
- Keyboard shortcuts for power users
- Drag panel dividers to resize
- Auto-save with visual confirmation
- Warn before navigating with unsaved changes

### Accessibility Requirements
- Editor fully keyboard accessible
- Formatting toolbar keyboard operable
- Screen reader announcements for save status
- High contrast mode support
- Focus management on control switch

---

## 5. Data Requirements

### Data Model

Extends models from F2 (Control Package Pull):

```
Table: implementation_statements (additions)
├── content_html: TEXT -- Rich text content
├── content_plain: TEXT -- Plain text for search
├── word_count: INTEGER
├── local_modified_at: TIMESTAMP
├── local_modified_by: UUID (FK users)
├── original_content_html: TEXT -- Snapshot at last pull
└── version: INTEGER -- Optimistic locking

Table: statement_autosave
├── id: UUID (PK)
├── statement_id: UUID (FK implementation_statements)
├── user_id: UUID (FK users)
├── content_html: TEXT
├── saved_at: TIMESTAMP
└── UNIQUE(statement_id, user_id)
```

### Data Validation
- Content: Maximum 50,000 characters
- HTML: Sanitized (allowlist-based)
- No embedded scripts or external resources

---

## 6. Technical Constraints

### From ADR
- **Frontend:** React with TipTap editor
- **State:** TanStack Query for server state, Zustand for editor state
- **API:** RESTful endpoints for save operations
- **Testing:** Vitest + RTL for component testing

### Editor Requirements
- TipTap 2.x for rich text editing
- ProseMirror extensions for table support
- HTML sanitization before save
- Markdown export capability (future)

---

## 7. API/Integration Specifications

### Internal API Endpoints

**GET /api/v1/controls/{id}/statement**
```json
Response: {
  "data": {
    "id": "uuid",
    "control_id": "uuid",
    "content_html": "<p>The organization...</p>",
    "status": "draft",
    "local_modified": true,
    "local_modified_at": "2026-01-27T10:30:00Z",
    "version": 3
  }
}
```

**PUT /api/v1/statements/{id}**
```json
Request: {
  "content_html": "<p>Updated content...</p>",
  "version": 3
}
Response: {
  "data": {
    "id": "uuid",
    "version": 4,
    "updated_at": "2026-01-27T10:35:00Z"
  }
}
```

**POST /api/v1/statements/{id}/autosave**
```json
Request: {
  "content_html": "<p>Work in progress...</p>"
}
Response: {
  "data": {
    "saved_at": "2026-01-27T10:35:30Z"
  }
}
```

**GET /api/v1/controls/{id}**
```json
Response: {
  "data": {
    "id": "uuid",
    "control_number": "AC-1",
    "title": "Access Control Policy and Procedures",
    "description": "The organization: a. Develops...",
    "responsibility": "common",
    "enhancements": [],
    "odps": [
      { "param": "[organization-defined frequency]", "value": "annually" }
    ]
  }
}
```

---

## 8. Non-Functional Requirements

| Requirement | Target |
|-------------|--------|
| Editor load time | < 1 second |
| Auto-save latency | < 500ms |
| Undo/redo levels | 50 minimum |
| Max content size | 50,000 characters |
| Keyboard shortcut response | < 100ms |

---

## 9. Feature Boundaries (Non-Goals)

**Not Included:**
- AI-assisted drafting (Phase 2)
- Collaborative editing (multiple users same statement)
- Image or file embedding
- Markdown editing mode
- Inline commenting/review

**Future Enhancements:**
- AI draft generation (F6)
- AI refinement suggestions (F7)
- Completeness validation (F8)
- Template application (F10)

---

## 10. Dependencies

| Dependency | Type | Status |
|------------|------|--------|
| F2: Control Package Pull | Feature | Required for data |
| TipTap editor library | Package | To be installed |
| Control data in database | Data | From F2 pull |

---

## 11. Success Criteria

| Metric | Target |
|--------|--------|
| Editor stability | No data loss |
| Save reliability | 100% success rate |
| User productivity | < 30 min per statement (baseline) |
| User satisfaction | Editor rated 4+/5 |

---

## 12. Testing Requirements

### Unit Tests
- Editor content operations
- Auto-save logic
- Navigation state management

### Integration Tests
- Save/load roundtrip
- Auto-save persistence
- Version conflict handling

### E2E Tests
- Full editing workflow
- Navigation between controls
- Paste handling

---

## 13. Implementation Considerations

### Complexity Assessment
**Medium-High:** Rich text editing, state management, keyboard handling

### Recommended Approach
1. TipTap editor setup with extensions
2. Editor component with toolbar
3. Control navigator component
4. Requirements panel component
5. Layout integration with resizable panels
6. Auto-save implementation
7. Keyboard shortcut system

### Potential Challenges
- TipTap configuration complexity
- HTML sanitization edge cases
- Managing editor state with external data
- Cross-browser paste handling

---

## 14. Open Questions

- [ ] Should we support offline editing (local-first)?
- [ ] What HTML elements are allowed in statements?
- [ ] Should tables have a maximum size?
- [ ] How should we handle very long control descriptions in the panel?
