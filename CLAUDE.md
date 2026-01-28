# Project: ControlCRUD - Control Statement Authoring and Lifecycle Management System

## Current Status
- **Phase:** Phase 1 Implementation - Infrastructure & F1 In Progress
- **Last Session:** 2026-01-28 - Created Docker infrastructure, started F1 implementation
- **Next Steps:** Complete F1 Task 1.7 (run migration), then Task 2.0 (Crypto Service)
- **Active Document:** 0xcc/tasks/001_FTASKS|ServiceNow_Connection.md
- **Current Feature:** F1 - ServiceNow GRC Connection (Task 1.6 complete, 1.7 ready)

## Project Summary
**ControlCRUD** is a web-based application for authoring NIST 800-53 control implementation statements with:
- AI-assisted drafting and refinement
- Reference oracle integration (NIST 800-53, FedRAMP, CIS, organizational templates)
- Control tailoring workflows aligned with NIST RMF
- Bidirectional synchronization with ServiceNow GRC (Policy & Compliance module)
- Full lifecycle support from blank-sheet development to continuous updates

## Key Business Drivers
1. Accelerate ATO timelines (target: 40% reduction)
2. Improve implementation statement quality/consistency
3. Reduce SME burden - enable junior staff authoring
4. Centralize authoring outside ServiceNow UI limitations

## Primary Stakeholders
- **CISO** - Executive Sponsor
- **System Owners/ISSOs** - Primary users (review/approve)
- **Control Authors** - Primary users (drafting)
- **Compliance Managers** - Secondary users (templates, inherited controls)
- **Assessors** - Reference users (QA)

## Quick Resume Commands
```bash
# XCC session start sequence
"Please help me resume where I left off"
# Or manual if needed:
@CLAUDE.md
@0xcc/session_state.json
ls -la 0xcc/*/

# Research integration (requires ref MCP server)
# Format: "Use /mcp ref search '[context-specific search term]'"

# Load project context (after Project PRD exists)
@0xcc/prds/000_PPRD|[project-name].md
@0xcc/adrs/000_PADR|[project-name].md

# Load current work area based on phase
@0xcc/prds/      # For PRD work
@0xcc/tdds/      # For TDD work  
@0xcc/tids/      # For TID work
@0xcc/tasks/     # For task execution
```

## Housekeeping Commands
```bash
"Please create a checkpoint"        # Save complete state
"Please help me resume"            # Restore context for new session
"My context is getting too large"  # Clean context, restore essentials
"Please save the session transcript" # Save session transcript
"Please show me project status"    # Display current state
```

## Project Standards

### Technology Stack

- **Frontend:** React 18 + TypeScript 5, Vite, TanStack Query, Zustand, Shadcn/ui, Tailwind CSS, TipTap editor
- **Backend:** Go 1.22+, Chi/Echo router, sqlc + pgx for PostgreSQL
- **Database:** PostgreSQL 15+ with JSONB for flexible data
- **Testing:** Vitest + RTL (frontend), Go testing + testify (backend), Playwright (E2E)
- **Auth:** Enterprise SSO (SAML/OIDC), JWT sessions
- **AI:** Direct Claude API integration

### Code Organization

**Backend Structure:**
```
backend/
├── cmd/server/          # Entry point
├── internal/
│   ├── api/handlers/    # HTTP handlers
│   ├── domain/          # Business logic by domain
│   ├── infrastructure/  # External service clients
│   └── pkg/             # Shared utilities
└── migrations/          # SQL migrations
```

**Frontend Structure:**
```
frontend/src/
├── components/          # UI components (ui/, common/, feature-specific)
├── features/            # Feature modules (auth, workspace, sync)
├── hooks/               # Custom React hooks
├── services/            # API service layer
├── stores/              # Zustand state stores
└── types/               # TypeScript definitions
```

### Coding Patterns

**Go Conventions:**
- Use `internal/` for non-exported packages
- Explicit error handling: `fmt.Errorf("failed to get control: %w", err)`
- Interfaces for external dependencies (testability)
- Structured logging with zerolog

**TypeScript/React Conventions:**
- Functional components with hooks only
- TanStack Query for all server state
- Zustand for minimal client state
- Zod schemas for runtime validation
- Explicit prop interfaces (no `any`)

**API Conventions:**
- RESTful URLs: `/api/v1/{resource}/{id}`
- JSON request/response with `data` wrapper
- Structured error responses with codes
- Pagination via `?page=1&per_page=20`

