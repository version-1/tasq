---
id: tq-cli
title: tq CLI
sidebar_position: 5
---

# tq CLI

`tq` は、Tasq への安定した local interface を必要とする agents、scripts、developers のための command-line interface です。

issue-tracker APIs、local service management、project registration、workflow configuration、migrations、logs、Web UI launch behavior を wrap します。

## 設計目標

- common issue operations を scriptable に保つ。
- default では human-readable output を返す。
- tools と agents 向けに `--output json` を support する。
- すべての command に API URL を渡さなくても local service URLs を解決する。
- issue commands から direct orchestration mutations を避ける。

## API URL の解決

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

## 主な操作面

`tq issue` と `tq comment` は issue-tracker data に対して動作します。`tq project` は repositories を登録し、workflow setup を validate します。`tq workflow` は project workflow overrides を管理します。`tq service`、`tq logs`、`tq migrate` は local runtime state に対して動作します。
