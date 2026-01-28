# Task List: Statement Editor (F3)

**Feature ID:** F3
**Related PRD:** 003_FPRD|Statement_Editor.md
**Related TDD:** 003_FTDD|Statement_Editor.md
**Related TID:** 003_FTID|Statement_Editor.md
**Depends On:** F2 (Control Package Pull)

---

## Relevant Files

### Backend - Database
- `backend/migrations/20260127_003_add_statement_history.sql` - Edit history table
- `backend/internal/infrastructure/database/queries/statement.sql` - Statement queries (modify)

### Backend - Domain
- `backend/internal/domain/statement/models.go` - Add edit tracking fields (modify)
- `backend/internal/domain/statement/service.go` - Add save/history methods (modify)
- `backend/internal/domain/statement/service_test.go` - Additional tests (modify)
- `backend/internal/domain/statement/repository.go` - Add edit queries (modify)

### Backend - API
- `backend/internal/api/handlers/statement/handler.go` - Statement HTTP handlers
- `backend/internal/api/handlers/statement/handler_test.go` - Handler tests
- `backend/internal/api/handlers/statement/schemas.go` - Statement DTOs

### Frontend - Editor Feature
- `frontend/src/features/editor/components/StatementEditor.tsx` - Main editor container
- `frontend/src/features/editor/components/EditorPane.tsx` - TipTap editor pane
- `frontend/src/features/editor/components/ContextPane.tsx` - Control context display
- `frontend/src/features/editor/components/EditorToolbar.tsx` - Formatting toolbar
- `frontend/src/features/editor/components/SaveIndicator.tsx` - Auto-save status
- `frontend/src/features/editor/components/ChangeTracker.tsx` - Modified indicator
- `frontend/src/features/editor/extensions/index.ts` - TipTap extension exports
- `frontend/src/features/editor/extensions/tableExtension.ts` - Table support
- `frontend/src/features/editor/extensions/placeholderExtension.ts` - Placeholder text
- `frontend/src/features/editor/hooks/useEditor.ts` - TipTap editor hook
- `frontend/src/features/editor/hooks/useAutoSave.ts` - Auto-save logic
- `frontend/src/features/editor/hooks/useStatement.ts` - Statement query/mutation
- `frontend/src/features/editor/stores/editorStore.ts` - Zustand editor state
- `frontend/src/features/editor/api/statementApi.ts` - API client
- `frontend/src/features/editor/types.ts` - TypeScript types

### Frontend - UI Components
- `frontend/src/components/ui/split-pane.tsx` - Resizable split pane

### Frontend - Pages
- `frontend/src/pages/EditorPage.tsx` - Editor page route

### Frontend - Tests
- `frontend/src/features/editor/components/StatementEditor.test.tsx` - Editor tests
- `frontend/src/features/editor/components/EditorToolbar.test.tsx` - Toolbar tests
- `frontend/src/features/editor/hooks/useAutoSave.test.ts` - Auto-save tests

### Notes

- TipTap 2.x with ProseMirror extensions
- Auto-save debounced to 30 seconds after last keystroke
- Zustand for editor-specific UI state
- TanStack Query for server state
- Run frontend tests: `cd frontend && pnpm test`

---

## Tasks

- [ ] 1.0 Create Edit History Database Schema
  - [ ] 1.1 Create migration file `20260127_003_add_statement_history.sql`
  - [ ] 1.2 Create `statement_edits` table with before/after content
  - [ ] 1.3 Add edit_session_id column for grouping related edits
  - [ ] 1.4 Add edited_by and edited_at columns
  - [ ] 1.5 Create index on statement_id for history queries
  - [ ] 1.6 Create index on edit_session_id for session grouping
  - [ ] 1.7 Add modified_by column to statements table
  - [ ] 1.8 Run migration and verify schema

- [ ] 2.0 Update Statement Repository for Editing
  - [ ] 2.1 Add SaveContent query to statement.sql
  - [ ] 2.2 Implement content update preserving sync_status logic
  - [ ] 2.3 Add RecordEdit query for history tracking
  - [ ] 2.4 Add GetEditHistory query with pagination
  - [ ] 2.5 Add RevertToEdit query for undo functionality
  - [ ] 2.6 Run sqlc generate
  - [ ] 2.7 Add tests for new queries

