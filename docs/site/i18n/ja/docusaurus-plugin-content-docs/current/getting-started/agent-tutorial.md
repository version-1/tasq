---
id: agent-tutorial
title: エージェントチュートリアル
sidebar_position: 3
---

# エージェントチュートリアル

このチュートリアルは、[クイックスタート](pathname:///getting-started/quickstart) を
完了したあとに実行します。Codex や Claude Code にタスクの計画を作らせ、その計画を
Tasq の issue として登録し、エージェントのキューに流して GitHub pull request が
作成されるところまで確認します。

このチュートリアルでは、次の準備が終わっている前提です。

- `tq` がインストールされ、`PATH` に入っている。
- `tq service start` がすでに実行されている。
- `tq project add` で project が登録されている。
- Codex または Claude Code が local repository と GitHub に対して認証済みである。

## 1. エージェントに plan を作成させる

Tasq に登録した repository で Codex または Claude Code を起動します。issue を作成
する前に、コードベースを確認して短い実装 plan を書くように依頼します。

プロンプト例:

```md
Create a concise implementation plan for adding <task>. Do not edit files yet.
Include the goal, concrete steps, verification commands, and any risks.
```

続行する前に plan を確認します。plan は 1 つの pull request に収まる大きさで、
別のエージェントが迷わず実行できる程度に具体的であるべきです。

## 2. エージェントに Tasq issue を作成させる

plan の内容が正しければ、その plan から Tasq issue を作成するようにエージェントへ
依頼します。

プロンプト例:

```md
Create a Tasq issue for this plan with `tq issue create`. Use the registered
project key, include the plan in the issue description, and report the issue ID.
```

エージェントは次のようなコマンドを実行します。

```sh
tq issue create \
  --project tasq-demo \
  --title "Add <task>" \
  --description "<plan>"
```

## 3. Web UI で issue の内容を確認する

issue を実行可能にする前に、Web UI で title、description、plan が正しいことを確認
します。

インストール済みの `tq` が issue 単位の Web navigation をサポートしている場合は、
次を使います。

```sh
tq issue web <issue-id>
```

そうでない場合は Web UI を開き、対象の issue を選択します。

```sh
tq web
```

## 4. issue を ready にする

issue description が実行できる内容になったら、ready queue に移動します。

```sh
tq issue ready <issue-id>
```

この時点で、issue は orchestrator と agent runner が扱える対象になります。

## 5. エージェントの進行を待つ

Web UI で issue detail page を開いたままにします。エージェントが作業を開始すると、
issue は `ready` から `in_progress` に変わります。

進行状況は Web UI の activity と comments で確認します。run が blocked になった場合は、
Web UI に表示される run context を使って CLI から復旧します。

## 6. GitHub pull request を確認する

エージェントがタスクを完了したら、GitHub pull request が作成され、issue activity
または comments から参照できることを確認します。

pull request を review し、エージェントが報告した verification commands を確認して、
通常の review と merge process に進みます。
