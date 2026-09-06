---
id: tq-cli
title: tq CLI
sidebar_position: 5
---

# tq CLI

`tq` is the command-line interface for agents, scripts, and developers who need
a stable local interface to Tasq.

It wraps issue-tracker APIs, local service management, project registration,
workflow configuration, migrations, logs, and Web UI launch behavior. In
day-to-day Tasq usage, `tq` is often driven by agents such as Codex or Claude
Code, so keeping it installed on `PATH` and exposing the `tasq-cli` guide helps
agents choose the right command without guessing.

## Design Goals

- Keep common issue operations scriptable, with human-readable output by default and `--output json` for tools and agents.
- Give agents one stable command surface for issue lookup, artifacts, comments, and status
  transitions.
- Avoid direct orchestration mutations from issue commands.

See [Architecture: tq](https://github.com/version-1/tasq/blob/main/docs/design/architecture.md#tq) for the full responsibility list, including API URL resolution order and service management.

## API URL Resolution

```mermaid
flowchart TD
  Start[tq command] --> Flag{--api-url set?}
  Flag -->|yes| UseFlag[Use flag value]
  Flag -->|no| Env{TQ_API_URL set?}
  Env -->|yes| UseEnv[Use env value]
  Env -->|no| State{state.json has tracker addr?}
  State -->|yes| UseState[Use discovered local service]
  State -->|no| Default[Use http://localhost:37651]
```

## Common Surfaces

`tq issue`, `tq artifact`, and `tq comment` operate on issue-tracker data. `tq project`
registers repositories and validates workflow setup. `tq workflow` manages
project workflow overrides. `tq service`, `tq logs`, and `tq migrate` operate
on local runtime state. `tq web` opens the local Web UI after the services are
running.
