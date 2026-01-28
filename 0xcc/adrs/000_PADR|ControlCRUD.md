# Architecture Decision Record: ControlCRUD
## Control Statement Authoring and Lifecycle Management System

**Document Version:** 1.0
**Date:** 2026-01-27
**Status:** Approved
**Source Documents:** BRD_ControlCRUD.md, 000_PPRD|ControlCRUD.md

---

## Table of Contents

1. [Decision Summary](#1-decision-summary)
2. [Technology Stack Decisions](#2-technology-stack-decisions)
3. [Development Standards](#3-development-standards)
4. [Architectural Principles](#4-architectural-principles)
5. [Package and Library Standards](#5-package-and-library-standards)
6. [Integration Guidelines](#6-integration-guidelines)
7. [Development Environment](#7-development-environment)
8. [Security Standards](#8-security-standards)
9. [Performance Guidelines](#9-performance-guidelines)
10. [Decision Rationale](#10-decision-rationale)
11. [Implementation Guidelines](#11-implementation-guidelines)
12. [CLAUDE.md Project Standards](#12-claudemd-project-standards)

---

## 1. Decision Summary

### Date and Project Context
- **Decision Date:** 2026-01-27
- **Project:** ControlCRUD - NIST 800-53 Control Statement Authoring System
- **Scale:** Department-level (5-20 concurrent users)
- **Deployment:** Developer instance initially; production TBD
- **Development Approach:** Production-ready from start, phased delivery

### Key Architectural Decisions Overview

| Area | Decision | Rationale |
|------|----------|-----------|
| **Architecture** | Traditional 3-tier | Clear separation, independent scaling, team flexibility |
| **Frontend** | React + TypeScript | Large ecosystem, complex UI support, type safety |
| **Backend** | Go | Performance, simple deployment, excellent for API services |
| **Database** | PostgreSQL | Robust, JSONB support, excellent for structured + semi-structured data |
| **AI Integration** | Direct Claude API | Simple, full control, no abstraction overhead |
| **Testing** | Balanced pyramid | Unit + integration + limited E2E coverage |
| **Auth** | Enterprise SSO (SAML/OIDC) | Per BRD requirement - no standalone auth |

### Decision-Making Criteria
1. **Production Readiness:** Code quality and architecture that doesn't require significant rework
2. **Maintainability:** Clear patterns, strong typing, comprehensive documentation
3. **Performance:** Sub-3-second page loads, responsive UI, efficient API calls
4. **Security:** Enterprise-grade authentication, audit logging, data protection
5. **Developer Experience:** Modern tooling, clear conventions, efficient workflows

---

## 2. Technology Stack Decisions

### 2.1 Frontend Stack

#### Primary Framework: React 18+ with TypeScript
**Rationale:** React's component model excels for complex, interactive UIs like the split-pane editor. TypeScript provides compile-time safety critical for a production-ready approach.

| Component | Technology | Version | Purpose |
|-----------|------------|---------|---------|
| Framework | React | 18.x | UI component library |
| Language | TypeScript | 5.x | Type safety and tooling |
| Build Tool | Vite | 5.x | Fast development builds |
| Routing | React Router | 6.x | Client-side routing |
| State Management | TanStack Query + Zustand | Latest | Server state + client state |
| UI Components | Shadcn/ui + Tailwind CSS | Latest | Accessible, customizable components |
| Rich Text Editor | TipTap | 2.x | ProseMirror-based extensible editor |
| HTTP Client | Axios | 1.x | API communication |
| Form Handling | React Hook Form + Zod | Latest | Forms with validation |

#### UI Component Approach
- **Design System:** Shadcn/ui for accessible, customizable primitives
- **Styling:** Tailwind CSS for utility-first styling with consistent design tokens
- **Icons:** Lucide React for consistent iconography
- **Layout:** CSS Grid and Flexbox for responsive split-pane layouts

#### State Management Solution
- **Server State:** TanStack Query (React Query) for ServiceNow data, caching, and sync state
- **Client State:** Zustand for UI state (editor mode, panel visibility, preferences)
- **Form State:** React Hook Form for complex form management
- **Principle:** Prefer server state; minimize client-side state

#### Build Tools and Development Environment
- **Bundler:** Vite for fast HMR and optimized production builds
- **Package Manager:** pnpm for efficient dependency management
- **Linting:** ESLint with TypeScript rules
- **Formatting:** Prettier with consistent configuration
- **Testing:** Vitest for unit tests, React Testing Library for components

### 2.2 Backend Stack

#### Primary Technology: Go 1.22+
**Rationale:** Go provides excellent performance, simple deployment (single binary), strong concurrency for API orchestration, and a mature ecosystem for web services.

| Component | Technology | Version | Purpose |
|-----------|------------|---------|---------|
| Language | Go | 1.22+ | Core backend language |
| Web Framework | Chi or Echo | Latest | HTTP routing and middleware |
| ORM/Database | sqlc + pgx | Latest | Type-safe SQL, PostgreSQL driver |
| Validation | go-playground/validator | v10 | Request validation |
| Configuration | Viper | Latest | Environment configuration |
| Logging | zerolog | Latest | Structured logging |
| HTTP Client | net/http + retryablehttp | stdlib | ServiceNow API calls |

#### API Design Approach
- **Style:** RESTful API with consistent resource naming
- **Format:** JSON request/response bodies
- **Versioning:** URL path versioning (e.g., `/api/v1/`)
- **Documentation:** OpenAPI 3.0 specification (auto-generated)
- **Error Handling:** Structured error responses with codes and messages

#### Authentication and Authorization Strategy
- **Authentication:** Enterprise SSO via SAML 2.0 or OIDC
- **Session Management:** JWT tokens with refresh mechanism
- **Authorization:** Role-based access control (RBAC)
- **Middleware:** Authentication middleware on all protected routes

| Role | Permissions |
|------|-------------|
| `author` | Create, edit own statements; view all |
| `reviewer` | All author permissions + approve/reject |
| `admin` | All permissions + user management, templates |
| `readonly` | View only (for assessors) |

#### Background Job Processing
- **Approach:** Go routines with worker pool pattern
- **Use Cases:** Batch sync operations, template imports, audit log cleanup
- **Queue:** In-process for MVP; Redis-backed for production scale if needed

### 2.3 Database & Data

#### Primary Database: PostgreSQL 15+
**Rationale:** PostgreSQL's JSONB support handles semi-structured control data, while relational capabilities ensure data integrity for sync state and audit logs.

| Aspect | Decision |
|--------|----------|
| Version | PostgreSQL 15+ |
| Driver | pgx (native Go driver) |
| Query Builder | sqlc (generates type-safe Go from SQL) |
| Migrations | golang-migrate |
| Connection Pool | pgxpool with configurable limits |

#### Data Modeling Approach
```
Core Entities:
├── systems              # Information systems being documented
├── controls             # Control catalog (800-53 reference)
├── control_baselines    # System-specific control selections
├── implementation_statements  # The actual authored content
├── evidence_references  # Links to evidence artifacts
├── templates            # Organizational implementation patterns
├── sync_state           # ServiceNow synchronization tracking
├── audit_log            # All data modifications
└── users                # User profiles and roles
```

**Key Patterns:**
- Use JSONB for flexible metadata and ServiceNow response caching
- Maintain `created_at`, `updated_at`, `created_by`, `updated_by` on all tables
- Soft deletes with `deleted_at` for audit trail preservation
- Optimistic locking with version columns for conflict detection

#### Caching Strategy
- **Application Cache:** In-memory LRU cache for reference data (800-53 catalog, templates)
- **Query Cache:** TanStack Query on frontend with stale-while-revalidate
- **Session Cache:** JWT validation cache to reduce IdP calls
- **No Redis Initially:** Simplify deployment; add if needed for scale

#### Data Migration and Backup Approach
- **Migrations:** Versioned SQL migrations via golang-migrate
- **Backup:** PostgreSQL pg_dump for developer instance
- **Production:** Managed database service with automated backups (TBD)

### 2.4 Infrastructure & Deployment

#### Deployment Platform
- **Development & Production:** Docker Compose containerized deployment
- **Base URL:** `autogrc.mcslab.io`
- **SSL:** TLS 1.2+ with certificate management in nginx/ssl/

#### Container Architecture
```
┌─────────────────────────────────────────────────────────────────┐
│                    Docker Compose Network                        │
│                   (controlcrud-network)                          │
│                                                                  │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐  │
│  │   Nginx     │    │   Backend   │    │     Frontend        │  │
│  │   :80/:443  │───▶│    :8080    │    │       :80           │  │
│  │  (ingress)  │    │   (Go API)  │    │ (React static)      │  │
│  └──────┬──────┘    └──────┬──────┘    └─────────────────────┘  │
│         │                  │                                     │
│         │                  ▼                                     │
│         │           ┌─────────────┐                              │
│         │           │  PostgreSQL │                              │
│         │           │    :5432    │                              │
│         │           │  (database) │                              │
│         │           └─────────────┘                              │
└─────────┴───────────────────────────────────────────────────────┘
```

#### Container Services

| Service | Image | Purpose | Port |
|---------|-------|---------|------|
| `postgres` | postgres:16-alpine | PostgreSQL database | 5432 |
| `backend` | Custom (Go 1.22) | API server | 8080 |
| `frontend` | Custom (Node 20 + Nginx) | React SPA | 80 |
| `nginx` | nginx:alpine | Reverse proxy/ingress | 80, 443 |

#### Container Strategy

**Backend Dockerfile (Multi-stage):**
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

FROM alpine:3.19
RUN addgroup -g 1000 appgroup && adduser -u 1000 -G appgroup -s /bin/sh -D appuser
WORKDIR /app
COPY --from=builder /build/server /app/server
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/server"]
```

**Frontend Dockerfile (Multi-stage):**
```dockerfile
FROM node:20-alpine AS builder
WORKDIR /build
COPY package.json pnpm-lock.yaml* ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY . .
RUN pnpm run build

FROM nginx:alpine
COPY --from=builder /build/dist /usr/share/nginx/html
# SPA routing: try_files $uri $uri/ /index.html
EXPOSE 80
```

#### Nginx Reverse Proxy Configuration

- **HTTP → HTTPS redirect** on port 80
- **API routing:** `/api/*` proxied to backend:8080
- **Frontend routing:** All other routes to frontend (SPA fallback)
- **Rate limiting:** 10 req/s for API, 30 req/s for static content
- **Security headers:** HSTS, X-Frame-Options, X-Content-Type-Options
- **SSL:** Modern TLS configuration (TLSv1.2, TLSv1.3)

#### Database Initialization

Migrations are automatically applied on container startup:
- Migrations mounted to `/docker-entrypoint-initdb.d/` (read-only)
- PostgreSQL executes `.sql` files alphabetically on first start

#### Environment Configuration

```bash
# Required environment variables
POSTGRES_USER=controlcrud
POSTGRES_PASSWORD=<secure-password>
POSTGRES_DB=controlcrud
ENCRYPTION_KEY=<32-byte-base64-key>  # For credential encryption

# Optional configuration
LOG_LEVEL=info
SN_CONNECTION_TIMEOUT=10s
SN_RETRY_MAX=3
CORS_ALLOWED_ORIGINS=https://autogrc.mcslab.io
```

#### Health Checks

| Service | Endpoint | Interval |
|---------|----------|----------|
| PostgreSQL | `pg_isready` | 10s |
| Backend | `GET /health` | 30s |
| Frontend | `GET /health` | 30s |
| Nginx | `nginx -t` | 30s |

#### Docker Compose Commands

```bash
# Start all services
docker compose up -d

# View logs
docker compose logs -f [service]

# Rebuild after code changes
docker compose build [service]
docker compose up -d [service]

# Stop and remove containers
docker compose down

# Reset database (removes volume)
docker compose down -v
```

#### Environment Management
| Environment | Purpose | Configuration |
|-------------|---------|---------------|
| `development` | Local development | `.env.development` |
| `test` | Automated testing | `.env.test` |
| `staging` | Pre-production validation | Environment variables |
| `production` | Live system | Environment variables + secrets management |

#### Monitoring and Logging Strategy
- **Logging:** Structured JSON logs via zerolog
- **Metrics:** Prometheus-compatible metrics endpoint
- **Health Checks:** `/health` and `/ready` endpoints
- **Tracing:** OpenTelemetry instrumentation (prepared, not required for MVP)

---

## 3. Development Standards

### 3.1 Code Organization

#### Directory Structure

**Backend (Go):**
```
backend/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/         # HTTP handlers by domain
│   │   ├── middleware/       # Auth, logging, CORS
│   │   └── routes.go         # Route definitions
│   ├── config/               # Configuration loading
│   ├── domain/               # Business logic and entities
│   │   ├── control/          # Control-related logic
│   │   ├── statement/        # Statement authoring logic
│   │   ├── sync/             # ServiceNow sync logic
│   │   └── user/             # User management
│   ├── infrastructure/
│   │   ├── database/         # Database connection, migrations
│   │   ├── servicenow/       # ServiceNow API client
│   │   ├── claude/           # Claude API client
│   │   └── auth/             # SSO integration
│   └── pkg/                  # Shared utilities
├── migrations/               # SQL migration files
├── api/                      # OpenAPI specifications
└── tests/                    # Integration tests
```

**Frontend (React):**
```
frontend/
├── src/
│   ├── components/
│   │   ├── ui/               # Shadcn/ui primitives
│   │   ├── common/           # Shared components
│   │   ├── editor/           # Statement editor components
│   │   ├── controls/         # Control browser components
│   │   └── layout/           # Layout components
│   ├── features/
│   │   ├── auth/             # Authentication feature
│   │   ├── workspace/        # Control workspace feature
│   │   ├── sync/             # ServiceNow sync feature
│   │   └── templates/        # Template management
│   ├── hooks/                # Custom React hooks
│   ├── lib/                  # Utilities and helpers
│   ├── services/             # API service layer
│   ├── stores/               # Zustand stores
│   ├── types/                # TypeScript type definitions
│   └── App.tsx               # Root component
├── public/                   # Static assets
└── tests/                    # Test files (co-located or here)
```

#### File Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Go files | snake_case | `control_handler.go` |
| Go test files | `_test.go` suffix | `control_handler_test.go` |
| React components | PascalCase | `StatementEditor.tsx` |
| React hooks | camelCase with `use` prefix | `useControlSync.ts` |
| TypeScript types | PascalCase | `ControlStatement.ts` |
| CSS/Tailwind | Component co-location | N/A (Tailwind classes) |
| SQL migrations | Timestamp prefix | `20260127120000_create_controls.up.sql` |

#### Module Organization and Dependency Management

**Go Modules:**
- One `go.mod` at repository root
- Internal packages use `internal/` to prevent external imports
- Domain logic in `internal/domain/` with clear boundaries
- Infrastructure adapters in `internal/infrastructure/`

**Frontend Modules:**
- Feature-based organization in `src/features/`
- Shared components in `src/components/`
- Barrel exports (`index.ts`) for clean imports
- Path aliases configured in `tsconfig.json`

#### Code Style and Formatting Standards

**Go:**
- `gofmt` for formatting (enforced)
- `golangci-lint` with strict configuration
- Effective Go guidelines followed
- Error handling: Always check errors, wrap with context

**TypeScript/React:**
- Prettier for formatting (enforced)
- ESLint with `@typescript-eslint` rules
- Functional components with hooks
- Props interfaces defined explicitly

#### Documentation Requirements

| What | When | Format |
|------|------|--------|
| Package documentation | Every Go package | `doc.go` file |
| Exported functions | All exported items | GoDoc comments |
| Complex logic | Non-obvious code | Inline comments |
| API endpoints | All endpoints | OpenAPI annotations |
| React components | Complex components | JSDoc + Storybook (optional) |
| Architecture decisions | Major decisions | ADR updates |

### 3.2 Quality Assurance

#### Testing Strategy: Balanced Pyramid

```
                    /\
                   /  \     E2E Tests (Critical paths only)
                  /----\    ~10% of test effort
                 /      \
                /--------\   Integration Tests (API, DB)
               /          \  ~30% of test effort
              /------------\
             /              \ Unit Tests (Business logic)
            /----------------\ ~60% of test effort
```

**Coverage Expectations:**
| Layer | Target Coverage | Focus |
|-------|-----------------|-------|
| Domain logic | 80%+ | Business rules, validation |
| API handlers | 70%+ | Request/response, error handling |
| React components | 60%+ | User interactions, state |
| Integration | Key paths | ServiceNow sync, auth flow |
| E2E | Critical flows | Complete authoring workflow |

#### Testing Frameworks

| Layer | Go | React |
|-------|-----|-------|
| Unit | `testing` + `testify` | Vitest + React Testing Library |
| Integration | `testing` + `testcontainers` | Vitest + MSW (Mock Service Worker) |
| E2E | N/A | Playwright |

#### Code Review Process and Standards

**Pull Request Requirements:**
1. All tests passing
2. No linting errors
3. Coverage not decreased
4. At least one approval required
5. Linked to task/issue

**Review Checklist:**
- [ ] Code follows project conventions
- [ ] Error handling is comprehensive
- [ ] Security considerations addressed
- [ ] Performance implications considered
- [ ] Documentation updated if needed

#### Continuous Integration Approach

```yaml
# CI Pipeline Stages
stages:
  - lint        # Format check, lint errors
  - test        # Unit and integration tests
  - build       # Compile and build artifacts
  - security    # Dependency vulnerability scan
```

**CI Tools:**
- GitHub Actions (or GitLab CI)
- Pre-commit hooks for local checks
- Automated dependency updates (Dependabot/Renovate)

### 3.3 Development Workflow

#### Version Control Strategy: GitHub Flow

```
main ─────────────────────────────────────────────────►
       \                    /
        └── feature/F1-xxx ┘   (short-lived feature branches)
```

- `main` branch is always deployable
- Feature branches from `main`, merge back via PR
- Branch naming: `feature/F{number}-{short-description}`
- Commit messages: Conventional Commits format

#### Conventional Commits Format

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `test`: Adding or correcting tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(sync): implement ServiceNow pull operation
fix(editor): resolve cursor position after paste
docs(api): add OpenAPI annotations for control endpoints
```

#### Development Environment Setup

```bash
# Prerequisites
- Go 1.22+
- Node.js 20+ LTS
- pnpm 8+
- PostgreSQL 15+ (or Docker)
- Git

# Setup commands
git clone <repository>
cd controlcrud

# Backend
cd backend
cp .env.example .env.development
go mod download
go run cmd/server/main.go

# Frontend
cd frontend
pnpm install
pnpm dev
```

#### Package Management

**Go:**
- Use `go mod` for dependency management
- Pin major versions in `go.mod`
- Regular `go mod tidy` to clean unused deps
- Security: `govulncheck` for vulnerability scanning

**Frontend:**
- Use `pnpm` for fast, disk-efficient installs
- Lock file (`pnpm-lock.yaml`) committed
- Regular updates via Renovate or manual review
- Security: `pnpm audit` for vulnerability scanning

---

## 4. Architectural Principles

### 4.1 Core Design Principles

1. **Separation of Concerns**
   - Frontend handles presentation and user interaction
   - Backend handles business logic and external integrations
   - Database handles persistence and data integrity

2. **Domain-Driven Design (Lite)**
   - Organize code around business domains (controls, statements, sync)
   - Clear boundaries between domains
   - Shared types only in dedicated packages

3. **Dependency Injection**
   - Interfaces for external dependencies (database, APIs)
   - Constructor injection for testability
   - No global state or singletons for business logic

4. **Fail Fast, Recover Gracefully**
   - Validate inputs at boundaries
   - Return meaningful errors with context
   - Graceful degradation when external services fail

5. **Explicit Over Implicit**
   - Explicit error handling (no panic for recoverable errors)
   - Explicit type definitions (no `any` in TypeScript)
   - Explicit configuration (no magic defaults)

### 4.2 Scalability Considerations

**Current Scale:** 5-20 concurrent users

**Scalability Path:**
1. **Vertical:** Increase server resources (sufficient for MVP)
2. **Horizontal:** Stateless backend enables multiple instances
3. **Database:** Connection pooling, read replicas if needed
4. **Caching:** Add Redis layer for session/reference data

**Design for Scale:**
- Stateless backend (no in-memory session state)
- Database connection pooling
- Efficient bulk operations for sync
- Pagination for large result sets

### 4.3 Security Requirements

| Requirement | Implementation |
|-------------|----------------|
| Authentication | Enterprise SSO (SAML 2.0 / OIDC) |
| Authorization | RBAC with role middleware |
| Data in Transit | TLS 1.3 required |
| Data at Rest | PostgreSQL encryption, AES-256 |
| Audit Logging | All mutations logged with user attribution |
| Input Validation | Server-side validation on all inputs |
| CORS | Strict origin configuration |
| Rate Limiting | Per-user rate limits on API endpoints |

### 4.4 Maintainability Standards

1. **Code Readability**
   - Self-documenting code preferred over comments
   - Comments explain "why", not "what"
   - Consistent naming conventions

2. **Modularity**
   - Features can be developed independently
   - Clear interfaces between modules
   - Minimal coupling between domains

3. **Testability**
   - Code designed for testing from the start
   - Dependencies injected, not hardcoded
   - Test utilities and factories provided

---

## 5. Package and Library Standards

### 5.1 Approved Libraries

#### Go Backend

| Category | Library | Purpose |
|----------|---------|---------|
| HTTP Router | chi or echo | Routing, middleware |
| Database | pgx, sqlc | PostgreSQL driver, type-safe SQL |
| Validation | go-playground/validator | Request validation |
| Configuration | viper | Environment and file config |
| Logging | zerolog | Structured JSON logging |
| Testing | testify, testcontainers | Assertions, integration tests |
| Auth | coreos/go-oidc, crewjam/saml | SSO integration |
| HTTP Client | hashicorp/go-retryablehttp | Resilient HTTP calls |

#### React Frontend

| Category | Library | Purpose |
|----------|---------|---------|
| UI Framework | React 18 | Core UI library |
| Type Safety | TypeScript 5 | Static typing |
| Routing | React Router 6 | Client-side routing |
| Server State | TanStack Query | Data fetching, caching |
| Client State | Zustand | Lightweight state management |
| Forms | React Hook Form + Zod | Form handling, validation |
| UI Components | Shadcn/ui | Accessible component primitives |
| Styling | Tailwind CSS | Utility-first CSS |
| Rich Text | TipTap | Extensible editor |
| HTTP | Axios | API client |
| Testing | Vitest, RTL, Playwright | Unit, component, E2E |

### 5.2 Package Selection Criteria

When adding new dependencies:

1. **Necessity:** Can this be done with existing tools or stdlib?
2. **Maintenance:** Is the package actively maintained? (Last commit < 6 months)
3. **Security:** Any known vulnerabilities? (Check advisories)
4. **Size:** What's the bundle size impact? (Frontend)
5. **License:** Compatible with project license?
6. **Community:** Reasonable adoption and documentation?

### 5.3 Version Management

- Pin to specific major versions
- Update minor/patch versions regularly
- Major version updates require review and testing
- Use Renovate or Dependabot for automated PRs

### 5.4 Custom vs. Third-Party Guidelines

**Build Custom When:**
- Core business logic (control authoring, sync)
- Simple utilities (< 50 lines of code)
- Security-critical components where control is essential

**Use Third-Party When:**
- Complex, well-solved problems (auth, rich text editing)
- Significant development time savings
- Active community and maintenance

---

## 6. Integration Guidelines

### 6.1 API Design Standards

#### URL Structure
```
/api/v1/{resource}              # Collection
/api/v1/{resource}/{id}         # Single resource
/api/v1/{resource}/{id}/{sub}   # Sub-resource
```

#### HTTP Methods
| Method | Purpose | Idempotent |
|--------|---------|------------|
| GET | Retrieve resource(s) | Yes |
| POST | Create resource | No |
| PUT | Replace resource | Yes |
| PATCH | Partial update | Yes |
| DELETE | Remove resource | Yes |

#### Request/Response Format

**Request:**
```json
{
  "data": { /* resource data */ },
  "meta": { /* optional metadata */ }
}
```

**Success Response:**
```json
{
  "data": { /* resource or array */ },
  "meta": {
    "total": 100,
    "page": 1,
    "per_page": 20
  }
}
```

**Error Response:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      { "field": "title", "message": "Title is required" }
    ]
  }
}
```

#### Pagination
- Default page size: 20
- Maximum page size: 100
- Use `?page=1&per_page=20` query parameters
- Return total count in meta

### 6.2 ServiceNow Integration

#### API Client Design
```go
type ServiceNowClient interface {
    // Control operations
    GetControls(ctx context.Context, systemID string) ([]Control, error)
    GetControlStatement(ctx context.Context, controlID string) (*Statement, error)
    UpdateControlStatement(ctx context.Context, id string, stmt *Statement) error

    // Evidence operations
    GetEvidenceReferences(ctx context.Context, controlID string) ([]Evidence, error)

    // Health
    Ping(ctx context.Context) error
}
```

#### Error Handling
- Retry transient failures (5xx, timeouts) with exponential backoff
- Map ServiceNow errors to application error codes
- Log all API interactions for debugging

#### Rate Limiting
- Respect ServiceNow API rate limits
- Implement client-side throttling
- Queue bulk operations with controlled concurrency

### 6.3 Claude API Integration

#### Client Design
```go
type ClaudeClient interface {
    GenerateDraft(ctx context.Context, req DraftRequest) (*DraftResponse, error)
    RefineStatement(ctx context.Context, req RefineRequest) (*RefineResponse, error)
    ValidateCompleteness(ctx context.Context, stmt string, ctrl Control) (*ValidationResult, error)
}
```

#### Prompt Management
- Store prompt templates in configuration/database
- Version prompts for reproducibility
- Include system context (800-53 requirements, organizational style)

#### Error Handling
- Handle rate limits with backoff
- Timeout long-running requests (30s max)
- Fallback gracefully if AI unavailable

### 6.4 Cross-Service Communication Patterns

```
┌──────────────┐     REST/JSON     ┌──────────────┐
│   Frontend   │ ◄───────────────► │   Backend    │
└──────────────┘                   └──────┬───────┘
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    │                     │                     │
                    ▼                     ▼                     ▼
            ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
            │  PostgreSQL  │     │ ServiceNow   │     │  Claude API  │
            │              │     │  REST API    │     │              │
            └──────────────┘     └──────────────┘     └──────────────┘
```

---

## 7. Development Environment

### 7.1 Required Development Tools

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.22+ | Backend development |
| Node.js | 20 LTS | Frontend development |
| pnpm | 8+ | Package management |
| PostgreSQL | 15+ | Database |
| Docker | Latest | Containerization |
| Git | 2.40+ | Version control |
| VS Code / GoLand | Latest | IDE |

### 7.2 IDE Configuration

**VS Code Extensions:**
- Go (official)
- ESLint
- Prettier
- Tailwind CSS IntelliSense
- GitLens
- Thunder Client (API testing)

**Settings (`.vscode/settings.json`):**
```json
{
  "editor.formatOnSave": true,
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  },
  "go.lintTool": "golangci-lint",
  "typescript.preferences.importModuleSpecifier": "relative"
}
```

### 7.3 Local Development Setup

```bash
# 1. Clone repository
git clone <repo-url>
cd controlcrud

# 2. Start database
docker compose up -d postgres

# 3. Backend setup
cd backend
cp .env.example .env.development
go mod download
go run cmd/migrate/main.go up
go run cmd/server/main.go

# 4. Frontend setup (new terminal)
cd frontend
pnpm install
pnpm dev

# 5. Access application
# Frontend: http://localhost:5173
# Backend API: http://localhost:8080
# API Docs: http://localhost:8080/api/docs
```

### 7.4 Testing Environment

**Unit Tests:**
```bash
# Backend
cd backend && go test ./...

# Frontend
cd frontend && pnpm test
```

**Integration Tests:**
```bash
# Requires running database
cd backend && go test -tags=integration ./...
```

**E2E Tests:**
```bash
cd frontend && pnpm test:e2e
```

### 7.5 Debugging Tools

- **Go:** Delve debugger, VS Code debug configuration
- **React:** React Developer Tools, browser DevTools
- **API:** Thunder Client, Postman, curl
- **Database:** pgAdmin, psql CLI
- **Logging:** Structured log viewer (jq for JSON logs)

---

## 8. Security Standards

### 8.1 Authentication Patterns

**SSO Flow (OIDC):**
```
User → Frontend → Backend → IdP
                     ↓
              Token Validation
                     ↓
              JWT Session Token
                     ↓
              Return to Frontend
```

**Session Management:**
- JWT access tokens (15-minute expiry)
- Refresh tokens (7-day expiry, rotated on use)
- Secure, HttpOnly cookies for token storage
- Token revocation on logout

### 8.2 Data Validation and Sanitization

**Input Validation Rules:**
```go
type CreateStatementRequest struct {
    ControlID   string `json:"control_id" validate:"required,uuid"`
    Title       string `json:"title" validate:"required,max=500"`
    Content     string `json:"content" validate:"required,max=50000"`
    Status      string `json:"status" validate:"required,oneof=draft review approved"`
}
```

**Sanitization:**
- HTML content sanitized server-side (allowlist-based)
- SQL parameters always parameterized (sqlc handles this)
- File paths validated against traversal attacks

### 8.3 Secure Coding Practices

1. **Never log sensitive data** (tokens, passwords, PII)
2. **Parameterize all database queries** (use sqlc)
3. **Validate all input** at API boundaries
4. **Escape output** based on context (HTML, JSON)
5. **Use constant-time comparison** for sensitive values
6. **Implement rate limiting** on authentication endpoints
7. **Set security headers** (CSP, HSTS, X-Frame-Options)

### 8.4 Vulnerability Management

- **Dependency Scanning:** Weekly automated scans
- **SAST:** golangci-lint security rules, ESLint security plugin
- **Secrets:** No secrets in code; use environment variables
- **Updates:** Critical vulnerabilities patched within 48 hours

---

## 9. Performance Guidelines

### 9.1 Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| Page load (initial) | < 3s | Lighthouse |
| Page load (cached) | < 1s | Browser timing |
| API response (simple) | < 200ms | P95 latency |
| API response (complex) | < 1s | P95 latency |
| AI generation | < 10s | P95 latency |
| ServiceNow sync (single) | < 5s | P95 latency |

### 9.2 Optimization Strategies

**Frontend:**
- Code splitting by route
- Lazy loading for heavy components (editor)
- Image optimization
- TanStack Query caching with stale-while-revalidate

**Backend:**
- Database connection pooling
- Query optimization (indexes, EXPLAIN ANALYZE)
- Response compression (gzip)
- Efficient JSON serialization

**Database:**
- Appropriate indexes on query patterns
- JSONB indexes for filtered queries
- Prepared statements for repeated queries
- Connection pool sizing based on load

### 9.3 Caching Policies

| Data | Cache Location | TTL | Invalidation |
|------|---------------|-----|--------------|
| 800-53 catalog | Memory + disk | 24h | Manual refresh |
| User session | JWT | 15m | Refresh token |
| Control list | TanStack Query | 5m | On mutation |
| Statement draft | Local storage | N/A | On sync |

### 9.4 Resource Management

- **Database Connections:** Pool size = 10-20 (adjust based on load)
- **HTTP Client:** Connection reuse, reasonable timeouts
- **Memory:** Monitor Go heap, React bundle size
- **File Uploads:** Size limits enforced (evidence references only)

---

## 10. Decision Rationale

### 10.1 Major Decision Trade-offs

#### Go Backend vs. Python/FastAPI

| Factor | Go | Python/FastAPI |
|--------|-----|----------------|
| Performance | Excellent | Good |
| Deployment | Single binary | Requires runtime |
| Type Safety | Compile-time | Runtime (with hints) |
| Ecosystem | Growing | Mature (AI/ML) |
| Learning Curve | Moderate | Lower |
| **Decision:** | **Selected** | Good alternative |

**Rationale:** Production-ready approach favors Go's deployment simplicity, performance, and compile-time safety. The team can be productive with Go's straightforward patterns.

#### PostgreSQL vs. MongoDB

| Factor | PostgreSQL | MongoDB |
|--------|------------|---------|
| Data Integrity | ACID, constraints | Eventually consistent |
| Flexible Schema | JSONB support | Native document model |
| Query Power | SQL, complex joins | Aggregation pipeline |
| Ecosystem | Mature, tooling | Growing |
| **Decision:** | **Selected** | Good for pure document |

**Rationale:** Control data has relational aspects (systems, baselines, statements). PostgreSQL's JSONB handles flexibility while maintaining referential integrity.

#### TanStack Query vs. Redux

| Factor | TanStack Query | Redux |
|--------|---------------|-------|
| Server State | Excellent | Manual |
| Boilerplate | Minimal | Significant |
| Caching | Built-in | Manual |
| Learning Curve | Lower | Higher |
| **Decision:** | **Selected** | Overkill for this use case |

**Rationale:** Most state is server-derived. TanStack Query provides caching, sync, and loading states with minimal code. Zustand handles the small amount of client-only state.

### 10.2 Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Go learning curve | Medium | Medium | Provide Go training resources; follow idiomatic patterns |
| ServiceNow API limitations | Medium | High | Early API exploration; design for fallback |
| AI response quality | Medium | Medium | Human review required; prompt iteration |
| Performance under load | Low | Medium | Load testing before production |

### 10.3 Future Flexibility

**Easy to Change:**
- UI component library (Shadcn is headless primitives)
- AI provider (abstracted behind interface)
- Caching layer (add Redis if needed)

**Harder to Change:**
- Programming language (Go backend)
- Database (PostgreSQL schema)
- Core architecture (3-tier)

**Design Principle:** Abstract external dependencies, own the business logic.

---

## 11. Implementation Guidelines

### 11.1 Applying These Decisions

1. **New Features:** Reference this ADR for technology choices
2. **Code Reviews:** Verify adherence to standards
3. **Exceptions:** Document and discuss deviations in PR
4. **Updates:** Propose ADR amendments for major changes

### 11.2 Exception Process

1. Identify need for deviation
2. Document rationale in PR description
3. Discuss with team lead
4. If approved, add to ADR "Exceptions" section
5. Review exceptions quarterly

### 11.3 Documentation Requirements

| What | Where | When |
|------|-------|------|
| API changes | OpenAPI spec | With implementation |
| New packages | ADR Section 5 | Before adding |
| Architecture changes | ADR amendment | Before implementing |
| Feature designs | Feature TDD | Before coding |

### 11.4 Team Onboarding

New team members should:
1. Read this ADR completely
2. Set up development environment
3. Review CLAUDE.md project standards
4. Complete a small task with pair programming
5. Pass code review on first contribution

---

## 12. CLAUDE.md Project Standards

The following section should be copied into the project's `CLAUDE.md` file to ensure consistent AI-assisted development.

---

```markdown
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
- Explicit error handling with wrapped context: `fmt.Errorf("failed to get control: %w", err)`
- Interfaces for external dependencies (testability)
- Structured logging with zerolog

**TypeScript/React Conventions:**
- Functional components with hooks only
- TanStack Query for all server state
- Zustand for minimal client state (UI preferences, editor mode)
- Zod schemas for runtime validation
- Explicit prop interfaces (no implicit `any`)

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
- Run tests before committing: `go test ./...` and `pnpm test`

**Code Review:**
- All changes via pull request
- One approval required
- Tests passing, no lint errors
- Coverage not decreased

**Commits:**
- Conventional Commits format: `feat(scope): description`
- Reference task number when applicable
- Atomic commits (one logical change per commit)

### Architecture Principles

1. **Separation of Concerns:** Frontend for UI, Backend for logic, Database for persistence
2. **Dependency Injection:** Interfaces for external services, constructor injection
3. **Fail Fast:** Validate at boundaries, return meaningful errors
4. **Explicit Over Implicit:** No magic, clear configuration, typed everything

### Security Requirements

- SSO authentication required (no standalone auth)
- RBAC authorization on all protected endpoints
- Input validation on all API requests
- Audit logging for all data mutations
- No secrets in code (use environment variables)

### ServiceNow Integration

- Use retryable HTTP client with exponential backoff
- Respect API rate limits
- Map ServiceNow errors to application error codes
- Log all API interactions (without sensitive data)

### Claude AI Integration

- Direct API calls (no LangChain abstraction)
- Prompts stored in configuration
- 30-second timeout on generation requests
- Graceful degradation if AI unavailable

### Test Commands

```bash
# Backend
cd backend && go test ./...                    # Unit tests
cd backend && go test -tags=integration ./...  # Integration tests

# Frontend
cd frontend && pnpm test                       # Unit tests
cd frontend && pnpm test:e2e                   # E2E tests

# Full suite
make test                                      # All tests
```

### Commit Message Format

```
<type>(<scope>): <description>

Types: feat, fix, docs, refactor, test, chore
Scopes: sync, editor, api, auth, ui, db

Examples:
feat(sync): implement ServiceNow pull operation
fix(editor): resolve cursor position after paste
test(api): add integration tests for control endpoints
```
```

---

**End of ADR Document**

---

## Document Approval

| Role | Name | Date | Signature |
|------|------|------|-----------|
| Technical Lead | | | |
| Senior Developer | | | |
| Product Owner | | | |

---

**Next Document:** First Feature PRD - ServiceNow Integration (`@0xcc/instruct/003_create-feature-prd.md`)
