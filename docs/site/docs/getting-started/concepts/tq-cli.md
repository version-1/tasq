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

- Keep common issue operations scriptable.
- Return human-readable output by default.
- Support `--output json` for tools and agents.
- Resolve local service URLs without requiring every command to pass an API URL.
- Give agents one stable command surface for issue lookup, comments, and status
  transitions.
- Avoid direct orchestration mutations from issue commands.

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

`tq issue` and `tq comment` operate on issue-tracker data. `tq project`
registers repositories and validates workflow setup. `tq workflow` manages
project workflow overrides. `tq service`, `tq logs`, and `tq migrate` operate
on local runtime state. `tq web` opens the local Web UI after the services are
running.
