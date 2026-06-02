# ADR-0002: Use a Single Dev Container for Local Development

## Context

Tasq local development previously used separate Compose services for issue-tracker, orchestrator, Web, and Go tooling. That made the runtime topology closer to multiple services, but it also made `localhost`, service names, `TQ_HOME`, and `system/state.json` mean different things depending on whether a command ran on the host or in a container.

This was especially confusing for agent-oriented development. The orchestrator launches `codex app-server`, while `tq`, TUI, Web, issue-tracker, and orchestrator all need a consistent issue-tracker endpoint. Sharing one `TQ_HOME` between host and containers can store an address that works in one network namespace but not another.

Codex also needs an isolation boundary. Running the agent runner directly on the host would make filesystem and credential exposure harder to reason about.

## Decision

Use a single `dev` container for local development, plus the standalone `openapi` documentation UI service.

The `dev` container runs issue-tracker, orchestrator, Web, `tq`, TUI, and Codex CLI in the same container namespace. It uses:

- `TQ_HOME=/workspace/.tasq`
- `CODEX_HOME=/home/codex/.codex`
- a fixed non-root user
- named volumes for Go caches, Web `node_modules`, and Codex credentials

Codex authentication is done by running `codex login` inside the dev container. The resulting credentials are stored in the `codex-home` named volume. The Docker image does not contain credentials, and the default workflow does not mount host Codex credentials.

Process management remains Makefile-based. The Makefile starts issue-tracker, orchestrator, and Web inside the dev container, stores background logs under `.tmp/dev-logs/`, and keeps TUI as an interactive command.

The split Compose services for issue-tracker, orchestrator, Web, and Go tools are removed from the default development Compose file.

## Alternatives

### Keep Split Compose Services

This preserves a clearer service-per-container topology and makes individual logs easy to inspect through Compose. It was rejected for the default dev workflow because it keeps host/container address translation in the critical path and complicates `TQ_HOME` state sharing.

### Run Everything on the Host

This makes `localhost` and host paths natural, but it removes the container isolation boundary around Codex and makes local dependencies harder to reproduce. It remains possible as an advanced manual workflow, but it is not the default.

### Use a Process Manager

Tools such as supervisord, overmind, or foreman could manage multiple processes inside the dev container. This was deferred because Makefile-based process management is sufficient for the first migration and avoids adding another runtime dependency.

## Consequences

Local development uses one network namespace, so `state.json` can point at `127.0.0.1` and be valid for issue-tracker, orchestrator, `tq`, and TUI when they run in the dev container.

The dev container remains an isolation boundary for Codex. The `codex-home` volume must be treated as secret-bearing local state. Removing that volume requires logging in again.

The default `make tq` path now runs inside the dev container. This improves endpoint consistency, but project path handling needs care: commands that persist project paths can see the container workspace path rather than a host-local path unless run through a host-aware workflow. ADR-0001 remains the durable product model for project records.

The Makefile owns process lifecycle. Duplicate process prevention relies on narrow process patterns, and background logs are written under `.tmp/dev-logs/`.

Existing scripts or habits that addressed old Compose service names need to migrate to the dev-container targets.
