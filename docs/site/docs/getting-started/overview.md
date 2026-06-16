---
id: overview
title: Overview
slug: /
sidebar_position: 1
---

# Overview

Tasq is an AI coding-agent task manager for running multiple implementation tasks in parallel with Claude Code, Codex, and similar agents.

It creates an isolated workspace for each task and supports the workflow from task registration, agent execution, state management, review, and integration.

## Problem

AI coding agents make it possible to work on multiple implementation tasks at the same time. The bottleneck moves from writing code to managing parallel work.

### Human Context Switching

Agents can run in parallel, but humans still need to track which tasks were assigned, which agents are running, how far each task has progressed, and what should be reviewed next.

### Workspace Conflicts

Running multiple agents in one repository checkout can create branch switching issues, unfinished-change conflicts, and overlapping file edits.

### Repeated Setup Work

Each agent task often needs the same preparation steps: create a branch, create a worktree, verify dependencies, and run the right setup commands.

## Solution

Tasq manages executable tasks as a queue and creates agent-ready workspaces for tasks that are ready to run.

![Tasq task queue to parallel agent workspaces](/img/agent-task-queue.svg)

## What Tasq Provides

- Task management for implementation work.
- Isolated Git worktrees for agent execution.
- Parallel agent execution without sharing one mutable checkout.
- Review workflow support from running work to reviewed output.
- A local issue tracker, `tq` CLI, Web UI, and orchestration boundary for run history and workspace metadata.

## Goal

Tasq is not only about faster code generation. It is about reducing the management cost introduced by parallel AI-agent work.

It gives developers one workflow for task management, workspace isolation, and agent execution.

## Documentation Map

Start with [QuickStart](pathname:///getting-started/quickstart) when you want the shortest path to a running service. Use [Setup Guide](pathname:///getting-started/setup-guide) when preparing Codex and local command permissions for repeated agent work.

The [Concepts](pathname:///getting-started/concepts/overview) pages explain the architecture. The [Guides](pathname:///guides/workflow-configuration) pages cover common operations. The [Reference](pathname:///reference/cli-reference) pages define commands, APIs, configuration, and schemas.