### Quality Requirements

**Testing:**
- Unit tests for business logic (80%+ coverage)
- Integration tests for API endpoints (70%+)
- E2E tests for critical user flows

**Code Review:**
- All changes via pull request
- One approval required
- Tests passing, no lint errors

**Commits:**
- Conventional Commits: `feat(scope): description`
- Types: feat, fix, docs, refactor, test, chore

### Architecture Principles

1. **Separation of Concerns:** Frontend for UI, Backend for logic, Database for persistence
2. **Dependency Injection:** Interfaces for external services
3. **Fail Fast:** Validate at boundaries, meaningful errors
4. **Explicit Over Implicit:** No magic, typed everything

### Security Requirements

- SSO authentication required (no standalone auth)
- RBAC authorization on protected endpoints
- Input validation on all API requests
- Audit logging for all data mutations

### Test Commands

```bash
# Backend
cd backend && go test ./...

# Frontend
cd frontend && pnpm test
cd frontend && pnpm test:e2e
```

## AI Dev Tasks Framework Workflow

### Document Creation Sequence
1. **Project Foundation**
   - `000_PPRD|[project-name].md` → `0xcc/prds/` (Project PRD)
   - `000_PADR|[project-name].md` → `0xcc/adrs/` (Architecture Decision Record)
   - Update this CLAUDE.md with Project Standards from ADR

2. **Feature Development** (repeat for each feature)
   - `[###]_FPRD|[feature-name].md` → `0xcc/prds/` (Feature PRD)
   - `[###]_FTDD|[feature-name].md` → `0xcc/tdds/` (Technical Design Doc)
   - `[###]_FTID|[feature-name].md` → `0xcc/tids/` (Technical Implementation Doc)
   - `[###]_FTASKS|[feature-name].md` → `0xcc/tasks/` (Task List)

### Instruction Documents Reference
- `@0xcc/instruct/001_create-project-prd.md` - Creates project vision and feature breakdown
- `@0xcc/instruct/002_create-adr.md` - Establishes tech stack and standards
- `@0xcc/instruct/003_create-feature-prd.md` - Details individual feature requirements
- `@0xcc/instruct/004_create-tdd.md` - Creates technical architecture and design
- `@0xcc/instruct/005_create-tid.md` - Provides implementation guidance and coding hints
- `@0xcc/instruct/006_generate-tasks.md` - Generates actionable development tasks
- `@0xcc/instruct/007_process-task-list.md` - Guides task execution and progress tracking
- `@0xcc/instruct/008_housekeeping.md` - Session management and context preservation

## Document Inventory

### Pre-Framework Documents
- ✅ 0xcc/docs/BRD_ControlCRUD.md (BABOK-aligned Business Requirements Document)

### Project Level Documents
- ✅ 0xcc/prds/000_PPRD|ControlCRUD.md (Project PRD)
- ✅ 0xcc/adrs/000_PADR|ControlCRUD.md (Architecture Decision Record)

### Phase 1 Feature Documents (Core Sync MVP)

**F1: ServiceNow GRC Connection**
- ✅ 0xcc/prds/001_FPRD|ServiceNow_Connection.md (Feature PRD)
- ✅ 0xcc/tdds/001_FTDD|ServiceNow_Connection.md (Technical Design Doc)
- ✅ 0xcc/tids/001_FTID|ServiceNow_Connection.md (Technical Implementation Doc)
- ✅ 0xcc/tasks/001_FTASKS|ServiceNow_Connection.md (Task List)

**F2: Control Package Pull**
- ✅ 0xcc/prds/002_FPRD|Control_Package_Pull.md (Feature PRD)
- ✅ 0xcc/tdds/002_FTDD|Control_Package_Pull.md (Technical Design Doc)
- ✅ 0xcc/tids/002_FTID|Control_Package_Pull.md (Technical Implementation Doc)
- ✅ 0xcc/tasks/002_FTASKS|Control_Package_Pull.md (Task List)

**F3: Statement Editor**
- ✅ 0xcc/prds/003_FPRD|Statement_Editor.md (Feature PRD)
- ✅ 0xcc/tdds/003_FTDD|Statement_Editor.md (Technical Design Doc)
- ✅ 0xcc/tids/003_FTID|Statement_Editor.md (Technical Implementation Doc)
- ✅ 0xcc/tasks/003_FTASKS|Statement_Editor.md (Task List)

