---
id: running-locally
title: Running Locally
sidebar_position: 2
---

# Running Locally

Use local services when you need to exercise the full issue-tracker, orchestrator, and Web flow.

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

Stop services when the local session is done.

```sh
tq service stop
```

## Compose Commands

```sh
make dev-up
make dev-ports
```

In Compose, `make run-web` starts the Go Web server with backend URLs pointing at the issue-tracker and orchestrator inside the dev container.
