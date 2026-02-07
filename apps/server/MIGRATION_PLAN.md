# Server Package Structure Migration Plan

## Goal

Migrate the server app from layer-first folders (`internal/core`, `internal/repo`, `internal/handlers`) to a domain-first structure with modules under `internal/modules/<domain>` and clear `service/`, `repository/`, and `http/` boundaries, while preserving behavior.

## Target Folder Structure

```text
apps/server
├── cmd/
│   ├── api/main.go
│   ├── worker/main.go
│   ├── migrate/main.go
│   ├── seed/main.go
│   └── metrics/main.go
├── internal/
│   ├── bootstrap/
│   │   ├── api/
│   │   └── architecture/
│   ├── platform/
│   │   ├── auth/
│   │   ├── http/
│   │   └── ...
│   ├── modules/
│   │   └── <domain>/
│   │       ├── service/
│   │       ├── repository/
│   │       │   ├── commands.go   # write paths
│   │       │   ├── queries.go    # read paths
│   │       │   └── models.go
│   │       └── http/
│   ├── migrations/
│   ├── seeding/
│   ├── sse/
│   └── taskhandlers/
└── pkg/
```

## Domain Mapping

Each domain is moved from:

- `internal/core/<domain>` -> `internal/modules/<domain>/service/`
- `internal/repo/<domain>repo` -> `internal/modules/<domain>/repository/`
- `internal/handlers/<domain>grp` -> `internal/modules/<domain>/http/`

Planned domains:

- `activities`, `attachments`, `chatsessions`, `comments`, `documents`, `epics`
- `invitations`, `keyresults`, `labels`, `links`, `notifications`
- `objectives`, `objectivestatus`, `okractivities`, `reports`, `search`
- `sprints`, `states`, `stories`, `subscriptions`, `teams`, `teamsettings`
- `users`, `workflowtemplates`, `workspaces`

Special packages:

- `internal/mux` -> `internal/platform/http/mux`
- `internal/web/mid` -> `internal/platform/http/middleware`
- `internal/handlers/handlers.go` -> `internal/bootstrap/api/routes.go`

## Non-Goals (for this migration)

- No business logic rewrites.
- No schema changes.
- No API contract changes.
- No large algorithmic refactors.

## Safety Strategy

1. Move packages mechanically and keep existing package names where practical.
2. Update imports immediately after each phase.
3. Run tests at every phase gate.
4. Keep migration incremental and reversible.

## Phase Plan

### Phase 0: Baseline and Plan

- Capture baseline tests.
- Create this plan and use it as the single migration reference.

### Phase 1: Platform/App Skeleton Moves

- Move `internal/mux` -> `internal/platform/http/mux`.
- Move `internal/web/mid` -> `internal/platform/http/middleware`.
- Move top-level route registration (`internal/handlers/handlers.go`) -> `internal/bootstrap/api/routes.go`.
- Keep existing route group implementations in place for now.
- Test gate: `go test ./...`.

### Phase 2: Domain Package Moves

- For each domain:
  - Move `core/<domain>` files to `internal/modules/<domain>/service/`.
  - Move `repo/<domain>repo/*` to `internal/modules/<domain>/repository/`.
  - Move `handlers/<domain>grp/*` to `internal/modules/<domain>/http/`.
- Update imports for moved paths.
- Test gate: `go test ./...`.

### Phase 3: Cleanup

- Remove now-empty old directories (`internal/core`, `internal/repo`, `internal/handlers`, `internal/web`).
- Verify no imports reference old paths.
- Test gate: `go test ./...`.

## Validation Checklist

- `go test ./...` passes except known pre-existing failures.
- `go list ./...` succeeds.
- No `internal/core/`, `internal/repo/`, `internal/handlers/`, `internal/web/mid` imports remain.
- API and worker binaries compile.

## Execution Status

- Phase 0 completed: baseline captured and plan established.
- Phase 1 completed: moved `mux`, middleware, and route registration into `internal/platform/http/*` and `internal/bootstrap/api`.
- Phase 2 completed: moved domain code into `internal/<domain>/{service,repository,http}`.
- Phase 3 completed: removed old empty layer folders (`internal/core`, `internal/repo`, `internal/handlers`, `internal/web`).
- Follow-up completed: fixed notification message panic caused by unsafe type assertion for `status_name`.
- DDD hardening completed:
  - legacy package names (`*grp`, `*repo`) replaced with explicit adapter package names (`*http`, `*repository`)
  - transport context access extracted to `internal/platform/auth`
  - domain `service` and `repository` packages no longer depend on transport middleware package
  - architecture guard test added at `internal/bootstrap/architecture/boundaries_test.go`
- Module organization completed:
  - domain packages moved under `internal/modules/<domain>`
  - non-domain packages kept under `internal/{app,platform,seeding,sse,migrations,taskhandlers}`
  - repository methods split into `commands.go` and `queries.go` across module repositories