**F4: Control Package Push**
- ✅ 0xcc/prds/004_FPRD|Control_Package_Push.md (Feature PRD)
- ✅ 0xcc/tdds/004_FTDD|Control_Package_Push.md (Technical Design Doc)
- ✅ 0xcc/tids/004_FTID|Control_Package_Push.md (Technical Implementation Doc)
- ✅ 0xcc/tasks/004_FTASKS|Control_Package_Push.md (Task List)

**F5: Sync Audit Trail**
- ✅ 0xcc/prds/005_FPRD|Sync_Audit_Trail.md (Feature PRD)
- ✅ 0xcc/tdds/005_FTDD|Sync_Audit_Trail.md (Technical Design Doc)
- ✅ 0xcc/tids/005_FTID|Sync_Audit_Trail.md (Technical Implementation Doc)
- ✅ 0xcc/tasks/005_FTASKS|Sync_Audit_Trail.md (Task List)

### Phase 2 Feature Documents (AI-Assisted Authoring) - Pending
- ❌ F6: AI Draft Generation
- ❌ F7: AI Refinement Assistant
- ❌ F8: Completeness Validator
- ❌ F9: Reference Oracle Panel
- ❌ F10: Organizational Template Library

### Phase 3 Feature Documents (Advanced Features) - Pending
- ❌ F11: Evidence Management
- ❌ F12: Review Workflow
- ❌ F13: Inherited Control Visualization
- ❌ F14: ODP Management
- ❌ F15: Administration Dashboard

### Status Indicators
- ✅ **Complete:** Document finished and reviewed
- ⏳ **In Progress:** Currently being worked on
- ❌ **Pending:** Not yet started
- 🔄 **Needs Update:** Requires revision based on changes

## Housekeeping Status
- **Last Checkpoint:** [Date/Time] - [Brief description]
- **Last Transcript Save:** [Date/Time] - [File location in 0xcc/transcripts/]
- **Context Health:** Good/Moderate/Needs Cleanup
- **Session Count:** [Number] sessions since project start
- **Total Development Time:** [Estimated hours]

## Task Execution Standards

### Completion Protocol
- ✅ One sub-task at a time, ask permission before next
- ✅ Mark sub-tasks complete immediately: `[ ]` → `[x]`
- ✅ When parent task complete: Run tests → Stage → Clean → Commit → Mark parent complete
- ✅ Never commit without passing tests
- ✅ Always clean up temporary files before commit

### Commit Message Format
```bash
git commit -m "feat: [brief description]" -m "- [key change 1]" -m "- [key change 2]" -m "Related to [Task#] in [PRD]"
```

### Test Commands
*[Will be defined in ADR, examples:]*
- **Frontend:** `npm test` or `npm run test:unit`
- **Backend:** `pytest` or `python -m pytest` 
- **Full Suite:** `[project-specific command]`

## Code Quality Checklist

### Before Any Commit
- [ ] All tests passing
- [ ] No console.log/print debugging statements
- [ ] No commented-out code blocks
- [ ] No temporary files (*.tmp, .cache, etc.)
- [ ] Code follows project naming conventions
- [ ] Functions/methods have docstrings if required
- [ ] Error handling implemented per ADR standards

### File Organization Rules
*[Will be defined in ADR, examples:]*
- Place test files alongside source files: `Component.tsx` + `Component.test.tsx`
- Follow directory structure from ADR
- Use naming conventions: `[Feature][Type].extension`
- Import statements organized: external → internal → relative
- Framework files in `0xcc/` directory, project files in standard locations

## Context Management

### Session End Protocol
```bash
# 1. Update CLAUDE.md status section
# 2. Create session summary
"Please create a checkpoint"
# 3. Commit progress
git add .
git commit -m "docs: completed [task] - Next: [specific action]"
```

### Context Recovery (If Lost)
```bash
# Mild context loss
@CLAUDE.md
@0xcc/session_state.json
ls -la 0xcc/*/
@0xcc/instruct/[current-phase].md

# Severe context loss  
@CLAUDE.md
@0xcc/prds/000_PPRD|[project-name].md
@0xcc/adrs/000_PADR|[project-name].md
ls -la 0xcc/*/
@0xcc/instruct/
```

### Resume Commands for Next Session
```bash
# Standard resume sequence
"Please help me resume where I left off"
# Or manual if needed:
@CLAUDE.md
@0xcc/session_state.json
@[specific-file-currently-working-on]
# Specific next action: [detailed action]
```

## Progress Tracking

