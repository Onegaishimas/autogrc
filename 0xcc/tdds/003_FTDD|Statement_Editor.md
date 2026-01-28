# Technical Design Document: Statement Editor (F3)

**Document Version:** 1.0
**Date:** 2026-01-27
**Feature ID:** F3
**Related PRD:** 003_FPRD|Statement_Editor.md

---

## 1. Executive Summary

This TDD defines the technical architecture for the statement editor - the core authoring interface of ControlCRUD. The editor provides rich text editing with TipTap, split-pane layout for control context, and robust auto-save functionality.

**Key Technical Decisions:**
- TipTap 2.x for rich text editing (ProseMirror-based)
- Split-pane layout with resizable panels
- Debounced auto-save every 30 seconds
- Optimistic updates with version-based conflict detection

---

## 2. System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           FRONTEND                                       │
│  ┌───────────────┬────────────────────────────┬───────────────────────┐ │
│  │   Control     │       Statement            │     Control           │ │
│  │   Navigator   │       Editor               │     Requirements      │ │
│  │               │                            │     Panel             │ │
│  │  ┌─────────┐  │  ┌────────────────────┐   │  ┌─────────────────┐  │ │
│  │  │ Family  │  │  │ TipTap Editor      │   │  │ Control Meta    │  │ │
│  │  │ Filter  │  │  │                    │   │  │ Description     │  │ │
│  │  ├─────────┤  │  │ [Toolbar]          │   │  │ ODPs            │  │ │
│  │  │ Control │  │  │ [Content Area]     │   │  │ Enhancements    │  │ │
│  │  │ List    │  │  │                    │   │  └─────────────────┘  │ │
│  │  └─────────┘  │  └────────────────────┘   │                       │ │
│  └───────────────┴────────────────────────────┴───────────────────────┘ │
│                              │                                           │
└──────────────────────────────┼───────────────────────────────────────────┘
                               │ TanStack Query
                               ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           BACKEND (Go)                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────┐ │
│  │ Control         │  │ Statement       │  │ Autosave                │ │
│  │ Handler         │  │ Handler         │  │ Handler                 │ │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────────────┘ │
│           │                    │                    │                   │
│           ▼                    ▼                    ▼                   │
│  ┌─────────────────────────────────────────────────────────────────────┐│
│  │                      Statement Service                              ││
│  │  - Get/update statements                                           ││
│  │  - Version management                                              ││
│  │  - Autosave handling                                               ││
│  └─────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────┘
```

### Component Data Flow

```
User types in editor
       │
       ▼
TipTap editor state updates
       │
       ├──► onChange handler (debounced)
       │           │
       │           ▼
       │    Zustand store (dirty flag)
       │           │
       │           ▼
       │    Auto-save timer (30s)
       │           │
       │           ▼
       │    API: POST /statements/:id/autosave
       │
       └──► Manual save (Ctrl+S)
                   │
                   ▼
            API: PUT /statements/:id
                   │
                   ▼
            TanStack Query cache update
```

---

## 3. Technical Stack

### Frontend Components

| Component | Technology | Justification |
|-----------|------------|---------------|
| Rich Text | TipTap 2.x | Extensible, ProseMirror-based |
| Editor Extensions | @tiptap/extension-* | Tables, lists, formatting |
| Split Pane | react-resizable-panels | Flexible resizable layout |
| State | TanStack Query + Zustand | Server + client state |
| Shortcuts | react-hotkeys-hook | Keyboard shortcut handling |

### TipTap Extensions

```typescript
// Required extensions for statement editing
import StarterKit from '@tiptap/starter-kit';
import Table from '@tiptap/extension-table';
import TableRow from '@tiptap/extension-table-row';
import TableCell from '@tiptap/extension-table-cell';
import TableHeader from '@tiptap/extension-table-header';
import Placeholder from '@tiptap/extension-placeholder';
import CharacterCount from '@tiptap/extension-character-count';
```

### Backend Components

| Component | Technology | Justification |
|-----------|------------|---------------|
| HTML Sanitization | bluemonday | Go HTML sanitizer |
| Text Extraction | html-to-text | Plain text for search |
| Diff Generation | go-diff | Version comparison |

---

## 4. Data Design

### Statement Schema Updates

```sql
-- Already defined in F2, ensuring fields for editor
ALTER TABLE implementation_statements
    ADD COLUMN IF NOT EXISTS content_hash VARCHAR(64),
    ADD COLUMN IF NOT EXISTS autosave_content TEXT,
    ADD COLUMN IF NOT EXISTS autosave_at TIMESTAMPTZ;
```

### Editor State Model

```typescript
// src/features/editor/types.ts

