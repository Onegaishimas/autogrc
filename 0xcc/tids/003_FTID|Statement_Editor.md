# Technical Implementation Document: Statement Editor (F3)

**Document Version:** 1.0
**Date:** 2026-01-27
**Feature ID:** F3
**Related PRD:** 003_FPRD|Statement_Editor.md
**Related TDD:** 003_FTDD|Statement_Editor.md

---

## 1. Implementation Overview

### Summary
This TID provides implementation guidance for the TipTap-based rich text editor for control implementation statements, including the split-pane layout, auto-save functionality, and change tracking.

### Key Implementation Principles
- **TipTap Foundation:** Use TipTap 2.x with ProseMirror extensions
- **Split-Pane UX:** Control context (read-only) on left, editor on right
- **Optimistic Updates:** Save locally immediately, sync to server async
- **Debounced Auto-Save:** 30-second debounce after last keystroke

### Integration Points
- **F2 Dependency:** Reads controls/statements from pull feature
- **F4 Forward:** Provides modified statements for push feature
- **Database:** PostgreSQL for statement content and edit history
- **State:** Zustand for editor state, TanStack Query for server sync

---

## 2. File Structure and Organization

### Backend Files to Create/Modify

```
backend/
├── internal/
│   ├── api/
│   │   └── handlers/
│   │       └── statement/
│   │           ├── handler.go            # HTTP handlers (create)
│   │           ├── handler_test.go       # Handler tests (create)
│   │           └── schemas.go            # Request/response DTOs (create)
│   ├── domain/
│   │   └── statement/
│   │       ├── models.go                 # Add edit tracking fields (modify)
│   │       ├── service.go                # Add save/history methods (modify)
│   │       └── repository.go             # Add edit queries (modify)
│   └── infrastructure/
│       └── database/
│           └── queries/
│               └── statement.sql         # Add edit queries (modify)
└── migrations/
    └── 20260127_003_add_statement_history.sql  # Edit history table (create)
```

### Frontend Files to Create

```
frontend/src/
├── features/
│   └── editor/
│       ├── components/
│       │   ├── StatementEditor.tsx       # Main editor container (create)
│       │   ├── EditorPane.tsx            # TipTap editor pane (create)
│       │   ├── ContextPane.tsx           # Control context display (create)
│       │   ├── EditorToolbar.tsx         # Formatting toolbar (create)
│       │   ├── SaveIndicator.tsx         # Auto-save status (create)
│       │   └── ChangeTracker.tsx         # Modified indicator (create)
│       ├── extensions/
│       │   ├── index.ts                  # Extension exports (create)
│       │   ├── tableExtension.ts         # Table support (create)
│       │   └── placeholderExtension.ts   # Placeholder text (create)
│       ├── hooks/
│       │   ├── useEditor.ts              # TipTap editor hook (create)
│       │   ├── useAutoSave.ts            # Auto-save logic (create)
│       │   └── useStatement.ts           # Statement query/mutation (create)
│       ├── stores/
│       │   └── editorStore.ts            # Zustand editor state (create)
│       ├── api/
│       │   └── statementApi.ts           # API client (create)
│       └── types.ts                      # TypeScript types (create)
├── components/
│   └── ui/
│       └── split-pane.tsx                # Resizable split pane (create)
└── pages/
    └── EditorPage.tsx                    # Editor page route (create)
```

### Package Dependencies

```json
// package.json additions
{
  "dependencies": {
    "@tiptap/react": "^2.2.0",
    "@tiptap/starter-kit": "^2.2.0",
    "@tiptap/extension-table": "^2.2.0",
    "@tiptap/extension-table-row": "^2.2.0",
    "@tiptap/extension-table-cell": "^2.2.0",
    "@tiptap/extension-table-header": "^2.2.0",
    "@tiptap/extension-placeholder": "^2.2.0",
    "@tiptap/extension-underline": "^2.2.0",
    "react-resizable-panels": "^2.0.0"
  }
}
```

---

## 3. Component Implementation Hints

### TipTap Editor Setup