### Task List Maintenance
- Update task list file after each sub-task completion
- Add newly discovered tasks as they emerge
- Update "Relevant Files" section with any new files created/modified
- Include one-line description for each file's purpose
- Distinguish between framework files (0xcc/) and project files (src/, tests/, etc.)

### Status Indicators for Tasks
- `[ ]` = Not started
- `[x]` = Completed
- `[~]` = In progress (use sparingly, only for current sub-task)
- `[?]` = Blocked/needs clarification

### Session Documentation
After each development session, update:
- Current task position in this CLAUDE.md
- Any blockers or questions encountered
- Next session starting point
- Files modified in this session (both 0xcc/ and project files)

## Implementation Patterns

### Error Handling
*[Will be defined in ADR - placeholder for standards]*
- Use project-standard error handling patterns from ADR
- Always handle both success and failure cases
- Log errors with appropriate level (error/warn/info)
- User-facing error messages should be friendly

### Testing Patterns
*[Will be defined in ADR - placeholder for standards]*
- Each function/component gets a test file
- Test naming: `describe('[ComponentName]', () => { it('should [behavior]', () => {})})`
- Mock external dependencies
- Test both happy path and error cases
- Aim for [X]% coverage per ADR standards

## Debugging Protocols

### When Tests Fail
1. Read error message carefully
2. Check recent changes for obvious issues
3. Run individual test to isolate problem
4. Use debugger/console to trace execution
5. Check dependencies and imports
6. Ask for help if stuck > 30 minutes

### When Task is Unclear
1. Review original PRD requirements
2. Check TDD for design intent
3. Look at TID for implementation hints
4. Ask clarifying questions before proceeding
5. Update task description for future clarity

## Feature Priority Order
*From Project PRD - Phased Delivery Approach*

**Phase 1 - Core Sync MVP (Must Have):**
1. F1: ServiceNow GRC Connection
2. F2: Control Package Pull
3. F3: Statement Editor
4. F4: Control Package Push
5. F5: Sync Audit Trail

**Phase 2 - AI-Assisted Authoring (Should Have):**
6. F6: AI Draft Generation
7. F7: AI Refinement Assistant
8. F8: Completeness Validator
9. F9: Reference Oracle Panel
10. F10: Organizational Template Library

**Phase 3 - Advanced Features (Could Have):**
11. F11: Evidence Management
12. F12: Review Workflow
13. F13: Inherited Control Visualization
14. F14: ODP Management
15. F15: Administration Dashboard

## Session History Log

### Session 1: 2026-01-27 - Project Initialization & BRD Creation
- **Accomplished:**
  - Defined project scope for ControlCRUD (ServiceNow GRC control statement authoring)
  - Created comprehensive BABOK-aligned Business Requirements Document
  - Established NIST RMF and 800-53 terminology standards
  - Identified key stakeholders: CISO, ISSOs, Control Authors, Compliance Managers, Assessors
  - Defined core functional areas: Control Workspace, Statement Authoring, ServiceNow Integration, Evidence Management, Reference Oracles
- **Files Created:**
  - 0xcc/docs/BRD_ControlCRUD.md (BABOK-aligned BRD)
- **Key Decisions:**
  - Web-based application with SSO (SAML/OIDC)
  - Bidirectional sync with ServiceNow GRC (out-of-box Policy & Compliance)
  - AI-assisted authoring (full suite: draft, refine, validate, evidence mapping)
  - Support NIST 800-53 Rev 5 + FedRAMP baselines + organizational baseline
  - Department-level scale (5-20 concurrent users)

### Session 1 (cont): 2026-01-27 - Project PRD Creation
- **Accomplished:**
  - Created Project PRD with phased delivery approach
  - Defined 15 features across 3 phases
  - Established success metrics and KPIs
  - Documented user personas (Control Author, ISSO, Compliance Manager, Assessor)
- **Files Created:**
  - 0xcc/prds/000_PPRD|ControlCRUD.md (Project PRD)
- **Key Decisions:**
  - Phase 1 MVP: Core sync loop (Pull → Edit → Push) without AI
  - Phase 2: Add AI-assisted authoring and reference oracles
  - Phase 3: Evidence management, workflows, administration
  - All 800-53 control families supported equally from start
  - Organizational templates imported from existing Common Controls (Word/Excel)

### Session 1 (cont): 2026-01-27 - Architecture Decision Record
- **Accomplished:**
  - Created comprehensive ADR with technology stack decisions
  - Defined development standards and coding patterns
  - Established security, performance, and quality guidelines
  - Added Project Standards section to CLAUDE.md
