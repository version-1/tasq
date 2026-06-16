---
id: overview
title: 概要
slug: /
sidebar_position: 1
---

# 概要

Tasq は、Claude Code、Codex、および同種のエージェントで複数の実装タスクを並列実行するための AI コーディングエージェント向けタスクマネージャーです。

各タスクに分離された workspace を作成し、タスク登録、エージェント実行、状態管理、レビュー、統合までの workflow を支援します。

## 問題

AI コーディングエージェントによって、複数の実装タスクを同時に進められるようになります。ボトルネックはコードを書くことから、並列作業を管理することへ移ります。

### 人間側のコンテキストスイッチ

エージェントは並列実行できますが、人間はどのタスクを割り当てたか、どのエージェントが実行中か、各タスクがどこまで進んだか、次に何をレビューすべきかを追跡する必要があります。

### Workspace の競合

1 つの repository checkout で複数のエージェントを動かすと、ブランチ切り替えの問題、未完了変更の衝突、重複したファイル編集が発生する可能性があります。

### 繰り返しのセットアップ作業

各エージェントタスクでは、ブランチを作成する、worktree を作成する、依存関係を確認する、適切なセットアップコマンドを実行する、という同じ準備手順がしばしば必要です。

## 解決策

Tasq は実行可能なタスクを queue として管理し、実行準備ができたタスクのためにエージェントがすぐ使える workspace を作成します。

![Tasq のタスクキューから並列エージェント workspace への流れ](/img/agent-task-queue.svg)

## Tasq が提供するもの

- 実装作業のタスク管理。
- エージェント実行用に分離された Git worktree。
- 1 つの変更可能な checkout を共有しない並列エージェント実行。
- 実行中の作業からレビュー済みの成果物までを扱うレビュー workflow。
- run history と workspace metadata のための local issue tracker、`tq` CLI、Web UI、orchestration boundary。

## 目標

Tasq は単にコード生成を速くするためのものではありません。並列 AI エージェント作業によって生じる管理コストを減らすためのものです。

開発者に、タスク管理、workspace 分離、エージェント実行のための 1 つの workflow を提供します。

## ドキュメントマップ

サービスを最短で起動したい場合は [クイックスタート](pathname:///getting-started/quickstart) から始めてください。Codex と local command permissions を繰り返しのエージェント作業に向けて準備する場合は [セットアップガイド](pathname:///getting-started/setup-guide) を使ってください。

[概念](pathname:///getting-started/concepts/overview)ページは architecture を説明します。[ガイド](pathname:///guides/workflow-configuration)ページは一般的な操作を扱います。[リファレンス](pathname:///reference/cli-reference)ページはコマンド、API、設定、schema を定義します。
