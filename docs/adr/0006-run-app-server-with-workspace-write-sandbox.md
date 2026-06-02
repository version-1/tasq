# ADR-0006: Run App-Server with Workspace-Write Sandbox

## Context

Tasq runs Codex app-server from inside the local dev container. In this topology, the container is
the first isolation boundary for agent execution, credentials, mounted files, and network access.

Codex can also sandbox generated shell commands. In the current dev-container environment, commands
such as `pwd` and `git status` can fail under Bubblewrap with a namespace error:

```text
bwrap: No permissions to create a new namespace
```

That failure causes Codex to request command approval for low-risk repository inspection commands.
Tasq then blocks the issue through the approval workflow. Blocking is correct for unknown or broad
requests, but repeatedly blocking on basic workspace inspection makes local orchestration hard to
use.

## Decision

The repository workflow starts Codex app-server with:

```sh
codex --sandbox workspace-write app-server
```

This makes the app-server use the workspace-write sandbox posture by default for Tasq runs.

Tasq still keeps the approval workflow from ADR-0005. If Codex requests command-execution or
file-change approval, Tasq cancels the request, fails the run with `approval_required`, and blocks a
still-ready issue with the request details.

The dev container remains the primary local-development isolation boundary. The project does not make
the dev container privileged by default just to make Bubblewrap namespace creation work.

Repository-managed Codex rules live in `codex/rules/` and are mounted read-only into
`/home/codex/.codex/rules` inside the dev container. The rest of `CODEX_HOME` remains backed by the
`codex-home` named volume so authentication, personal settings, and future generated approval
decisions are not stored in the repository.

## Alternatives

### Keep the App-Server Command Without an Explicit Sandbox

This leaves sandbox behavior to Codex defaults and local config. It is less explicit and makes local
Tasq behavior harder to reproduce across hosts and containers.

### Make the Dev Container Privileged

This may allow Bubblewrap to create namespaces, but it weakens the container boundary. Since the
container is the primary local isolation layer, making it privileged by default is the wrong tradeoff
for routine development.

### Disable Codex Sandbox Completely

This can reduce approval prompts in an externally sandboxed environment, but it is too broad as the
default posture. It should only be an explicit opt-in profile if needed.

### Rely Only on Codex Rules or Exec Policy

Codex rules and exec policy are still useful for low-risk commands, but they do not replace a clear
runtime sandbox posture. They also may not cover sandbox-escape approvals triggered after namespace
creation fails.

### Mount the Entire Codex Home from the Repository

This would make shared configuration easy to inspect, but it risks mixing authentication, personal
settings, caches, and generated state into source control. Only shared rules are mounted from the
repository.

## Consequences

Local Tasq runs have a documented, reproducible Codex sandbox mode.

Basic workspace write access is available to Codex within the per-issue workspace, while broader
approval requests still become blocked issue work.

The project continues to avoid privileged containers by default. If a developer needs a different
sandbox posture, it should be introduced as an explicit opt-in workflow or profile.

Shared baseline command rules are reviewable in the repository. Because the rules mount is
read-only, Codex cannot persist new approval rules there at runtime. Future generated approval
decisions need a separate structured store or a volume-backed local rules path.

## Notes

This ADR complements ADR-0002's single dev-container topology and ADR-0005's blocked approval
workflow.
