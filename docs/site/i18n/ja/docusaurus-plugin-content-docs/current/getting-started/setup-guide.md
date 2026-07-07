---
id: setup-guide
title: セットアップガイド
sidebar_position: 3
---

# セットアップガイド

このガイドでは、Codex を使ったエージェント作業を Tasq で繰り返し実行する前に必要な、
`tq` 外の最小ローカルセットアップを扱います。

これらの設定は信頼済みのローカルプロジェクトに限定してください。Git history、
リモート状態、認証情報、システム設定を変更できるコマンドに広い global permission
を与えることは避けてください。

## Codex 認証

Codex 認証は、Codex を使う Tasq エージェント実行の前提です。このガイドでは、
使用するアカウントとワークスペースで Codex CLI の認証が済んでいるものとして扱います。

ログイン方法、device authentication、API-key authentication、認証情報の保存、
CI/CD authentication の詳細は、公式の
[Codex Authentication](https://developers.openai.com/codex/auth) documentation を
参照してください。

## エージェントが迷わず `tq` を使えるようにする

`tq` コマンドは、人間が毎回直接入力するよりも、Codex や Claude Code などの
エージェントから使う場面が多いコマンドです。Tasq をインストールして `tq` コマンドを
`PATH` 上で使えるようにしたうえで、エージェント環境に `tasq-cli` agent guide を
インストールするか、参照できるようにしてください。

`tasq-cli` が使える状態だと、エージェントは代表的な `tq` コマンド、課題の確認、
コメント、状態遷移、サービス操作を小さなリファレンスとして参照できます。これにより、
エージェントがタスクを確認したり、進捗を更新したり、作業を戻す前にメモを残したりする
ときの迷いを減らせます。

Tasq を最短でインストールする手順は [QuickStart](./quickstart) を参照してください。
コマンドの挙動と API 解決の詳細は [tq CLI](./concepts/tq-cli) を参照してください。

## 最小限の Codex 権限設定

自律実行の詳細なセットアップは
[Codex 自律実行セットアップ](../guides/codex-autonomy-setup)にまとめます。この
セットアップガイドでは、Tasq checkout を信頼し、checkout とリポジトリの Git
metadata に workspace-write access を与える最小限の `~/.codex/config.toml` 例だけを
示します。

```toml
# ~/.codex/config.toml

default_permissions = "tasq_workspace"

[projects."/Users/YOU/src/tasq"]
trust_level = "trusted"

[permissions.tasq_workspace]
description = "Tasq checkout with Git metadata writes enabled."
extends = ":workspace"

[permissions.tasq_workspace.filesystem.":workspace_roots"]
"." = "write"

[permissions.tasq_workspace.workspace_roots]
"/Users/YOU/src/tasq" = true
"/Users/YOU/src/tasq/.git" = true
```

ローカル checkout に合わせて、正確な絶対パスを指定してください。追加の worktree、
cache directories、SDK tool directories からエージェントを実行する場合は、その
ワークフローに必要な path だけを追加します。worktree、cache writes、command
rules、復旧経路を含む完全なチェックリストは
[Codex 自律実行セットアップ](../guides/codex-autonomy-setup)を参照してください。
