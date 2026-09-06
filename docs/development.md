# Development Workflow

## Local Development Quick Start

Use Docker Compose through `make` for local development:

```sh
make dev-up
```

This starts the `dev` container and OpenAPI UI, then launches the issue-tracker, orchestrator, and Web UI inside the `dev` container. Docker Compose assigns host ports automatically, and the command prints the assigned URLs.

Show the URLs again:

```sh
make dev-ports
```

Stop the environment:

```sh
make dev-down
```

List all available development commands:

```sh
make help
```

### Linux/WSL2 Sandbox Prerequisite

Codex uses Bubblewrap for Linux sandboxing. The dev image installs `bubblewrap`, but Linux and WSL2 hosts must also allow unprivileged user namespace creation for Codex sandboxed commands to work reliably. If Codex reports `bwrap: No permissions to create a new namespace`, treat it as a host or Docker runtime capability issue, not just a missing package in the image.

## Worktree Usage

Create worktrees under sequential directories from `.worktrees/1` to `.worktrees/n` at the repository root, and use them as the working directories for tasks.

When working on multiple tasks at the same time, assign an unused number to each task and keep it separate from existing worktrees.

Example:

```sh
git worktree add .worktrees/1 <branch>
git worktree add .worktrees/2 <branch>
```

Before starting work, check the existing numbers under `.worktrees/` and use the next available number.

## Task Flow

Use this flow from the start of a task to handoff.

1. Confirm the task scope, the expected output, and the files or components likely to be affected.
2. Start the work with `cmd-start-branch` when creating a new task branch.
3. Check the current branch and working tree before editing:

   ```sh
   git status --short --branch
   ```

4. Read the relevant design and workflow documents before changing code or documentation:

   - [docs/design.md](design.md)
   - The component-level workflow document for the area being changed.

5. Make focused changes that match the existing component boundary and ownership.
6. Update related documentation and generated artifacts when the change affects contracts, setup, or developer workflow.
7. Run the narrowest useful verification first, then broaden verification when the change affects shared behavior, contracts, persistence, or user-facing flows.
8. Review the final diff before creating a pull request:

   ```sh
   git diff
   git status --short
   ```

9. Create or update a pull request for the task with `cmd-create-pr`.
10. Handoff with the pull request URL, a concise summary of changed files, verification performed, and any remaining risks or skipped checks.

## GitHub Operations

Use the GitHub CLI (`gh`) for GitHub operations such as viewing pull requests, creating pull requests, and checking pull request status.

## Verification

Run the standard Compose-backed checks:

```sh
make dev-test
```

Run the broader build check before handing off changes that affect both Go services and the Web UI:

```sh
make dev-build
```

## API Generation

Use `generate:api` for frontend API client generation.

Generated Web UI API clients cover these services:

- Issue Tracker: `docs/openapi/issue-tracker.yml` to `cmd/web/frontend/src/lib/generated/issue-tracker.ts`.
- Orchestrator: `docs/openapi/orchestrator.yml` to `cmd/web/frontend/src/lib/generated/orchestrator.ts`.

When changing an API contract, update these artifacts in the same change:

1. The owning service OpenAPI document under `docs/openapi`.
2. The generated Web UI API clients by running `npm run generate:api` from `cmd/web/frontend`.

Also update MSW handlers and fixtures under `cmd/web/frontend/src/mocks` when the changed endpoint is used by standalone frontend development.

## Documentation Updates

When updating documentation, keep the English `.md` file and the Japanese `*.ja.md` file synchronized.

- Update both files for the same content change.
- Add the missing counterpart when only one language file exists.
- Keep links between the English and Japanese versions aligned.
- Do not link Japanese `*.ja.md` counterparts from `AGENTS.md`; link the English `.md` document there.
- Agent instruction files such as `AGENTS.md`, `WORKFLOW.md`, and `cmd/web/frontend/AGENTS.md` are exempt from the pairing rule; keep them as single-language operational files.
- Treat ADRs as historical decision records. Do not rewrite an earlier ADR to fit a later decision, except for clearly mechanical fixes such as typos or broken links. When a new decision changes or constrains an earlier one, write the change in a new ADR and describe the relationship there.

## Repository Documentation

This section is the single index of repository documentation. Do not duplicate this list elsewhere; link to this section instead.

- [WORKFLOW.md](../WORKFLOW.md): Symphony runtime workflow contract used by the orchestrator.
- [docs/design.md](design.md): system architecture and service boundaries.
- [docs/design/deployment.md](design/deployment.md): release tag creation, GitHub Actions, and GoReleaser deployment flow.
- [docs/references/makefile.md](references/makefile.md): Makefile targets, variables, and local development command reference.
- [cmd/issue-tracker/WORKFLOW.md](../cmd/issue-tracker/WORKFLOW.md): issue-tracker development workflow.
- [cmd/orchestrator/WORKFLOW.md](../cmd/orchestrator/WORKFLOW.md): orchestrator development workflow.
- [cmd/web/WORKFLOW.md](../cmd/web/WORKFLOW.md): Web UI development workflow.
- [docs/design/web.md](design/web.md): Web UI structure and styling conventions.
- [docs/openapi/issue-tracker.yml](openapi/issue-tracker.yml): issue-tracker OpenAPI contract.
- [docs/symphony/README.md](symphony/README.md): Symphony documentation index.
- [docs/symphony/SPEC.md](symphony/SPEC.md): Symphony orchestration and runner specification.
- [docs/symphony/DEVIATIONS.md](symphony/DEVIATIONS.md): intentional deviations from the Symphony specification.

## Operational Notes

- Runtime state and SQLite files are created under `.tasq/` in the repository and are ignored by git.
- Compose stores Go module/build caches, `cmd/web/frontend/node_modules`, Codex login state, and GitHub CLI login state in named Docker volumes.
- The orchestrator resolves each project's `WORKFLOW.md` for Symphony-oriented runtime settings and the per-issue agent prompt, with `$TQ_HOME/WORKFLOW.md` as the fallback.
- The Web UI calls local backends through the Go server proxy paths `/tracker/*` and `/orchestrator/*`.
- Run `make dev-codex-login` once to authenticate Codex with device auth and persist credentials in the `codex-home` Docker volume.
- Run `make dev-gh-login` once to authenticate GitHub CLI, configure Git to use `gh` as its HTTPS credential helper, and persist credentials in the `gh-config` Docker volume. Use an HTTPS Git remote for pushes from the dev container.
- Use `make dev-codex-status` and `make dev-gh-status` to confirm the dev container is authenticated before running agent workflows that need Codex or GitHub access.

## tq CLI

The user-facing [CLI Reference](site/docs/reference/cli-reference.md) is the
canonical specification for `tq`. For the one Compose-specific invocation and
its prerequisites, see [tq in the Compose Development Environment](references/tq.md).
