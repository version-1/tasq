# ADR-0003: Separate Dev Container Command Intent in Makefile Targets

## Context

ADR-0002 moved local development into a single `dev` container. After that migration, the Makefile still carried ambiguous command boundaries. Some targets were host-facing entry points, some were Docker Compose operations, and some ran processes inside the dev container, but their names did not make those roles obvious.

That ambiguity caused surprising behavior. A command intended to enter an existing container or run `tq` could also start, rebuild, wait for, or stop other processes. This made routine commands harder to reason about and made failures look unrelated to the command that the developer actually ran.

The Makefile is still the process-management surface for local development, so the command names and target dependencies need to communicate intent clearly.

## Decision

Separate Makefile targets by command intent:

- `dev-*` targets are host-facing local development entry points.
- `dc-*` targets manage Docker Compose and dev-container operations.
- `run-*` targets run processes or commands inside the already-running dev container.

`dc-shell` and `dc-exec` attach to an existing `dev` container. They do not start, rebuild, or recreate the container. If the container is missing, the command should fail instead of changing environment state.

`run-tq` is scoped to running the `tq` command inside the dev container. It does not start issue-tracker, wait for issue-tracker, stop existing processes, or manage service lifecycle. If issue-tracker is not running, `run-tq` fails with the underlying API connection error.

Service-oriented targets such as `run-orchestrator`, `run-web`, and `run-tui` may ensure issue-tracker readiness because those workflows need the service dependency. That readiness behavior is not shared by `run-tq`.

The Makefile help output documents the prefix taxonomy and groups targets into sections.

## Alternatives

### Keep Compatibility Aliases

Keeping old target names would reduce short-term migration friction, but it would preserve ambiguous command meanings. The default local development surface should make intent visible, so compatibility aliases are not kept.

### Let Every `run-*` Target Ensure Services

This would make some commands more convenient, but it makes simple commands mutate process state. It also hides whether a failure came from the requested command or from service startup. Only targets that actually operate as service workflows should manage readiness.

### Use Compose Service Names Directly

Developers could call `docker compose` directly, but that would reintroduce scattered command knowledge and make local workflows less consistent. The Makefile remains the documented interface.

## Consequences

The command surface is more explicit. Developers can distinguish environment lifecycle commands from commands that only attach to or run inside the existing dev container.

`run-tq` is less magical. It requires issue-tracker to already be running, but it no longer touches unrelated service processes. This makes it safer for diagnostics and avoids accidental restarts.

Some old commands and habits need to migrate to the new prefixes. The Makefile reference documents the new names and examples.

Failures become more direct. If `dc-shell` fails, the dev container is not running. If `run-tq` fails with a connection error, issue-tracker is not reachable from inside the dev container.

## Notes

This ADR refines the Makefile behavior introduced after ADR-0002. It does not replace ADR-0002's single-container topology or ADR-0001's host-local project path model.
