---
id: setup-guide
title: セットアップガイド
sidebar_position: 4
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

Tasq をインストールする手順は [インストール](./install) を参照してください。
コマンドの挙動と API 解決の詳細は [tq CLI](./concepts/tq-cli) を参照してください。

## Project workflow を追加する

repository root に `WORKFLOW.md` を追加し、その project の Tasq タスクでエージェントが
たどる作業手順を明確にします。内容は運用に直結するものに絞ります。issue の確認方法、
変更の進め方、検証コマンド、どの時点で作業を戻すかを書いてください。

例:

```md
# WORKFLOW.md

## Task Intake

1. Run `tq issue get <issue-id>` and read the title, description, comments, and
   attachments before editing files.
2. Run `git status --short` and inspect the current branch.
3. Restate the goal, scope, and verification commands before making changes.

## Implementation

1. Keep changes scoped to the issue.
2. Update tests or documentation when behavior or setup instructions change.
3. Add progress comments with `tq comment add <issue-id> --type progress` when
   the task takes more than one focused step.

## Verification

1. Run the smallest relevant tests first.
2. For docs-site changes, run `npm run build` from `docs/site`.
3. Report any command that could not be run and why.

## Handoff

1. Move the issue to review only after verification passes.
2. Summarize changed files, verification results, and remaining risks.
3. If requested, create a pull request and link it from the issue comments.
```

workflow を repository と一緒に管理する場合は、`WORKFLOW.md` を project に commit
します。machine-local な override が必要な場合は、Tasq に保存します。

```sh
tq workflow add --project tasq-demo --file WORKFLOW.md
tq workflow show --project tasq-demo
```

解決順序と override の挙動は
[Workflow Configuration](../guides/workflow-configuration) を参照してください。

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

<a id="verify-command-permission-coverage"></a>

## コマンド権限のカバレッジを確認する

自律実行の前に、リポジトリの `tasq-permission-check` スキルで、設定した Rules
ファイルに対する Tasq ライフサイクルのコマンド許可を確認します。このスキルは
`codex execpolicy check` を使い、列挙したライフサイクルコマンド自体は実行しないため、
project、issue、commit、pull request を作らずに不足している Rule を見つけられます。

エージェントでまだ利用できない場合は、Tasq リポジトリからスキルをインストールします。

```sh
python "${CODEX_HOME:-$HOME/.codex}/skills/.system/skill-installer/scripts/install-skill-from-github.py" \
  --repo version-1/tasq \
  --path .agents/skills/tasq-permission-check
```

次に、プロジェクトで使う Rules ファイルに対して `tasq-permission-check` を実行するよう
Codex に依頼してください。Tasq checkout 内で直接確認する場合は、次を実行します。

```sh
bash .agents/skills/tasq-permission-check/scripts/check-execpolicy.sh \
  --rules codex/rules/tasq-dev.rules
```

`allow` 以外の結果は権限不足です。該当コマンドに必要な最小限の Rule だけを追加してから、
もう一度確認してください。ライフサイクルコマンドの一覧と報告形式は
[スキルの手順](https://github.com/version-1/tasq/blob/main/.agents/skills/tasq-permission-check/SKILL.md)
を参照してください。