interface EditorState {
  controlId: string;
  statementId: string;
  content: string;          // HTML content
  isDirty: boolean;         // Has unsaved changes
  isSaving: boolean;        // Save in progress
  lastSavedAt: Date | null;
  version: number;          // For conflict detection
}

interface ControlContext {
  controlNumber: string;
  title: string;
  description: string;
  responsibility: 'common' | 'hybrid' | 'system_specific';
  odps: ODP[];
  enhancements: Enhancement[];
}
```

---

## 5. API Design

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/v1/controls/:id | Get control with statement |
| PUT | /api/v1/statements/:id | Update statement |
| POST | /api/v1/statements/:id/autosave | Autosave content |
| GET | /api/v1/statements/:id/autosave | Get autosaved content |

### Update Statement

```go
// PUT /api/v1/statements/:id

type UpdateStatementRequest struct {
    ContentHTML string `json:"content_html" validate:"required,max=50000"`
    Version     int    `json:"version" validate:"required,min=1"`
}

type UpdateStatementResponse struct {
    ID        uuid.UUID `json:"id"`
    Version   int       `json:"version"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Returns 409 Conflict if version mismatch
```

### Autosave

```go
// POST /api/v1/statements/:id/autosave

type AutosaveRequest struct {
    ContentHTML string `json:"content_html" validate:"required"`
}

type AutosaveResponse struct {
    SavedAt time.Time `json:"saved_at"`
}

// Lightweight save, no version check
// Overwrites previous autosave
```

---

## 6. Component Architecture

### Frontend Component Structure

```
src/features/editor/
├── components/
│   ├── EditorLayout.tsx        # Main split-pane layout
│   ├── ControlNavigator/
│   │   ├── Navigator.tsx       # Control list with filters
│   │   ├── ControlItem.tsx     # Individual control item
│   │   └── NavigatorFilters.tsx
│   ├── StatementEditor/
│   │   ├── Editor.tsx          # TipTap editor wrapper
│   │   ├── Toolbar.tsx         # Formatting toolbar
│   │   ├── EditorContent.tsx   # Editor content area
│   │   └── SaveStatus.tsx      # Save indicator
│   └── RequirementsPanel/
│       ├── Panel.tsx           # Control requirements
│       ├── ControlMeta.tsx     # Title, number, responsibility
│       ├── Description.tsx     # Scrollable description
│       └── ODPList.tsx         # Organization-defined params
├── hooks/
│   ├── useEditor.ts            # TipTap editor setup
│   ├── useStatement.ts         # Statement CRUD
│   ├── useAutosave.ts          # Autosave logic
│   └── useKeyboardShortcuts.ts # Hotkeys
├── stores/
│   └── editorStore.ts          # Zustand store
└── utils/
    ├── htmlSanitizer.ts        # Client-side sanitization
    └── contentHash.ts          # Change detection
```

### TipTap Editor Setup

```typescript
// src/features/editor/hooks/useEditor.ts

import { useEditor } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import Table from '@tiptap/extension-table';
import TableRow from '@tiptap/extension-table-row';
import TableCell from '@tiptap/extension-table-cell';
import TableHeader from '@tiptap/extension-table-header';
import Placeholder from '@tiptap/extension-placeholder';
import CharacterCount from '@tiptap/extension-character-count';

export function useStatementEditor(initialContent: string) {
  const editor = useEditor({
    extensions: [
      StarterKit.configure({
        heading: { levels: [1, 2, 3] },
      }),
      Table.configure({ resizable: true }),
      TableRow,
      TableCell,
      TableHeader,
      Placeholder.configure({
        placeholder: 'Enter implementation statement...',
      }),
      CharacterCount.configure({ limit: 50000 }),
    ],
    content: initialContent,
    editorProps: {
      attributes: {
        class: 'prose prose-sm max-w-none focus:outline-none min-h-[300px]',
      },
    },
  });

  return editor;
}
```

### Autosave Implementation

```typescript
// src/features/editor/hooks/useAutosave.ts

import { useEffect, useRef } from 'react';
import { useMutation } from '@tanstack/react-query';
import { useEditorStore } from '../stores/editorStore';

const AUTOSAVE_INTERVAL = 30000; // 30 seconds
const AUTOSAVE_DEBOUNCE = 1000;  // 1 second after typing stops

export function useAutosave(statementId: string, getContent: () => string) {
  const { isDirty, setIsDirty, setSaveStatus } = useEditorStore();
  const lastSaveRef = useRef<string>('');

  const autosaveMutation = useMutation({
    mutationFn: (content: string) =>
      statementApi.autosave(statementId, { content_html: content }),
    onSuccess: () => {
      setSaveStatus('saved');
      lastSaveRef.current = getContent();
    },
    onError: () => {
      setSaveStatus('error');
    },
  });

  // Periodic autosave
  useEffect(() => {
    const interval = setInterval(() => {
      const content = getContent();
      if (isDirty && content !== lastSaveRef.current) {
        setSaveStatus('saving');
        autosaveMutation.mutate(content);
      }
    }, AUTOSAVE_INTERVAL);

    return () => clearInterval(interval);
  }, [isDirty, getContent]);

  // Save on blur
  useEffect(() => {
    const handleBlur = () => {
      const content = getContent();
      if (isDirty && content !== lastSaveRef.current) {
        autosaveMutation.mutate(content);
      }
    };

    window.addEventListener('blur', handleBlur);
    return () => window.removeEventListener('blur', handleBlur);
  }, [isDirty, getContent]);

  return autosaveMutation;
}
```

---

## 7. State Management

### Zustand Editor Store

```typescript
// src/features/editor/stores/editorStore.ts

import { create } from 'zustand';

interface EditorStore {
  // Current editing context
  currentControlId: string | null;
  currentStatementId: string | null;

  // Editor state
  isDirty: boolean;
  saveStatus: 'idle' | 'saving' | 'saved' | 'error';
  lastSavedAt: Date | null;

  // Panel state
  navigatorCollapsed: boolean;
  requirementsPanelCollapsed: boolean;

  // Actions
  setCurrentControl: (controlId: string, statementId: string) => void;
  setIsDirty: (dirty: boolean) => void;
  setSaveStatus: (status: 'idle' | 'saving' | 'saved' | 'error') => void;
  toggleNavigator: () => void;
  toggleRequirementsPanel: () => void;
}

export const useEditorStore = create<EditorStore>((set) => ({
  currentControlId: null,
  currentStatementId: null,
  isDirty: false,
  saveStatus: 'idle',
  lastSavedAt: null,
  navigatorCollapsed: false,
  requirementsPanelCollapsed: false,

  setCurrentControl: (controlId, statementId) =>
    set({ currentControlId: controlId, currentStatementId: statementId, isDirty: false }),
  setIsDirty: (dirty) => set({ isDirty: dirty }),
  setSaveStatus: (status) => set({
    saveStatus: status,
    lastSavedAt: status === 'saved' ? new Date() : undefined
  }),
  toggleNavigator: () => set((s) => ({ navigatorCollapsed: !s.navigatorCollapsed })),
  toggleRequirementsPanel: () => set((s) => ({ requirementsPanelCollapsed: !s.requirementsPanelCollapsed })),
}));
```

### TanStack Query for Server State

```typescript
// src/features/editor/hooks/useStatement.ts

export function useStatement(statementId: string) {
  return useQuery({
    queryKey: ['statement', statementId],
    queryFn: () => statementApi.get(statementId),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

export function useUpdateStatement() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateStatementRequest }) =>
      statementApi.update(id, data),
    onSuccess: (response, { id }) => {
      queryClient.setQueryData(['statement', id], (old: Statement) => ({
        ...old,
        ...response.data,
      }));
    },
    onError: (error) => {
      if (error.status === 409) {
        // Handle version conflict
        toast.error('This statement was modified elsewhere. Please refresh.');
      }
    },
  });
}
```

---

## 8. Security Considerations

### HTML Sanitization

```go
// internal/domain/statement/sanitizer.go

import "github.com/microcosm-cc/bluemonday"

var statementPolicy = bluemonday.NewPolicy()

func init() {
    // Allow safe HTML elements
    statementPolicy.AllowElements("p", "br", "hr")
    statementPolicy.AllowElements("strong", "em", "u", "s")
    statementPolicy.AllowElements("h1", "h2", "h3")
    statementPolicy.AllowElements("ul", "ol", "li")
    statementPolicy.AllowElements("table", "thead", "tbody", "tr", "th", "td")
    statementPolicy.AllowAttrs("colspan", "rowspan").OnElements("th", "td")

    // No scripts, iframes, or external content
    // No style attributes (use Tailwind classes if needed)
}

func SanitizeStatement(html string) string {
    return statementPolicy.Sanitize(html)
}
```

### Version Conflict Detection

```go
// internal/domain/statement/service.go

func (s *StatementService) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Statement, error) {
    // Start transaction
    tx, _ := s.db.BeginTx(ctx, nil)
    defer tx.Rollback()

    // Get current version with lock
    current, err := s.repo.GetForUpdate(ctx, tx, id)
    if err != nil {
        return nil, err
    }

    // Check version
    if current.Version != req.Version {
        return nil, ErrVersionConflict
    }

    // Update with incremented version
    stmt := &Statement{
        ID:              id,
        ContentHTML:     SanitizeStatement(req.ContentHTML),
        ContentPlain:    ExtractText(req.ContentHTML),
        Version:         req.Version + 1,
        LocalModifiedAt: timePtr(time.Now()),
        LocalModifiedBy: req.UserID,
    }

    if err := s.repo.Update(ctx, tx, stmt); err != nil {
        return nil, err
    }

    tx.Commit()
    return stmt, nil
}
```

---

## 9. Performance & Scalability

### Performance Targets

| Operation | Target | Approach |
|-----------|--------|----------|
| Editor load | < 1s | Lazy load TipTap |
| Keystroke response | < 16ms | ProseMirror efficiency |
| Autosave | < 500ms | Async, non-blocking |
| Control switch | < 500ms | Prefetch next controls |

### Optimization Strategies

1. **Lazy Loading TipTap**
   ```typescript
   const Editor = lazy(() => import('./StatementEditor'));
   ```

2. **Prefetch Adjacent Controls**
   ```typescript
   // Prefetch next/prev controls on hover
   const prefetchControl = (id: string) => {
     queryClient.prefetchQuery(['control', id], () => controlApi.get(id));
   };
   ```

3. **Debounced Change Detection**
   ```typescript
   const onUpdate = useMemo(
     () => debounce(() => setIsDirty(true), 300),
     []
   );
   ```

---

## 10. Testing Strategy

### Component Tests

```typescript
// src/features/editor/components/__tests__/Editor.test.tsx

describe('StatementEditor', () => {
  it('renders with initial content', () => {
    render(<Editor content="<p>Test content</p>" />);
    expect(screen.getByText('Test content')).toBeInTheDocument();
  });

  it('marks as dirty on content change', async () => {
    const { user } = render(<Editor content="" />);
    await user.type(screen.getByRole('textbox'), 'New text');
    expect(useEditorStore.getState().isDirty).toBe(true);
  });

  it('triggers autosave after interval', async () => {
    vi.useFakeTimers();
    const autosaveSpy = vi.spyOn(statementApi, 'autosave');

    render(<Editor content="" statementId="123" />);
    await user.type(screen.getByRole('textbox'), 'New text');

    vi.advanceTimersByTime(30000);
    expect(autosaveSpy).toHaveBeenCalled();
  });
});
```

### E2E Tests

```typescript
// tests/e2e/editor.spec.ts

test('user can edit and save statement', async ({ page }) => {
  await page.goto('/workspace/system-1/AC-1');

  const editor = page.locator('.tiptap-editor');
  await editor.fill('Updated implementation statement');

  await page.keyboard.press('Control+s');
  await expect(page.locator('.save-status')).toHaveText('Saved');
});

test('keyboard navigation between controls', async ({ page }) => {
  await page.goto('/workspace/system-1/AC-1');

  await page.keyboard.press('Control+ArrowDown');
  await expect(page).toHaveURL(/AC-2/);
});
```

---

## 11. Deployment & DevOps

### Bundle Optimization

```typescript
// vite.config.ts
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          tiptap: ['@tiptap/react', '@tiptap/starter-kit', '@tiptap/extension-table'],
        },
      },
    },
  },
});
```

### Feature Flags

```typescript
// For gradual rollout
const FEATURES = {
  EDITOR_V2: import.meta.env.VITE_EDITOR_V2 === 'true',
};
```

---

## 12. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| TipTap learning curve | Medium | Medium | Spike on complex extensions |
| Data loss on crash | Low | High | Autosave, localStorage backup |
| Performance on large content | Low | Medium | Character limit, virtualization |
| Cross-browser issues | Medium | Medium | Test on all target browsers |

---

## 13. Development Phases

### Phase 1: Core Editor (3 days)
- TipTap setup with extensions
- Basic toolbar
- Content save/load

### Phase 2: Layout & Navigation (2 days)
- Split-pane layout
- Control navigator
- Requirements panel

### Phase 3: Auto-save & State (2 days)
- Autosave implementation
- Dirty state tracking
- Version conflict handling

### Phase 4: Polish (2 days)
- Keyboard shortcuts
- Accessibility
- Error handling

### Phase 5: Testing (1 day)
- Component tests
- E2E tests

**Estimated Total: 10 days**

---

## 14. Open Technical Questions

1. **Collaborative Editing:** Should we plan for future real-time collaboration?
2. **Offline Support:** How critical is offline editing capability?
3. **Content Size:** What's the largest expected statement content?
4. **Table Complexity:** How complex can tables be (nested, merged cells)?
