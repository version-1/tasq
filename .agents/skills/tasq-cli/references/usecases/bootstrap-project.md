# Bootstrap a project

Register a repository with tasq end-to-end: API record, workflow file, gitignore entry, and validation.

## Prerequisite

`tq service status` shows the issue-tracker as running. If not, see [operate-services.md](operate-services.md) first.

## Steps

```bash
# 1. From the repo root, register the project. --key defaults to the kebab-case
#    of the current directory name; pass --key explicitly when the directory
#    name does not match what other tools call the project.
tq project add --key my-app

# 2. (Optional) Replace the default WORKFLOW.md with a custom override held
#    in the API. Front matter is required.
tq workflow add --project my-app --file ./WORKFLOW.md

# 3. Verify everything (WORKFLOW.md present, front matter valid, API accepts
#    the workflow, AGENTS.md present). Exits non-zero if any check fails.
tq project check
```

## What `tq project add` does locally

- Creates `WORKFLOW.md` from the default template if it does not exist.
- Appends `.worktrees` to `.gitignore` (creating the file if absent).
- Rolls back both file changes if the API call fails.

## Verifying without registering

If the project is already registered and you only want to re-run validation:

```bash
# Infer the project from the current directory.
tq project check

# Or check a specific project by key.
tq project check my-app
```

## Removing

```bash
tq project remove my-app
```

This clears the API record but does not touch local files (`WORKFLOW.md`, `.gitignore`).

## See also

- [resources/project.md](../resources/project.md)
- [resources/workflow.md](../resources/workflow.md)