```typescript
// features/editor/hooks/useEditor.ts
import { useEditor as useTipTapEditor } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import Table from '@tiptap/extension-table';
import TableRow from '@tiptap/extension-table-row';
import TableCell from '@tiptap/extension-table-cell';
import TableHeader from '@tiptap/extension-table-header';
import Placeholder from '@tiptap/extension-placeholder';
import Underline from '@tiptap/extension-underline';

export function useStatementEditor(content: string, onUpdate: (html: string) => void) {
  const editor = useTipTapEditor({
    extensions: [
      StarterKit.configure({
        heading: { levels: [1, 2, 3] },
      }),
      Underline,
      Table.configure({ resizable: true }),
      TableRow,
      TableCell,
      TableHeader,
      Placeholder.configure({
        placeholder: 'Enter implementation statement...',
      }),
    ],
    content,
    onUpdate: ({ editor }) => {
      onUpdate(editor.getHTML());
    },
    editorProps: {
      attributes: {
        class: 'prose prose-sm max-w-none focus:outline-none min-h-[300px] p-4',
      },
    },
  });

  return editor;
}
```

**Key Hints:**
- Use `StarterKit` for basic formatting (bold, italic, lists, etc.)
- Configure heading levels to match organizational standards
- Table extension required for tabular implementation details
- Set minimum height for comfortable editing
- Use Tailwind prose classes for consistent typography

### Split-Pane Layout

```typescript
// features/editor/components/StatementEditor.tsx
import { Panel, PanelGroup, PanelResizeHandle } from 'react-resizable-panels';

export function StatementEditor({ statementId }: { statementId: string }) {
  const { data: statement, isLoading } = useStatement(statementId);
  const { data: control } = useControl(statement?.controlId);

  if (isLoading) return <EditorSkeleton />;

  return (
    <PanelGroup direction="horizontal" className="h-full">
      <Panel defaultSize={35} minSize={25} maxSize={50}>
        <ContextPane control={control} />
      </Panel>

      <PanelResizeHandle className="w-1 bg-border hover:bg-primary/20 transition-colors" />

      <Panel defaultSize={65} minSize={50}>
        <EditorPane statement={statement} />
      </Panel>
    </PanelGroup>
  );
}
```

**Key Hints:**
- Default split: 35% context, 65% editor
- Min/max constraints prevent unusable pane sizes
- Styled resize handle for clear affordance
- Full height layout (`h-full` with parent constraints)

### Context Pane Component

```typescript
// features/editor/components/ContextPane.tsx
export function ContextPane({ control }: { control: Control | undefined }) {
  if (!control) return null;

  return (
    <div className="h-full overflow-y-auto p-4 bg-muted/30">
      <div className="space-y-4">
        {/* Control Header */}
        <div>
          <Badge variant="outline">{control.controlId}</Badge>
          <h2 className="text-lg font-semibold mt-1">{control.controlName}</h2>
          <p className="text-sm text-muted-foreground">{control.controlFamily}</p>
        </div>

        {/* Control Description */}
        <Card>
          <CardHeader className="py-3">
            <CardTitle className="text-sm">Control Description</CardTitle>
          </CardHeader>
          <CardContent className="text-sm">
            {control.description}
          </CardContent>
        </Card>

        {/* Implementation Guidance */}
        {control.guidance && (
          <Card>
            <CardHeader className="py-3">
              <CardTitle className="text-sm">Implementation Guidance</CardTitle>
            </CardHeader>
            <CardContent className="text-sm">
              {control.guidance}
            </CardContent>
          </Card>
        )}

        {/* Related Controls */}
        {control.relatedControls?.length > 0 && (
          <Card>
            <CardHeader className="py-3">
              <CardTitle className="text-sm">Related Controls</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex flex-wrap gap-1">
                {control.relatedControls.map((id) => (
                  <Badge key={id} variant="secondary" className="text-xs">
                    {id}
                  </Badge>
                ))}
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
```

---

## 4. Database Implementation Approach

### Edit History Schema

```sql
-- migrations/20260127_003_add_statement_history.sql

-- Edit history for undo/audit
CREATE TABLE statement_edits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    statement_id UUID NOT NULL REFERENCES statements(id) ON DELETE CASCADE,
    content_before TEXT NOT NULL,
    content_after TEXT NOT NULL,
    edited_by UUID REFERENCES users(id),
    edited_at TIMESTAMPTZ DEFAULT NOW(),

    -- For undo grouping (edits within 5 minutes = same session)
    edit_session_id UUID NOT NULL
);

CREATE INDEX idx_statement_edits_statement ON statement_edits(statement_id);
CREATE INDEX idx_statement_edits_session ON statement_edits(edit_session_id);

-- Add modification tracking to statements
ALTER TABLE statements ADD COLUMN IF NOT EXISTS
    modified_by UUID REFERENCES users(id);
```

