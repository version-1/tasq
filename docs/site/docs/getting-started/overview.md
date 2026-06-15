---
id: overview
title: Overview
slug: /
sidebar_position: 1
---

# Overview

Tasq is a local-first issue tracking and agent orchestration workspace for developers who want task state, automation state, and review handoff to stay close to the repository.

It keeps the human issue workflow simple while giving agents a structured command-line and API surface. A task can move from backlog to review through `tq`, the Web UI, or another client without each tool inventing its own storage format.

## What Tasq Provides

- A local issue tracker backed by SQLite.
- A `tq` CLI for scripts, agents, and terminal workflows.
- A Web UI for scanning issues, changing status, and inspecting task details.
- An orchestrator boundary for run history, workspaces, and future agent execution.
- Repository workflow documents that describe how work should be performed.

## Why Local First

Tasq is designed for private repositories, local agent runners, and work that should not require a hosted tracker before it can be organized. The issue-tracker and orchestrator run on loopback ports, store data under `TQ_HOME`, and expose APIs that local tools can compose.

This keeps setup lightweight and makes it possible to test workflow automation without introducing hosted infrastructure, organization-wide credentials, or external tracker synchronization.

## Documentation Map

Start with [QuickStart](quickstart.md) when you want the shortest path to a running service. Use [Setup Guide](setup-guide.md) when preparing Codex and local command permissions for repeated agent work.

The [Concepts](concepts/overview.md) pages explain the architecture. The [Guides](../guides/workflow-configuration.md) pages cover common operations. The [Reference](../reference/cli-reference.md) pages define commands, APIs, configuration, and schemas.
