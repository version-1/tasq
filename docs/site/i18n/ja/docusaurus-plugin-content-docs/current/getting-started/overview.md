---
id: overview
title: 概要
slug: /
sidebar_position: 1
---

# 概要

Tasq は、Claude Code、Codex などの AI coding agent で複数の実装タスクを並列実行するための task manager です。

タスクごとに独立した workspace を作成し、task registration、agent execution、state management、review、integration までの workflow を支援します。

<video src="/tasq/video/tasq-intro.mp4" controls width="100%">
  <a href="/tasq/video/tasq-intro.mp4">Tasq の紹介動画を見る。</a>
</video>

## Problem

AI coding agents により、複数の実装タスクを同時に進められるようになりました。ボトルネックは code generation そのものから、並列作業の管理へ移ります。

### 人間側の Context Switching

Agent は並列で実行できますが、人間はどのタスクを依頼したか、どの agent が実行中か、各タスクがどこまで進んでいるか、次に何を review すべきかを追う必要があります。

### Workspace Conflicts

複数 agent を 1 つの repository checkout で動かすと、branch switching、未完了変更の衝突、重複した file edits が起きやすくなります。

### 繰り返しの Setup Work

Agent task ごとに、branch 作成、worktree 作成、dependencies の確認、適切な setup commands の実行が必要になりがちです。

## Solution

Tasq は executable tasks を queue として管理し、ready になったタスクに対して agent-ready な workspace を作成します。

![Tasq のエージェント課題フロー](/img/agent-issue-flow.png)

## Tasq が提供するもの

- 実装作業の task management。
- Agent execution のための isolated Git worktrees。
- 1 つの mutable checkout を共有しない parallel agent execution。
- Running work から reviewed output までの review workflow support。
- `requestApproval` 待ちなどで Codex の実行が blocked になっても、Codex CLI から
  `codex resume <session-id>` を使って作業を続行できる復旧経路。詳しくは
  [blocked になったセッションを復旧する](pathname:///guides/recover-blocked-session)を参照してください。
- Local issue tracker、`tq` CLI、Web UI、run history と workspace metadata のための orchestration boundary。

## Goal

Tasq は code generation を速くするためだけのものではありません。parallel AI-agent work によって増える management cost を下げるための tool です。

Task management、workspace isolation、agent execution を 1 つの workflow として扱えるようにします。

## Documentation Map

| ドキュメント | 使う場面 |
| --- | --- |
| [QuickStart](pathname:///getting-started/quickstart) | 動作するサービスまでの最短経路が必要な場合。 |
| [Agent Tutorial](pathname:///getting-started/agent-tutorial) | エージェントが作成した計画を Tasq の課題にし、GitHub pull request まで進めたい場合。 |
| [Setup Guide](pathname:///getting-started/setup-guide) | 繰り返すエージェント作業に向けて、Codex とローカルコマンドの権限を準備する場合。 |
| [Concepts](pathname:///getting-started/concepts/overview) | アーキテクチャを理解したい場合。 |
| [Guides](pathname:///guides/workflow-configuration) | 一般的な操作の手順が必要な場合。 |
| [Reference](pathname:///reference/cli-reference) | コマンド、API、設定、スキーマの詳細を確認したい場合。 |