### sqlc Query Patterns

```sql
-- name: SaveStatementContent :one
UPDATE statements SET
    local_content = $2,
    local_content_html = $3,
    is_modified = true,
    modified_at = NOW(),
    modified_by = $4,
    sync_status = CASE
        WHEN sync_status = 'conflict' THEN 'conflict'
        ELSE 'modified'
    END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: RecordEdit :exec
INSERT INTO statement_edits (statement_id, content_before, content_after, edited_by, edit_session_id)
VALUES ($1, $2, $3, $4, $5);

-- name: GetEditHistory :many
SELECT * FROM statement_edits
WHERE statement_id = $1
ORDER BY edited_at DESC
LIMIT $2;

-- name: RevertToEdit :one
UPDATE statements SET
    local_content = (SELECT content_before FROM statement_edits WHERE id = $2),
    modified_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;
```

---

## 5. API Implementation Strategy

### Endpoints

```go
// internal/api/handlers/statement/handler.go
func (h *Handler) RegisterRoutes(r chi.Router) {
    r.Route("/api/v1/statements", func(r chi.Router) {
        r.Get("/{id}", h.GetStatement)
        r.Put("/{id}", h.SaveStatement)
        r.Get("/{id}/history", h.GetEditHistory)
        r.Post("/{id}/revert/{editId}", h.RevertToEdit)
    })
}
```

### Request/Response Schemas

```go
// internal/api/handlers/statement/schemas.go

type SaveStatementRequest struct {
    Content     string `json:"content" validate:"required"`
    ContentHTML string `json:"content_html" validate:"required"`
    SessionID   string `json:"session_id" validate:"required,uuid"`
}

type StatementResponse struct {
    ID              uuid.UUID  `json:"id"`
    ControlID       uuid.UUID  `json:"control_id"`
    StatementType   string     `json:"statement_type"`
    Content         string     `json:"content"`          // Original from ServiceNow
    ContentHTML     string     `json:"content_html"`
    LocalContent    *string    `json:"local_content"`    // Modified locally
    LocalContentHTML *string   `json:"local_content_html"`
    IsModified      bool       `json:"is_modified"`
    SyncStatus      string     `json:"sync_status"`
    ModifiedAt      *time.Time `json:"modified_at,omitempty"`
}

type EditHistoryResponse struct {
    Edits []EditEntry `json:"edits"`
}

type EditEntry struct {
    ID            uuid.UUID `json:"id"`
    ContentBefore string    `json:"content_before"`
    ContentAfter  string    `json:"content_after"`
    EditedAt      time.Time `json:"edited_at"`
    EditedBy      *string   `json:"edited_by,omitempty"`
}
```

---

## 6. Frontend Implementation Approach

### Zustand Editor Store

```typescript
// features/editor/stores/editorStore.ts
import { create } from 'zustand';

interface EditorState {
  // Current editing state
  statementId: string | null;
  isDirty: boolean;
  lastSavedAt: Date | null;
  isSaving: boolean;
  saveError: string | null;

  // Edit session for grouping history
  editSessionId: string;

  // Actions
  setStatementId: (id: string) => void;
  markDirty: () => void;
  markSaved: () => void;
  setSaving: (isSaving: boolean) => void;
  setSaveError: (error: string | null) => void;
  newEditSession: () => void;
}

export const useEditorStore = create<EditorState>((set) => ({
  statementId: null,
  isDirty: false,
  lastSavedAt: null,
  isSaving: false,
  saveError: null,
  editSessionId: crypto.randomUUID(),

  setStatementId: (id) => set({ statementId: id, isDirty: false }),
  markDirty: () => set({ isDirty: true }),
  markSaved: () => set({ isDirty: false, lastSavedAt: new Date(), saveError: null }),
  setSaving: (isSaving) => set({ isSaving }),
  setSaveError: (error) => set({ saveError: error }),
  newEditSession: () => set({ editSessionId: crypto.randomUUID() }),
}));
```

### Auto-Save Hook

