# Standards Unification Implementation Plan

> **Note**: This plan was generated using the `writing-plans` superpower skill to execute the alignment of project standards, skills, AGENTS.md, and codebase conventions.

## 1. Context & Objectives

The project had accumulated conflicting rules across three sources:
1. **Generic & Custom Skills** (`.agents/skills/`)
2. **Project Rules** (`AGENTS.md`)
3. **Existing Codebase** (`backend/`, `frontend/`)

Following the user's authoritative directives:
- **Standards are unified with Skills as the master reference.**
- **API Response Envelope adopts the generic skill `api-design` (RESTful standard: status codes 200/201/204/400/404, envelope with `data`, `meta`, `error: {code, message, details}`).**
- **Frontend skills (`frontend-patterns`, `shadcn`, `design-system`) remain untouched; `AGENTS.md` explicitly overrides them with highest priority rules (`bun` + `biome`).**
- **Project-specific skills (`wv-*`) must strictly follow the latest Cordis physical subpackage architecture (`controller/`, `service/`, `dao/`, `model/`, `consts/`, `migrations/`) and correct `Wavelet/...` import paths.**
- **Generic skill `go-logging` is transformed into project-specific skill `wv-logging` (`backend/pkg/logger` + `contracts.LoggerService`).**
- **`wv-database-guide` adds explicit constraint: migrations in business plugins MUST NOT perform DDL on tables owned by other plugins, but ARE PERMITTED to perform DML `INSERT` (e.g. customized downstream plugins injecting seed settings/parameters into shared tables).**

---

## 2. Proposed Changes

### Skills Layer
- **[RENAME & REWRITE]** `.agents/skills/go-logging` -> `.agents/skills/wv-logging/SKILL.md`
- **[MODIFY]** `.agents/skills/wv-new-api/SKILL.md` & `references/handler_example.go`
- **[MODIFY]** `.agents/skills/wv-new-async-task/SKILL.md`
- **[MODIFY]** `.agents/skills/wv-push-notification/SKILL.md`
- **[MODIFY]** `.agents/skills/wv-database-guide/SKILL.md`
- **[MODIFY]** `.agents/skills/README.md`

### Backend Libraries
- **[MODIFY]** `backend/pkg/response/response.go`
- **[MODIFY]** `backend/pkg/response/abort.go`
- **[MODIFY]** `backend/pkg/response/middleware.go`
- **[NEW]** `backend/pkg/response/response_test.go`

### Downstream Baseline Template
- **[MODIFY]** `backend/downstream/plugins/custom_example/plugin.go`
- **[MODIFY]** `backend/downstream/plugins/custom_example/controller/hello/hello.go`
- **[NEW]** `backend/downstream/plugins/custom_example/service/hello.go`
- **[NEW]** `backend/downstream/plugins/custom_example/dao/hello.go`
- **[NEW]** `backend/downstream/plugins/custom_example/model/entity/hello.go`
- **[NEW]** `backend/downstream/plugins/custom_example/migrations/postgres/00001_init.sql`
- **[NEW]** `backend/downstream/plugins/custom_example/migrations/sqlite/00001_init.sql`

### Top-Level Agent Instructions
- **[MODIFY]** `AGENTS.md`

---

## 3. Step-by-Step Execution Plan

### Task 1: Migrate `go-logging` to `wv-logging`
- Step 1: `git mv .agents/skills/go-logging .agents/skills/wv-logging`
- Step 2: Rewrite `wv-logging/SKILL.md` around `backend/pkg/logger`, `GlobalRingBuffer` (5000 lines), OTel integration, and `contracts.LoggerService`.
- Step 3: Update `.agents/skills/README.md` index.

### Task 2: Revise `wv-push-notification` to Cordis Architecture
- Step 1: Remove all outdated `internal/apps/admin/push/` text.
- Step 2: Rewrite to reflect `backend/plugins/domain/msg_gateway/` architecture, `contracts.PushService`, and event bus listeners.

### Task 3: Upgrade `backend/pkg/response` to support `api-design`
- Step 1: Enhance `backend/pkg/response` with standard RESTful envelopes (`DataResponse`, `PagedResponse`, `ErrorResponse`).
- Step 2: Update `abort.go` and `middleware.go` to emit RFC-compliant `{ "error": { "code": "...", "message": "..." } }`.
- Step 3: Add test coverage in `backend/pkg/response/response_test.go` and run `go test ./backend/pkg/response/...`.

### Task 4: Revise `wv-new-api` & `wv-new-async-task`
- Step 1: In `wv-new-api`, replace flat structure with physical subpackages (`controller/`, `service/`, `dao/`, `model/`).
- Step 2: Change response conventions in `wv-new-api` to `api-design` (200/201/204, `{ data, error }`).
- Step 3: In `wv-new-async-task`, update guidance from flat `tasks.go` to subpackage structuring.
- Step 4: Fix all import paths from `github.com/Rain-kl/Wavelet` to `Wavelet`.

### Task 5: Enhance `wv-database-guide` with Cross-Table Migration Rules
- Step 1: Add "Migration Single-Owner DDL Boundary & DML Exception" section:
  - Strictly forbid DDL (CREATE/ALTER/DROP TABLE/COLUMN) across plugins.
  - Explicitly allow DML `INSERT INTO` other tables for downstream customizations needing seed parameters.

### Task 6: Fill `backend/downstream/plugins/custom_example` as Full Subpackage Reference
- Step 1: Implement full `controller -> service -> dao -> model` stack in `custom_example`.
- Step 2: Provide `migrations/postgres/00001_init.sql` and `migrations/sqlite/00001_init.sql`.
- Step 3: Run `go test ./backend/downstream/plugins/custom_example/...`.

### Task 7: Synchronize `AGENTS.md`
- Step 1: Update API response guidelines to `api-design` RESTful standard.
- Step 2: Update skills table mapping `go-logging` to `wv-logging`.
- Step 3: Explicitly record frontend priority rule (`bun` + `biome`).
- Step 4: Record database migration DDL restriction & DML exception.

---

## 4. Verification

- `go test ./...` on affected packages
- `scripts/check_cordis_architecture.sh`
- `make code-check` & `make format`
