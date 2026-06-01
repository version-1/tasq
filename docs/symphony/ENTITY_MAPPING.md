# Tasq Entity Mapping to Symphony SPEC

This document maps Symphony SPEC domain model concepts to Tasq entities. It serves as a lookup
reference for developers working on Tasq, clarifying which Tasq types correspond to which SPEC
concepts and how they relate to each other.

For SPEC deviations and rationale, see [DEVIATIONS.md](DEVIATIONS.md).

## Project

### SPEC Correspondence

The Symphony SPEC does not define a Project entity. The closest concept is `tracker.project_slug`
(SPEC §5.3.1), which scopes issue queries to a single project in the external tracker (Linear).

### Tasq Type

`entity.Project` in `internal/issue/domain/entity/entity.go`

### Ownership

- A Project owns zero or more Workspaces (`entity.Workspace`).
- Issues exist independently of Projects in the current schema.

### Usage

- Managed by the issue-tracker service.
- `Project.Key` (e.g., `TASQ`) identifies the project in CLI and API contexts.
- Not referenced by the orchestrator. The orchestrator does not need project scope because it uses
  issue-tracker issue APIs that already apply the relevant issue filtering.

### Relationships

- `entity.Workspace.ProjectID` → `entity.Project.ID`

## entity.Workspace (Issue Tracker)

### SPEC Correspondence

No direct SPEC counterpart. The SPEC Workspace (§4.1.4) is an orchestrator-side runtime concept.
`entity.Workspace` is a Tasq-specific management record owned by the issue-tracker service.

### Tasq Type

`entity.Workspace` in `internal/issue/domain/entity/entity.go`

### Ownership

- Belongs to one Project (`ProjectID`).

### Usage

- Managed by the issue-tracker service for CRUD operations.
- Stores metadata about a named workspace: name, filesystem path, and lifecycle status
  (`active`, `inactive`, `archived`).
- Not used by the orchestrator when creating or managing per-issue workspaces.

### Relationships

- `entity.Workspace.ProjectID` → `entity.Project.ID`

## workspace.Workspace (Orchestrator)

### SPEC Correspondence

Maps directly to SPEC §4.1.4 Workspace.

| SPEC Field     | Tasq Field     |
|----------------|----------------|
| `path`         | `Path`         |
| `workspace_key`| `WorkspaceKey` |
| `created_now`  | `CreatedNow`   |

### Tasq Type

`workspace.Workspace` in `internal/orchestrator/workspace/workspace.go`

### Ownership

- Created and managed by `workspace.Manager`.
- Independent of `entity.Project` and `entity.Workspace`.

### Usage

- Created per issue by the orchestrator when dispatching agent work.
- Workspace key is derived from `issue-<ID>` (see [DEVIATIONS.md](DEVIATIONS.md#workspace-key)).
- Workspace directory is `<workspace.root>/<sanitized_key>`.
- Lifecycle hooks (`after_create`, `before_run`, `after_run`, `before_remove`) execute in the
  workspace directory. See [WORKFLOW_CONTRACT.md](WORKFLOW_CONTRACT.md) for hook configuration.
- Reused across runs for the same issue. Cleaned up when the issue reaches a terminal state.

### Relationships

- `workspace.Manager` holds the workspace root path and hook configuration from
  `workflow.Config`.
- No foreign key to `entity.Project` or `entity.Workspace`.
