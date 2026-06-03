# ADR-0007: Store Dev Process Logs Under TQ Home

## Context

Tasq local development starts issue-tracker, orchestrator, and Web processes inside the single dev
container described by ADR-0002. The Makefile previously wrote those background process logs under
`.tmp/dev-logs/` in the repository workspace.

That location made the logs easy to find during the initial dev-container migration, but it also
split runtime state across two places. Databases and service discovery state already live under
`$TQ_HOME/system/`, while process logs lived outside `$TQ_HOME`. This made cleanup, inspection, and
runtime-state ownership less consistent.

## Decision

Development service logs are written under:

```text
$TQ_HOME/system/log/
```

The Makefile keeps separate host and container expansions:

- Host-facing log commands follow `$(TQ_HOME)/system/log/*.log`.
- Dev-container process commands write to `${TQ_HOME}/system/log/*.log`.

The log directory is runtime state, not source-controlled project content.

## Alternatives

### Keep Logs Under `.tmp/dev-logs/`

This preserved the original ADR-0002 implementation detail and kept logs visually separate from
Tasq runtime data. It was rejected because service logs are operational runtime state and should
live with the rest of `$TQ_HOME/system/`.

### Store Logs Under `$TQ_HOME/system/data/`

This keeps all runtime artifacts under one subtree, but it mixes append-only process logs with
SQLite databases and attachment data. A separate `system/log/` directory keeps operational logs
easier to inspect and clean up.

### Use Docker Compose Logs Only

This would avoid file log management, but the default dev workflow runs multiple background
processes inside the same dev container. Per-process files remain easier to follow and attach to
bug reports.

## Consequences

`make run-logs` now follows logs from `$TQ_HOME/system/log/`.

For the default repository-local setup, logs appear under `.tasq/system/log/` on the host. Inside the
dev container, the same logical location is `/workspace/.tasq/system/log/`.

Cleaning or archiving Tasq runtime state can include logs, databases, and state discovery files from
one root. Existing local `.tmp/dev-logs/` files may remain on disk, but new dev process output no
longer goes there.

## Notes

This ADR records a later change to the process-log location and intentionally does not rewrite
ADR-0002, which remains the historical decision record for moving to a single dev container.
