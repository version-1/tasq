---
id: running-locally
title: Running Locally
sidebar_position: 2
---

# Running Locally

full issue-tracker、orchestrator、Web flow を試す必要がある場合は local services を使います。

## Service Modes

```mermaid
flowchart LR
  Host[Host-only] --> TQ[tq service start]
  TQ --> Tracker[issue-tracker :37651]
  TQ --> Orchestrator[orchestrator :37652]
  TQ --> Web[web :37653]

  Compose[Docker Compose] --> Dev[dev container]
  Dev --> CTracker[issue-tracker :8080]
  Dev --> COrchestrator[orchestrator :8081]
  Dev --> CWeb[web :3000]
```

## Host Commands

```sh
tq migrate
tq service start
tq service status
tq logs issue-tracker -n 200
tq logs orchestrator -n 200
tq logs web -n 200
```

local session が終わったら services を停止します。

```sh
tq service stop
```

## Compose Commands

```sh
make dev-up
make dev-ports
```

Compose では、`make run-web` が Go Web server を起動し、backend URLs は dev container 内の issue-tracker と orchestrator を指します。