```typescript
// features/editor/hooks/useAutoSave.ts
import { useEffect, useRef, useCallback } from 'react';
import { useDebouncedCallback } from 'use-debounce';

const AUTO_SAVE_DELAY = 30_000; // 30 seconds

export function useAutoSave(
  content: string,
  statementId: string,
  options?: { enabled?: boolean }
) {
  const { enabled = true } = options ?? {};
  const saveMutation = useSaveStatement();
  const { markDirty, markSaved, setSaving, setSaveError, editSessionId } = useEditorStore();
  const previousContent = useRef(content);

  const save = useCallback(async (contentToSave: string) => {
    if (!statementId || contentToSave === previousContent.current) return;

    setSaving(true);
    try {
      await saveMutation.mutateAsync({
        id: statementId,
        content: contentToSave,
        contentHtml: contentToSave, // TipTap HTML is the content
        sessionId: editSessionId,
      });
      previousContent.current = contentToSave;
      markSaved();
    } catch (error) {
      setSaveError(error instanceof Error ? error.message : 'Save failed');
    } finally {
      setSaving(false);
    }
  }, [statementId, editSessionId, saveMutation, markSaved, setSaving, setSaveError]);

  const debouncedSave = useDebouncedCallback(save, AUTO_SAVE_DELAY);

  useEffect(() => {
    if (!enabled) return;

    if (content !== previousContent.current) {
      markDirty();
      debouncedSave(content);
    }
  }, [content, enabled, markDirty, debouncedSave]);

  // Save immediately on unmount if dirty
  useEffect(() => {
    return () => {
      if (content !== previousContent.current) {
        save(content);
      }
    };
  }, [content, save]);

  // Manual save function
  const saveNow = useCallback(() => {
    debouncedSave.cancel();
    save(content);
  }, [content, save, debouncedSave]);

  return { saveNow };
}
```

### Save Indicator Component

```typescript
// features/editor/components/SaveIndicator.tsx
export function SaveIndicator() {
  const { isDirty, isSaving, lastSavedAt, saveError } = useEditorStore();

  if (saveError) {
    return (
      <div className="flex items-center gap-2 text-destructive">
        <AlertCircle className="h-4 w-4" />
        <span className="text-sm">Save failed</span>
      </div>
    );
  }

  if (isSaving) {
    return (
      <div className="flex items-center gap-2 text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        <span className="text-sm">Saving...</span>
      </div>
    );
  }

  if (isDirty) {
    return (
      <div className="flex items-center gap-2 text-muted-foreground">
        <Circle className="h-2 w-2 fill-current" />
        <span className="text-sm">Unsaved changes</span>
      </div>
    );
  }

  if (lastSavedAt) {
    return (
      <div className="flex items-center gap-2 text-muted-foreground">
        <Check className="h-4 w-4" />
        <span className="text-sm">
          Saved {formatDistanceToNow(lastSavedAt, { addSuffix: true })}
        </span>
      </div>
    );
  }

  return null;
}
```

### Editor Toolbar Component

```typescript
// features/editor/components/EditorToolbar.tsx
import { type Editor } from '@tiptap/react';

interface EditorToolbarProps {
  editor: Editor | null;
}

export function EditorToolbar({ editor }: EditorToolbarProps) {
  if (!editor) return null;

  return (
    <div className="flex items-center gap-1 p-2 border-b">
      {/* Text Formatting */}
      <ToolbarGroup>
        <ToolbarButton
          onClick={() => editor.chain().focus().toggleBold().run()}
          active={editor.isActive('bold')}
          icon={<Bold className="h-4 w-4" />}
          tooltip="Bold (Ctrl+B)"
        />
        <ToolbarButton
          onClick={() => editor.chain().focus().toggleItalic().run()}
          active={editor.isActive('italic')}
          icon={<Italic className="h-4 w-4" />}
          tooltip="Italic (Ctrl+I)"
        />
        <ToolbarButton
          onClick={() => editor.chain().focus().toggleUnderline().run()}
          active={editor.isActive('underline')}
          icon={<Underline className="h-4 w-4" />}
          tooltip="Underline (Ctrl+U)"
        />
      </ToolbarGroup>

      <Separator orientation="vertical" className="h-6" />

      {/* Lists */}
      <ToolbarGroup>
        <ToolbarButton
          onClick={() => editor.chain().focus().toggleBulletList().run()}
          active={editor.isActive('bulletList')}
          icon={<List className="h-4 w-4" />}
          tooltip="Bullet List"
        />
        <ToolbarButton
          onClick={() => editor.chain().focus().toggleOrderedList().run()}
          active={editor.isActive('orderedList')}
          icon={<ListOrdered className="h-4 w-4" />}
          tooltip="Numbered List"
        />
      </ToolbarGroup>

      <Separator orientation="vertical" className="h-6" />

      {/* Table */}
      <ToolbarGroup>
        <ToolbarButton
          onClick={() => editor.chain().focus().insertTable({ rows: 3, cols: 3 }).run()}
          icon={<Table className="h-4 w-4" />}
          tooltip="Insert Table"
        />
      </ToolbarGroup>

      <Separator orientation="vertical" className="h-6" />

      {/* Undo/Redo */}
      <ToolbarGroup>
        <ToolbarButton
          onClick={() => editor.chain().focus().undo().run()}
          disabled={!editor.can().undo()}
          icon={<Undo className="h-4 w-4" />}
          tooltip="Undo (Ctrl+Z)"
        />
        <ToolbarButton
          onClick={() => editor.chain().focus().redo().run()}
          disabled={!editor.can().redo()}
          icon={<Redo className="h-4 w-4" />}
          tooltip="Redo (Ctrl+Shift+Z)"
        />
      </ToolbarGroup>
    </div>
  );
}
```

