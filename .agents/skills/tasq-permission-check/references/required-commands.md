# Required Tasq Lifecycle Commands

Run each representative command through `codex execpolicy check` in this order.
The command is a policy probe only; do not execute its real side effect.

| Step | Representative command | Why it is needed |
| --- | --- | --- |
| Inspect tracker | `tq project list` | Find registered projects. |
| Inspect issue | `tq issue get 1` | Read the assigned issue. |
| Start work | `tq issue update 1 --status in_progress` | Record that the agent started. |
| Record handoff | `tq comment add 1 --author codex --body "started"` | Leave attributable progress or blocker comments. |
| Finish work | `tq issue update 1 --status review` | Move the issue to review. |
| Inspect repository | `git status --short` | Check the working tree before changing it. |
| Update base | `git fetch origin` | Obtain current remote references. |
| Create workspace | `git worktree add .worktrees/tasq/1-example -b agent/1-example` | Isolate issue work. |
| Stage change | `git add README.md` | Prepare the focused change. |
| Commit change | `git commit -m "Implement issue 1"` | Save the task result. |
| Publish branch | `git push origin HEAD` | Publish a normal task branch. |
| Create PR | `gh pr create --title "Implement issue 1" --body ""` | Create the pull request. |
| Inspect PR | `gh pr checks 1` | Check pull request validation. |
| Close workspace | `git worktree remove .worktrees/tasq/1-example` | Clean up only after the PR and review transition. |
| Remove merged branch | `git branch -d agent/1-example` | Remove the merged local task branch. |
| Update an existing PR (optional) | `gh pr edit 1 --title "Implement issue 1"` | Needed only when the workflow updates an existing PR. |

Treat force pushes, history rewrites, credential changes, broad filesystem
access, and unscoped remote mutations as outside the autonomous baseline. They
should remain forbidden or require explicit user approval.