- **Files Created:**
  - 0xcc/adrs/000_PADR|ControlCRUD.md (Architecture Decision Record)
- **Key Technology Decisions:**
  - Architecture: Traditional 3-tier (frontend, backend API, database)
  - Frontend: React 18 + TypeScript, Vite, TanStack Query, Shadcn/ui, TipTap
  - Backend: Go 1.22+, Chi/Echo, sqlc + pgx
  - Database: PostgreSQL 15+ with JSONB
  - AI: Direct Claude API (no abstraction layer)
  - Testing: Balanced pyramid (unit + integration + E2E)
  - Approach: Production-ready from start

### Session 1 (cont): 2026-01-27 - Phase 1 Feature PRDs & TDDs
- **Accomplished:**
  - Created all 5 Phase 1 Feature PRDs (F1-F5)
  - Created all 5 Phase 1 Technical Design Documents (F1-F5)
  - Established complete technical architecture for Core Sync MVP
- **Files Created:**
  - 0xcc/prds/001_FPRD|ServiceNow_Connection.md
  - 0xcc/prds/002_FPRD|Control_Package_Pull.md
  - 0xcc/prds/003_FPRD|Statement_Editor.md
  - 0xcc/prds/004_FPRD|Control_Package_Push.md
  - 0xcc/prds/005_FPRD|Sync_Audit_Trail.md
  - 0xcc/tdds/001_FTDD|ServiceNow_Connection.md
  - 0xcc/tdds/002_FTDD|Control_Package_Pull.md
  - 0xcc/tdds/003_FTDD|Statement_Editor.md
  - 0xcc/tdds/004_FTDD|Control_Package_Push.md
  - 0xcc/tdds/005_FTDD|Sync_Audit_Trail.md
- **Key Design Decisions:**
  - F1: AES-256-GCM credential encryption, connection pooling, status caching
  - F2: Background job pattern for pulls, ServiceNow Table API pagination
  - F3: TipTap editor with split-pane layout, 30-second auto-save
  - F4: Optimistic concurrency with sys_updated_on conflict detection
  - F5: Append-only audit tables, 7-year retention, CSV/PDF export
- **Estimated Development Time:** ~47 days total for Phase 1

### Session 1 (cont): 2026-01-27 - Phase 1 Technical Implementation Documents
- **Accomplished:**
  - Created all 5 Phase 1 Technical Implementation Documents (F1-F5)
  - Defined detailed file structures for backend and frontend
  - Provided code patterns, implementation hints, and security checklists
- **Files Created:**
  - 0xcc/tids/001_FTID|ServiceNow_Connection.md
  - 0xcc/tids/002_FTID|Control_Package_Pull.md
  - 0xcc/tids/003_FTID|Statement_Editor.md
  - 0xcc/tids/004_FTID|Control_Package_Push.md
  - 0xcc/tids/005_FTID|Sync_Audit_Trail.md
- **Key Implementation Patterns:**
  - F1: AES-256-GCM crypto service, retryablehttp client, TanStack Query hooks
  - F2: Background job executor, ServiceNow pagination, UPSERT queries
  - F3: TipTap editor setup, Zustand editor store, debounced auto-save
  - F4: Conflict checker, concurrent push executor, diff viewer component
  - F5: Partitioned audit tables, squirrel query builder, CSV/PDF exporters
### Session 1 (cont): 2026-01-28 - Phase 1 Task Lists
- **Accomplished:**
  - Created all 5 Phase 1 Task Lists (F1-F5)
  - Defined hierarchical tasks with parent tasks and detailed sub-tasks
  - Identified all relevant files for each feature
  - Phase 1 planning is now 100% complete
- **Files Created:**
  - 0xcc/tasks/001_FTASKS|ServiceNow_Connection.md (10 parent tasks, ~50 sub-tasks)
  - 0xcc/tasks/002_FTASKS|Control_Package_Pull.md (13 parent tasks, ~70 sub-tasks)
  - 0xcc/tasks/003_FTASKS|Statement_Editor.md (18 parent tasks, ~90 sub-tasks)
  - 0xcc/tasks/004_FTASKS|Control_Package_Push.md (17 parent tasks, ~80 sub-tasks)
  - 0xcc/tasks/005_FTASKS|Sync_Audit_Trail.md (17 parent tasks, ~90 sub-tasks)
