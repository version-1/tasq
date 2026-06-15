---
id: overview
title: 概要
slug: /
sidebar_position: 1
---

# 概要

Tasq は、タスクの状態、automation の状態、レビュー引き継ぎをリポジトリの近くに置きたい開発者向けの、local-first な issue tracking と agent orchestration の workspace です。

人間向けの issue workflow はシンプルに保ちながら、agent には構造化された command-line と API surface を提供します。タスクは `tq`、Web UI、または別の client から backlog から review へ進められ、各 tool が独自の storage format を作る必要はありません。

## Tasq が提供するもの

- SQLite を backing store にした local issue tracker。
- scripts、agents、local workflows 向けの `tq` CLI。
- issues の確認、status 変更、task details の inspection に使う Web UI。
- run history、workspaces、将来の agent execution のための orchestrator boundary。
- 作業の進め方を説明する repository workflow documents。

## Local First である理由

Tasq は、private repositories、local agent runners、hosted tracker を用意しなくても整理したい作業向けに設計されています。issue-tracker と orchestrator は loopback ports で動作し、`TQ_HOME` 配下に data を保存し、local tools が合成できる APIs を公開します。

これにより setup は軽量になり、hosted infrastructure、organization-wide credentials、外部 tracker synchronization を導入せずに workflow automation を検証できます。

## Documentation Map

動作する service までの最短経路が必要な場合は [QuickStart](quickstart.md) から始めてください。Codex と local command permissions を繰り返しの agent work 向けに準備する場合は [Setup Guide](setup-guide.md) を使います。

[Concepts](concepts/overview.md) pages は architecture を説明します。[Guides](../guides/workflow-configuration.md) pages は一般的な operation を扱います。[Reference](../reference/cli-reference.md) pages は commands、APIs、configuration、schemas を定義します。
