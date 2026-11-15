---
name: tq-orchestrator
description: Poll ready issues from the tq issue tracker with Monitor or a managed background session and delegate them to subagents
---

# Purpose

Poll ready issues from the Tasq issue tracker and delegate each issue to a subagent that follows the project's workflow.

# Constraints

- The orchestrator must delegate issue resolution to subagents and must not resolve issues itself.
- A project's `WORKFLOW.md` takes precedence. Use the default delegated workflow below only for behavior that the project workflow does not specify.

# Orchestration

1. Ask the user for the polling interval in seconds (default: 30), unless already specified in the conversation.
2. Use the **Monitor tool** when available; otherwise use the managed background-session procedure below. Run `tq issue watch --interval <seconds>` with the CLI requested by the user (for example, `tqdev`). Use that same CLI and API target for all subsequent issue, comment, and artifact commands.
3. For each `issue-ready` event, read the issue id from `body.id` and fetch the issue with `tq issue get <issue_id>`.
4. Delegate the issue to a subagent using the project `WORKFLOW.md` and fill only its unspecified behavior from the default delegated workflow below.
5. After a successful hand-off:
   - Change the issue status with `tq issue update <issue_id> --status in_progress`.
   - Record the subagent name with `tq comment add <issue_id> --author claude-code --body "<comment>"`.

# Monitoring without Monitor

The absence of Monitor does not prevent execution when a managed background command session is available.

1. Start the watch command once through a command-execution tool that returns a persistent session ID while the process is running. In Codex, use `exec_command` with a short `yield_time_ms`; retain its returned `session_id`. Use `tty: true` if needed to send Ctrl-C. Redirect stdout and stderr to separate files in a unique temporary spool directory before starting the process, and retain that directory with the session ID. This command intentionally runs continuously. Do not detach it with `nohup`, an untracked shell `&`, or a discarded asynchronous tool call.
2. Use `write_stdin` to check that session and read events from the stdout spool with a retained byte offset. Bound each wait to at most 30 seconds so user messages, worker completions, and stop requests can be handled. An empty read or an empty queue is not completion: continue watching while checking delegated workers. If no managed session tool is available, report the missing capability; do not claim monitoring has started.
3. Parse complete NDJSON lines directly from the spool file with a byte-offset reader, retaining incomplete trailing bytes for the next read. Do not parse the command tool’s displayed output as the event stream: it may be truncated or mix PTY stdout/stderr. Have the reader return bounded batches containing only event type and issue ID (fetch the full issue later), plus bounded error/info summaries, and advance the offset only after that batch is received intact. On a truncated reader response, retry from the previous offset with a smaller batch; on malformed complete lines, report the error and reconcile against the dependency-aware queue before continuing. Keep stderr in its separate spool file. Dispatch only `type=event` / `eventType=issue-ready` events; report `error` envelopes and allow the watch process to retry. `info` envelopes are informational.
4. Before delegation, fetch the current issue and verify it is still ready and eligible (including dependencies). Track only currently running or pending issue IDs locally to suppress duplicate events across reads and restarts. Release an ID after its worker result and tracker updates have been handled, or when a pending issue is no longer eligible. A later ready event for that ID may be dispatched again after checking current status and the absence of an active worker. Respect available subagent capacity; retain pending IDs and recheck eligibility before dispatch when a slot opens. Do not turn backlog issues into ready issues automatically.
5. Process worker results and apply the hand-off workflow while monitoring continues. Use an author matching the executing agent (`codex` in Codex, `claude-code` in Claude Code) for agent-created comments, replacing the examples below as appropriate.
6. If the watch process exits unexpectedly, report the exit and restart only while monitoring remains requested and the cause is recoverable. Preserve the running/pending ID set, recheck current issue state, and avoid a tight retry loop. On permission or configuration failures, report the blocker instead of repeatedly restarting. When setting `--seen-ttl`, keep it strictly greater than the polling interval.
7. On `stop` or an equivalent cancellation, stop dispatching immediately, send Ctrl-C to the owned watch session, and verify it exits. Retain spool files until all received events are accounted for, and report their paths if needed for recovery. Report any still-running workers and their issues/worktrees; stopping the watch does not imply authorization to cancel workers or delete their work. Do not leave an unmanaged watcher running, and do not end the monitoring turn while claiming that an unattended session will keep processing events.

# Default Delegated Workflow

1. Confirm the scope from the issue title and description.
2. Before editing, create an isolated task branch and worktree from the latest default branch:
   - Run `git fetch origin`.
   - Create the task branch from `origin/main` or the repository's default branch.
   - Create the worktree at `<project_root>/.worktrees/tasq/<issue_id>-<issue-summary-title>`.
   - Record the worktree and branch names with `tq comment add <issue_id> --author claude-code --body "<comment>"`.
3. Make focused changes and run the narrowest useful verification first, broadening checks when shared behavior is affected.
4. Commit the changes and create or update a pull request.
5. After the pull request is available:
   - Record its title and URL with `tq comment add <issue_id> --author claude-code --body "<comment>"`.
   - Register its URL with `tq artifact set <issue_id> --type pull_request <pr_url>`.
   - Leave a comment for any unresolved items or hand-off notes.
   - Change the issue status with `tq issue update <issue_id> --status review`.
6. Delete the task worktree and branch only after confirming that the pull request exists and the issue is in review. Otherwise, preserve the worktree.

If the task cannot be completed for any reason, including an error or missing command permission:

1. Explain the blocker with `tq comment add <issue_id> --author claude-code --body "<comment>"`.
2. Change the issue status with `tq issue update <issue_id> --status blocked`.
3. Preserve the task worktree for follow-up.

# Output Protocol

> [!NOTE]
> `tq issue watch` is experimental and intentionally isolated from the rest of the CLI.

`tq issue watch` ignores the global `--output` flag and always writes one JSON object per line. With Monitor, each stdout line becomes one notification; with a background session, consume complete lines incrementally as described above.

- `{"type":"error","body":"<message>"}` — a transient polling failure; the loop continues.
- `{"type":"event","eventType":"issue-ready","body":<issue>}` — a ready issue; `body` contains the full issue payload.
- `{"type":"info","body":"<message>"}` — startup configuration or a polling summary; emitted only with `--verbose`.

Flags:

- `--interval <seconds>`: polling interval (default: 30).
- `--seen-ttl <seconds>`: suppress duplicate issue events for this duration (default: 900; must be greater than `--interval`).
- `--verbose`: emit info envelopes.

The loop runs until stopped by SIGINT, kill, or the monitoring tool/session timeout; it does not exit on its own.