- [ ] 3.0 Update Statement Service for Editing
  - [ ] 3.1 Add `SaveContent` method to service
  - [ ] 3.2 Implement edit session logic (group edits within 5 minutes)
  - [ ] 3.3 Record edit history asynchronously (don't block save)
  - [ ] 3.4 Add HTML content sanitization using bluemonday
  - [ ] 3.5 Add `GetEditHistory` method
  - [ ] 3.6 Add `RevertToEdit` method
  - [ ] 3.7 Update service tests with edit scenarios

- [ ] 4.0 Implement Statement API Handlers
  - [ ] 4.1 Create `backend/internal/api/handlers/statement/schemas.go`
  - [ ] 4.2 Define SaveStatementRequest with session_id
  - [ ] 4.3 Define StatementResponse with local_content fields
  - [ ] 4.4 Define EditHistoryResponse
  - [ ] 4.5 Create `backend/internal/api/handlers/statement/handler.go`
  - [ ] 4.6 Implement `GetStatement` handler (GET /api/v1/statements/{id})
  - [ ] 4.7 Implement `SaveStatement` handler (PUT /api/v1/statements/{id})
  - [ ] 4.8 Implement `GetEditHistory` handler (GET /api/v1/statements/{id}/history)
  - [ ] 4.9 Implement `RevertToEdit` handler (POST /api/v1/statements/{id}/revert/{editId})
  - [ ] 4.10 Create `handler_test.go`
  - [ ] 4.11 Add integration tests for all endpoints

- [ ] 5.0 Install Frontend Dependencies
  - [ ] 5.1 Install TipTap core packages: @tiptap/react, @tiptap/starter-kit
  - [ ] 5.2 Install TipTap extensions: table, placeholder, underline
  - [ ] 5.3 Install react-resizable-panels for split pane
  - [ ] 5.4 Install use-debounce for auto-save debouncing
  - [ ] 5.5 Verify package.json updated correctly

- [ ] 6.0 Create TipTap Editor Extensions
  - [ ] 6.1 Create `frontend/src/features/editor/extensions/index.ts`
  - [ ] 6.2 Configure StarterKit with heading levels 1-3
  - [ ] 6.3 Configure Table extension with resizable columns
  - [ ] 6.4 Configure Placeholder extension with custom text
  - [ ] 6.5 Configure Underline extension
  - [ ] 6.6 Export configured extensions array

- [ ] 7.0 Implement Editor State Store
  - [ ] 7.1 Create `frontend/src/features/editor/stores/editorStore.ts`
  - [ ] 7.2 Define EditorState interface with editing state
  - [ ] 7.3 Add statementId, isDirty, lastSavedAt, isSaving, saveError
  - [ ] 7.4 Add editSessionId for history grouping
  - [ ] 7.5 Implement setStatementId action
  - [ ] 7.6 Implement markDirty action
  - [ ] 7.7 Implement markSaved action
  - [ ] 7.8 Implement setSaving and setSaveError actions
  - [ ] 7.9 Implement newEditSession action

- [ ] 8.0 Implement Frontend API Layer
  - [ ] 8.1 Create `frontend/src/features/editor/types.ts`
  - [ ] 8.2 Define Statement interface with all fields
  - [ ] 8.3 Create `frontend/src/features/editor/api/statementApi.ts`
  - [ ] 8.4 Implement getStatement API function
  - [ ] 8.5 Implement saveStatement API function
  - [ ] 8.6 Implement getEditHistory API function
  - [ ] 8.7 Implement revertToEdit API function
  - [ ] 8.8 Create `frontend/src/features/editor/hooks/useStatement.ts`
  - [ ] 8.9 Implement useStatement query hook
  - [ ] 8.10 Implement useSaveStatement mutation hook
  - [ ] 8.11 Implement useControl hook for context pane data

- [ ] 9.0 Implement Auto-Save Hook
  - [ ] 9.1 Create `frontend/src/features/editor/hooks/useAutoSave.ts`
  - [ ] 9.2 Define AUTO_SAVE_DELAY constant (30 seconds)
  - [ ] 9.3 Track previous content with useRef
  - [ ] 9.4 Implement debounced save using useDebouncedCallback
  - [ ] 9.5 Update store state (markDirty, markSaved, setSaving)
  - [ ] 9.6 Implement retry logic for failed saves (3 attempts)
  - [ ] 9.7 Save immediately on component unmount if dirty
  - [ ] 9.8 Return saveNow function for manual save
  - [ ] 9.9 Create `useAutoSave.test.ts`
  - [ ] 9.10 Test debounce behavior
  - [ ] 9.11 Test retry logic on failure

- [ ] 10.0 Implement TipTap Editor Hook
  - [ ] 10.1 Create `frontend/src/features/editor/hooks/useEditor.ts`
  - [ ] 10.2 Use useTipTapEditor with configured extensions
  - [ ] 10.3 Set initial content from statement
  - [ ] 10.4 Configure onUpdate callback to call content handler
  - [ ] 10.5 Set editorProps with Tailwind prose classes
  - [ ] 10.6 Set minimum height for comfortable editing
  - [ ] 10.7 Return editor instance

- [ ] 11.0 Implement Context Pane Component
  - [ ] 11.1 Create `frontend/src/features/editor/components/ContextPane.tsx`
  - [ ] 11.2 Display control ID and name in header
  - [ ] 11.3 Show control family badge
  - [ ] 11.4 Display control description in Card
  - [ ] 11.5 Display implementation guidance if available
  - [ ] 11.6 Display related controls as badges
  - [ ] 11.7 Add scrollable container for overflow
  - [ ] 11.8 Style with muted background

- [ ] 12.0 Implement Editor Toolbar Component
  - [ ] 12.1 Create `frontend/src/features/editor/components/EditorToolbar.tsx`
  - [ ] 12.2 Define ToolbarButton component with active state
  - [ ] 12.3 Add Bold, Italic, Underline buttons
  - [ ] 12.4 Add Bullet List, Ordered List buttons
  - [ ] 12.5 Add Insert Table button
  - [ ] 12.6 Add Undo, Redo buttons
  - [ ] 12.7 Add tooltips with keyboard shortcuts
  - [ ] 12.8 Disable undo/redo when not available
  - [ ] 12.9 Add separators between button groups
  - [ ] 12.10 Create `EditorToolbar.test.tsx`
  - [ ] 12.11 Test button active states

- [ ] 13.0 Implement Save Indicator Component
  - [ ] 13.1 Create `frontend/src/features/editor/components/SaveIndicator.tsx`
  - [ ] 13.2 Read state from editorStore
  - [ ] 13.3 Display error state with AlertCircle icon
  - [ ] 13.4 Display saving state with Loader2 spinner
  - [ ] 13.5 Display unsaved changes with dot indicator
  - [ ] 13.6 Display last saved time with formatDistanceToNow
  - [ ] 13.7 Style with muted foreground text

- [ ] 14.0 Implement Editor Pane Component
  - [ ] 14.1 Create `frontend/src/features/editor/components/EditorPane.tsx`
  - [ ] 14.2 Use useEditor hook with statement content
  - [ ] 14.3 Use useAutoSave hook with editor content
  - [ ] 14.4 Render EditorToolbar at top
  - [ ] 14.5 Render TipTap EditorContent
  - [ ] 14.6 Render SaveIndicator in footer
  - [ ] 14.7 Add unsaved changes warning on beforeunload
  - [ ] 14.8 Implement Ctrl+S keyboard shortcut for manual save

- [ ] 15.0 Implement Split Pane Component
  - [ ] 15.1 Create `frontend/src/components/ui/split-pane.tsx`
  - [ ] 15.2 Import Panel, PanelGroup, PanelResizeHandle from react-resizable-panels
  - [ ] 15.3 Export configured split pane with default sizes
  - [ ] 15.4 Style resize handle with hover state

- [ ] 16.0 Implement Main Statement Editor Component
  - [ ] 16.1 Create `frontend/src/features/editor/components/StatementEditor.tsx`
  - [ ] 16.2 Accept statementId prop
  - [ ] 16.3 Use useStatement hook to load statement
  - [ ] 16.4 Use useControl hook to load control context
  - [ ] 16.5 Display loading skeleton while fetching
  - [ ] 16.6 Render PanelGroup with horizontal direction
  - [ ] 16.7 Render ContextPane in left Panel (35% default)
  - [ ] 16.8 Render EditorPane in right Panel (65% default)
  - [ ] 16.9 Set min/max constraints on panels
  - [ ] 16.10 Create `StatementEditor.test.tsx`
  - [ ] 16.11 Test split pane rendering
  - [ ] 16.12 Test content loading

- [ ] 17.0 Create Editor Page
  - [ ] 17.1 Create `frontend/src/pages/EditorPage.tsx`
  - [ ] 17.2 Extract statementId from route params
  - [ ] 17.3 Render StatementEditor component
  - [ ] 17.4 Add page header with back navigation
  - [ ] 17.5 Add route definition in router config

- [ ] 18.0 End-to-End Integration
  - [ ] 18.1 Navigate to editor from statement list
  - [ ] 18.2 Verify control context loads in left pane
  - [ ] 18.3 Verify statement content loads in editor
  - [ ] 18.4 Test formatting toolbar functionality
  - [ ] 18.5 Test auto-save triggers after 30 seconds
  - [ ] 18.6 Verify save indicator shows correct state
  - [ ] 18.7 Test Ctrl+S manual save
  - [ ] 18.8 Verify is_modified flag set in database
  - [ ] 18.9 Test unsaved changes warning on navigation
  - [ ] 18.10 Test edit history recording
