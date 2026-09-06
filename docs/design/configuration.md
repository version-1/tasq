# Configuration

Japanese counterpart: [configuration.ja.md](configuration.ja.md).

For `TQ_HOME`, the directory layout, `config.yaml`, `state.json`, and resolution order, see the [Configuration reference](../site/docs/reference/configuration.md). This document covers only development-specific configuration: build profiles and the Compose dev container.

## Build Profiles

Development binaries can embed a build profile. An empty profile keeps the default home at `~/.tasq`; a profile such as `dev` resolves to `~/.tasq-dev`. Profiles contain only lowercase letters, digits, and hyphens. An explicitly set `TQ_HOME` always overrides the build profile.

The same profile must be embedded in `tq`, `issue-tracker`, `orchestrator`, and `web` so direct service invocation and managed startup use the same runtime state. Profile isolation does not provide concurrent service startup because service ports remain shared.

## Compose Dev Container

The default Compose development workflow runs tools inside the `dev` container with
`TQ_HOME=/workspace/.tasq`. `tq`, TUI, issue-tracker, and orchestrator all read the same runtime
state in that container.

Codex credentials are separate from `TQ_HOME`. The dev container uses
`CODEX_HOME=/home/codex/.codex`, backed by the `codex-home` named volume. Run `make dev-codex-login`
once to authenticate inside the container with device auth. Removing the `codex-home` volume removes
the login state. Device auth avoids browser redirects to a localhost callback that only exists inside
the container.

Repository-managed Codex rules live in `codex/rules/` and are mounted read-only into the dev
container at `/home/codex/.codex/rules`. Authentication, personal overrides, generated approval
decisions, and other secret-bearing Codex state stay in the `codex-home` volume instead of the
repository.