- **Task Summary:**
  - Total: ~75 parent tasks, ~380 sub-tasks across 5 features
  - F1 (Connection): Database, crypto, ServiceNow client, API, frontend
  - F2 (Pull): Pagination, domains (system/control/statement), job executor, wizard UI
  - F3 (Editor): TipTap setup, auto-save, split-pane, toolbar, context pane
  - F4 (Push): Conflict detection, concurrent executor, resolution UI, workflow
  - F5 (Audit): Partitioned tables, query builder, exporters, filters UI
- **Next:** Begin implementation using @0xcc/instruct/007_process-task-list.md

### Session 2: 2026-01-28 - Infrastructure Setup & F1 Implementation Start
- **Accomplished:**
  - Created complete Docker Compose infrastructure for containerized deployment
  - Set up PostgreSQL 16-alpine container with auto-migration support
  - Created Nginx reverse proxy configuration for autogrc.mcslab.io
  - Created multi-stage Dockerfiles for backend (Go) and frontend (React)
  - Started F1 implementation: completed database migration (Tasks 1.1-1.6)
  - Updated ADR with comprehensive containerization decisions
- **Files Created:**
  - docker-compose.yml (orchestration for postgres, backend, frontend, nginx)
  - backend/Dockerfile (Go 1.22 multi-stage build)
  - frontend/Dockerfile (Node 20 + Nginx multi-stage build)
  - nginx/nginx.conf (main configuration with rate limiting, gzip)
  - nginx/conf.d/autogrc.conf (server blocks for HTTPS and development)
  - .env.example (environment configuration template)
  - backend/migrations/20260127_001_create_servicenow_connections.sql
  - Backend directory structure (cmd/, internal/, migrations/)
- **Key Infrastructure Decisions:**
  - Base URL: autogrc.mcslab.io
  - PostgreSQL 16-alpine with auto-init migrations
  - Nginx with TLS 1.2+, rate limiting (10 req/s API, 30 req/s static)
  - Health checks on all containers
  - Non-root user in backend container for security
- **Next:** Run `docker compose up -d postgres` to verify migration, then continue F1 Task 2.0 (Crypto Service)

*[Add new sessions as they occur]*

## Research Integration

### MCP Research Support
When available, the framework supports research integration via:
```bash
# Use MCP ref server for contextual research
/mcp ref search "[context-specific query]"

# Research is integrated into all instruction documents as option B
# Example: "🔍 Research first: Use /mcp ref search 'MVP development timeline'"
```

### Research History Tracking
- Research queries and findings captured in session transcripts
- Key research decisions documented in session state
- Research context preserved across sessions for consistency

## Quick Reference

### 0xcc Folder Structure
```
project-root/
├── CLAUDE.md                       # This file (project memory)
├── 0xcc/                           # XCC Framework directory
│   ├── adrs/                       # Architecture Decision Records
│   ├── docs/                       # Additional documentation
│   ├── instruct/                   # Framework instruction files
│   ├── prds/                       # Product Requirements Documents
│   ├── tasks/                      # Task Lists
│   ├── tdds/                       # Technical Design Documents
│   ├── tids/                       # Technical Implementation Documents
│   ├── transcripts/                # Session transcripts
│   ├── checkpoints/                # Automated state backups
│   ├── scripts/                    # Optional automation scripts
│   ├── session_state.json          # Current session tracking
│   └── research_context.json       # Research history and context
├── src/                            # Your project code
├── tests/                          # Your project tests
└── README.md                       # Project README
```

### File Naming Convention
- **Project Level:** `000_PPRD|ProjectName.md`, `000_PADR|ProjectName.md`
- **Feature Level:** `001_FPRD|FeatureName.md`, `001_FTDD|FeatureName.md`, etc.
- **Sequential:** Use 001, 002, 003... for features in priority order
- **Framework Files:** All in `0xcc/` directory for clear organization
- **Project Files:** Standard locations (src/, tests/, package.json, etc.)

### Emergency Contacts & Resources
- **Framework Documentation:** @0xcc/instruct/000_README.md
- **Current Project PRD:** @0xcc/prds/000_PPRD|[project-name].md (after creation)
- **Tech Standards:** @0xcc/adrs/000_PADR|[project-name].md (after creation)
- **Housekeeping Guide:** @0xcc/instruct/008_housekeeping.md

---

**Framework Version:** 1.1  
**Last Updated:** [Current Date]  
**Project Started:** [Start Date]  
**Structure:** 0xcc framework with MCP research integration