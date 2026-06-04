# ADR-0001: Store Project Paths as Host-Local Paths

## Context

Tasq stores projects and workspaces through the issue-tracker API. During local development, the issue-tracker often runs in Docker Compose while `tq` may run either on the host or inside a Compose helper container.

When both `tq` and issue-tracker run inside Compose, the repository is visible as `/workspace`. That path is valid for the containers, but it is not the user's real local project path. Persisting `/workspace` in project records leaks a development-runtime detail into user data and does not scale to multiple local projects unless every project is mounted into the same container filesystem.

For the product model, a project represents a repository on the user's machine. The durable project location should therefore be the host-local absolute path selected by the user, such as `/Users/admin/Projects/Private/tasq`.

## Decision

Persist `Project.Location` and `Workspace.Path` as host-local absolute paths.

The `tq project add <path>` command resolves the path on the client host, checks that it exists locally, and sends the host-local absolute path to the issue-tracker API.

The issue-tracker API validates path shape only. It requires an absolute path and length limits, but it does not check directory existence with the server filesystem. The API server may run in a container or on another runtime that cannot see the client's host path.

Runner and container integrations must translate host-local paths to runtime paths only when they need to execute work in an isolated environment. That mapping belongs to the runner/container adapter boundary, not to the project record.

## Alternatives

### Store Container Paths

Storing `/workspace` works for the single-project Docker Compose development setup because `go-tools` and issue-tracker share the same bind mount.

This was rejected as the durable model because it stores an environment-specific runtime path and prevents registering arbitrary host projects unless they are also mounted into the issue-tracker container.

### Mount Host Paths at the Same Absolute Path

Docker Compose could mount host directories at the same absolute path inside the container, for example `/Users/admin/Projects:/Users/admin/Projects`.

This can make server-side existence checks pass, but it is machine-specific, brittle across operating systems and CI, and still couples issue-tracker validation to a particular runtime layout.

### Run Everything on the Host

Running issue-tracker and `tq` both on the host makes server-side path existence checks natural.

This does not cover the supported Compose development environment and would make API behavior depend on deployment topology.

## Consequences

Project records remain portable across execution backends because they describe the user's project location, not the container's view of it.

The API can accept valid host paths even when it cannot access the host filesystem directly.

Clients that can see the target filesystem are responsible for existence checks before creating records. `tq project add` performs this check.

Runtime components that execute inside containers need an explicit path mapping step, such as mapping `/Users/admin/Projects/Private/tasq` to `/workspace` or mounting each project path into the runner container.

API consumers that bypass `tq` can create records for non-existent paths if they send syntactically valid absolute paths. This is acceptable because issue-tracker cannot reliably validate client-host filesystem state.

## Notes

`make tq` should run `tq` on the host and connect to the Compose issue-tracker through the assigned localhost port. This keeps path resolution aligned with the user's filesystem while still supporting the Compose API service.

The schema reference documents this validation boundary in `docs/design/schema.md`.