---

## 7. Business Logic Implementation Hints

### Edit Session Logic

```go
// internal/domain/statement/service.go

const editSessionTimeout = 5 * time.Minute

func (s *Service) SaveContent(ctx context.Context, req SaveContentRequest) (*Statement, error) {
    // Get current statement
    stmt, err := s.repo.GetByID(ctx, req.ID)
    if err != nil {
        return nil, fmt.Errorf("failed to get statement: %w", err)
    }

    // Determine if this is a new edit session
    sessionID := req.SessionID
    if sessionID == "" {
        sessionID = uuid.New().String()
    }

    // Record edit for history (async, don't block save)
    go func() {
        contentBefore := stmt.LocalContent
        if contentBefore == nil {
            contentBefore = &stmt.Content
        }
        s.recordEdit(context.Background(), EditRecord{
            StatementID:   req.ID,
            ContentBefore: *contentBefore,
            ContentAfter:  req.Content,
            EditedBy:      req.UserID,
            SessionID:     sessionID,
        })
    }()

    // Save content
    return s.repo.SaveContent(ctx, req.ID, req.Content, req.ContentHTML, req.UserID)
}
```

### Content Sanitization

```go
// internal/domain/statement/sanitize.go
import "github.com/microcosm-cc/bluemonday"

var policy = bluemonday.UGCPolicy()

func SanitizeHTML(html string) string {
    return policy.Sanitize(html)
}

// Apply before saving
func (s *Service) SaveContent(ctx context.Context, req SaveContentRequest) (*Statement, error) {
    req.ContentHTML = SanitizeHTML(req.ContentHTML)
    // ... rest of save logic
}
```

---

## 8. Testing Implementation Approach

### Editor Integration Test

```typescript
// features/editor/components/StatementEditor.test.tsx
describe('StatementEditor', () => {
  it('renders split pane layout', () => {
    render(<StatementEditor statementId="stmt-1" />);

    expect(screen.getByTestId('context-pane')).toBeInTheDocument();
    expect(screen.getByTestId('editor-pane')).toBeInTheDocument();
  });

  it('loads statement content into editor', async () => {
    mockStatement({ content: '<p>Test content</p>' });

    render(<StatementEditor statementId="stmt-1" />);

    await waitFor(() => {
      expect(screen.getByText('Test content')).toBeInTheDocument();
    });
  });

  it('triggers auto-save after content change', async () => {
    jest.useFakeTimers();
    const saveSpy = jest.fn();
    mockSaveMutation(saveSpy);

    render(<StatementEditor statementId="stmt-1" />);

    const editor = screen.getByRole('textbox');
    await userEvent.type(editor, 'New content');

    // Fast-forward auto-save delay
    jest.advanceTimersByTime(30_000);

    await waitFor(() => {
      expect(saveSpy).toHaveBeenCalled();
    });
  });
});
```

### Backend Service Test

