# Workflow

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

   - [docs/design.md](docs/design.md)
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

## API Generation

Use `generate:api` for API generation.

## Documentation Updates

When updating documentation, keep the English `.md` file and the Japanese `*.ja.md` file synchronized.

- Update both files for the same content change.
- Add the missing counterpart when only one language file exists.
- Keep links between the English and Japanese versions aligned.
- Do not link Japanese `*.ja.md` counterparts from `AGENTS.md`; link the English `.md` document there.
- Treat ADRs as historical decision records. Do not rewrite an earlier ADR to fit a later decision, except for clearly mechanical fixes such as typos or broken links. When a new decision changes or constrains an earlier one, write the change in a new ADR and describe the relationship there.

## Component Workflows

Use the component-level workflow documents when working in a specific runtime area:

- [Issue Tracker](cmd/issue-tracker/WORKFLOW.md)
- [Orchestrator](cmd/orchestrator/WORKFLOW.md)
- [Web UI](web/WORKFLOW.md)
