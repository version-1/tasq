---
name: tq-orchestrator
description: Monitor ready Tasq issues and delegate them to subagents when the orchestrator service is stopped. Supports Claude Code and Codex.
---

# Purpose

Monitor ready issues and delegate their resolution. The parent manages monitoring and hand-offs; it must not resolve issues itself. The project's `WORKFLOW.md` takes precedence over the fallback task workflow.

# First: exclude a running orchestrator

Before asking for a polling interval, starting a watcher, spawning workers, or changing issues:

1. Select the CLI requested by the user (`tq` by default; `tqdev` when requested). Preserve the same profile, `TQ_HOME`, and API target for every operation. Run `<cli> --output json service status` (the JSON form of `tq service status`).
2. Inspect the `orchestrator` service entry. Continue only when the command succeeds and exactly one entry explicitly reports `state: stopped`. If it reports `running`, this skill is unavailable: report the conflict and return without starting monitoring or delegating work. Do not stop services automatically.
3. If the command fails, the entry is missing/duplicated, or the state is unknown, report that the prerequisite could not be verified and return. Do not interpret missing information as stopped. If the API target is remote or differs from the local service profile, local status is insufficient; do not start until status for that same target can be verified.

Repeat this check before restarting a watcher and before each new delegation. If another orchestrator starts or status becomes unverifiable, stop new dispatches and the owned watcher; preserve and report workers already running. This check reduces conflicts but is not an atomic lock against another agent starting the same skill.

# Select instructions

After the prerequisite passes, read [default.md](references/default.md) for the shared workflow, then only the reference for the executing environment:

- Claude Code: [claude.md](references/claude.md).
- Codex: [codex.md](references/codex.md).
- Other environments: use the shared workflow only if a managed monitor and subagent tools are available; otherwise report the missing capabilities without starting.