```go
// internal/domain/statement/service_test.go
func TestService_SaveContent(t *testing.T) {
    tests := []struct {
        name    string
        req     SaveContentRequest
        setup   func(*MockRepo)
        wantErr bool
    }{
        {
            name: "saves content and marks as modified",
            req: SaveContentRequest{
                ID:          uuid.New(),
                Content:     "<p>Updated content</p>",
                ContentHTML: "<p>Updated content</p>",
                UserID:      uuid.New(),
            },
            setup: func(repo *MockRepo) {
                repo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&Statement{
                    Content: "<p>Original</p>",
                }, nil)
                repo.EXPECT().SaveContent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
                    Return(&Statement{IsModified: true}, nil)
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Run test...
        })
    }
}
```

---

## 9. Keyboard Shortcuts

### Editor Shortcuts Configuration

```typescript
// features/editor/extensions/keyboardShortcuts.ts
import { Extension } from '@tiptap/core';

export const KeyboardShortcuts = Extension.create({
  name: 'keyboardShortcuts',

  addKeyboardShortcuts() {
    return {
      'Mod-s': () => {
        // Manual save
        const { saveNow } = useEditorStore.getState();
        saveNow?.();
        return true;
      },
      'Mod-Shift-s': () => {
        // Save and close (future: navigate back)
        return true;
      },
    };
  },
});
```

### Shortcut Reference
| Shortcut | Action |
|----------|--------|
| Ctrl+B | Bold |
| Ctrl+I | Italic |
| Ctrl+U | Underline |
| Ctrl+Z | Undo |
| Ctrl+Shift+Z | Redo |
| Ctrl+S | Save now |
| Tab | Indent list item |
| Shift+Tab | Outdent list item |

---

## 10. Performance Implementation Hints

### Editor Performance

```typescript
// Memoize toolbar to prevent re-renders
const MemoizedToolbar = React.memo(EditorToolbar);

// Debounce content updates to parent
const debouncedOnUpdate = useDebouncedCallback(onUpdate, 100);

// Use transaction batching for complex operations
editor.chain()
  .focus()
  .insertTable({ rows: 3, cols: 3 })
  .run();
```

### Content Storage

```go
// Store plain text separately for search indexing
type Statement struct {
    // ...
    ContentHTML  string  // Full HTML for rendering
    ContentText  string  // Plain text for search
}

// Extract text on save
func extractText(html string) string {
    doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
    return doc.Text()
}
```

---

## 11. Accessibility Requirements

### Editor Accessibility

```typescript
// Ensure editor has proper ARIA attributes
<EditorContent
  editor={editor}
  aria-label="Implementation statement editor"
  role="textbox"
  aria-multiline="true"
/>

// Toolbar buttons with labels
<ToolbarButton
  onClick={() => editor.chain().focus().toggleBold().run()}
  aria-label="Toggle bold"
  aria-pressed={editor.isActive('bold')}
/>
```

### Keyboard Navigation
- Tab through toolbar buttons
- Enter activates focused button
- Escape returns focus to editor
- All formatting available via keyboard shortcuts

---

## 12. Error Handling and Recovery

### Save Error Recovery

```typescript
// features/editor/hooks/useAutoSave.ts
const MAX_RETRY_ATTEMPTS = 3;
const RETRY_DELAY = 5000;

async function saveWithRetry(content: string, attempt = 1): Promise<void> {
  try {
    await saveMutation.mutateAsync({ /* ... */ });
  } catch (error) {
    if (attempt < MAX_RETRY_ATTEMPTS) {
      await new Promise(resolve => setTimeout(resolve, RETRY_DELAY));
      return saveWithRetry(content, attempt + 1);
    }
    throw error;
  }
}
```

### Unsaved Changes Warning

```typescript
// features/editor/components/EditorPane.tsx
useEffect(() => {
  const handleBeforeUnload = (e: BeforeUnloadEvent) => {
    if (isDirty) {
      e.preventDefault();
      e.returnValue = '';
    }
  };

  window.addEventListener('beforeunload', handleBeforeUnload);
  return () => window.removeEventListener('beforeunload', handleBeforeUnload);
}, [isDirty]);
```

---

## 13. Security Checklist

- [ ] HTML content sanitized before storage (bluemonday)
- [ ] XSS prevention in rendered content
- [ ] User can only edit statements they have access to
- [ ] Edit history preserved for audit
- [ ] Session ID validated as UUID
- [ ] Content size limits enforced (prevent DoS)
