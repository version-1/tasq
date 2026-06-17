---
name: tq-orchestrator
description: Poll ready issues from the tq issue tracker with the Monitor tool and delegate them to subagents
---

# Constraints

This skill is intended to run inside a Claude Code agent.
If it is run outside of Claude Code, return an error.

# Purpose

Provide the orchestrator layer for tasq: poll ready issues from the issue tracker and delegate them to subagents.

1. Use the Monitor tool to watch the issue tracker, detect ready issues, and hand them off to subagents.
2. Define the Workflow Template.

# Prohibitions

- The agent that runs this skill must not resolve issues itself. Issues must always be delegated to subagents.

# Procedure

1. Ask the user for the polling interval (default: 30 seconds).
2. Use the **Monitor tool** to run `./scripts/poller.sh` and fetch ready issues from the tq issue tracker.
3. When a ready issue exists, `scripts/poller.sh` emits a notification JSON.
4. When you receive the JSON output, extract the issue id from it and hand the work off to subagents based on the Workflow Template.
  - Once the hand-off to the subagent is complete, change the issue status to in_progress with `tq issue update <issue_id> --status in_progress`.
  - Record the subagent name as a comment on the issue with `tq comment add <issue_id> --author claude-code --body "<comment>"`.

# Workflow Template

1. Fetch the issue with `tq issue get <issue_id>`.
2. Confirm the task scope from the issue title and description above.
3. Before branch creation, run `git fetch origin`, then create the task branch from `origin/main` or the repository's default branch.
4. Create or switch to an isolated task branch and worktree before editing.
  - Use `git worktree` to create a working worktree at `<project_root>/.worktrees/tasq/<issue_id>-<issue-summary-title>` and start the work there.
  - Record the created worktree/branch name as a comment on the issue with `tq comment add <issue_id> --author claude-code --body "<comment>"`.
5. Make focused changes that satisfy the issue.
6. Run the narrowest useful verification first, then broaden checks when shared behavior is affected.
7. Commit the change and create or update a pull request.
  - Record the created pull request title and URL as a comment on the issue with `tq comment add <issue_id> --author claude-code --body "<comment>"`.
8. If there is anything to hand off, such as unresolved items, leave a comment with `tq comment add <issue_id> --author claude-code --body "<comment>"`.
9. When the work is complete, change the issue status to review with `tq issue update <issue_id> --status review`.
10. When the work is complete, delete the worktree/branch used for the work.
   - Run this only after confirming that the pull request has been created and the status has been changed to review.
   - If those conditions are not met, do not delete the working worktree.

## Handling errors or cases requiring command execution permission

If work cannot continue because of an error during task execution or because command execution permission is required:
  1. Leave a comment on the issue explaining why work cannot continue with `tq comment add <issue_id> --author claude-code --body "<comment>"`.
  2. Change the issue status to blocked with `tq issue update <issue_id> --status blocked`.
