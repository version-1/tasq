# Claude Code settings

Use only inside Claude Code, after the shared service prerequisite passes. Comment author: `claude-code`.

- Use Monitor to run the selected CLI's watch command and receive NDJSON notifications. Retain the monitor identifier and handle issue events using `default.md`.
- If Monitor is unavailable, report the missing capability and do not start an unmanaged watcher. Do not apply Codex-specific tool names in Claude Code.
- Use Claude Code's subagent tools for task execution and review. Stop the owned Monitor job on cancellation and verify termination; keep workers and their work unless separately authorized to cancel them.
