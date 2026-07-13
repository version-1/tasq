---
id: agent-tutorial
title: エージェントチュートリアル
sidebar_position: 3
---

# エージェントチュートリアル

このチュートリアルでは、小さなリポジトリを登録し、コーディングエージェントに TypeScript と React の TODO アプリを計画させ、その計画から Tasq の課題を作成して、GitHub pull request ができるまで追跡します。

先に[インストール](pathname:///getting-started/install)を完了してください。

## 前提

- `tq` コマンドがインストール済みで、`PATH` から実行できる。
- `tq service start` が実行中である。`tq service status` で確認してください。
- Codex CLI または Claude Code がインストール済みで、GitHub に対して認証されている。どちらも使えない場合は、このチュートリアルの `tq` コマンドを自分で実行できます。
- GitHub CLI の `gh` がインストール済みで、認証されている。`gh auth status` で確認してください。
- Git と GitHub アカウントを利用できる。

## 1. チュートリアル用プロジェクトを登録する

### `tasq-todo` を fork してクローンする

GitHub で [version-1/tasq-todo](https://github.com/version-1/tasq-todo) を fork してから、自分の fork をクローンします。`<your-account>` は自分の GitHub アカウント名に置き換えてください。

```sh
git clone https://github.com/<your-account>/tasq-todo.git
cd tasq-todo
```

自分のリポジトリを使っても構いません。その場合は、Tasq 上で識別しやすく、他のプロジェクトと重複しないプロジェクトキーを選んでください。

### Tasq にプロジェクトを追加する

リポジトリを登録し、ワークフローを検証します。

```sh
tq project add --key tasq-todo .
tq project check tasq-todo
```

`tq project add` は、現在のリポジトリとプロジェクトキーを関連付けます。リポジトリに `WORKFLOW.md` が存在しない場合に限り、既定の `WORKFLOW.md` も作成します。

## 2. `WORKFLOW.md` を理解する

`WORKFLOW.md` は、そのリポジトリで Tasq とエージェントがどのように作業するかを定めるプロジェクト単位の契約です。実行時設定と、エージェントに渡すタスク指示をまとめます。

[tasq-todo の `WORKFLOW.md`](https://github.com/version-1/tasq-todo/blob/main/WORKFLOW.md) には、次の 2 つが含まれています。

- front matter では、ポーリング、`.worktrees/agents` のワークスペースルート、エージェントの制限、Codex app-server のコマンドを設定しています。
- 本文では、すべてのエージェントへ課題のタイトル・説明と必須の流れを渡します。スコープ確認、分離したブランチまたは worktree の作成、変更、検証、pull request 作成、引き継ぎコメント、課題を `review` に移すところまでを定めています。

自分のプロジェクトでは、このファイルを出発点としてコピーし、コマンド、テスト要件、ワークスペース設定をリポジトリに合わせて置き換えてください。ワークフローの解決順序や override の詳細は、[ワークフロー設定](pathname:///guides/workflow-configuration)を参照してください。

## 3. 作業を計画し、課題を作成する

### Codex に `tasq-cli` スキルをインストールする

任意の `tasq-cli` スキルを使うと、Codex はプロジェクト、課題、ワークフロー、サービス、ログのコマンドを絞り込んだリファレンスとして参照できます。一度インストールしたら、Codex を再起動して新しいスキルを読み込ませます。

```sh
python "${CODEX_HOME:-$HOME/.codex}/skills/.system/skill-installer/scripts/install-skill-from-github.py" \
  --repo version-1/tasq \
  --path .agents/skills/tasq-cli
```

Claude Code も代替手段として使えます。この任意スキルがなくても同じ `tq` コマンドを実行できます。下のコマンド指示をプロンプトに含め、必要に応じて [tq CLI リファレンス](pathname:///reference/cli-reference)を参照してください。

### プランモードで計画を作る

`tasq-todo` ディレクトリで Codex をプランモードとして起動します。Claude Code ではプランモードを利用できる場合は選択し、利用できない場合は同じプロンプトで「ファイルを編集しない」と明示してください。

次のプロンプトを送ります。エージェントには、用意した計画を読み、計画をプロンプトへコピーしたり文章だけで課題を説明したりするのではなく、インストール済みの `tq` コマンドを使って Tasq の課題へ変換するよう依頼します。

```md
Read this plan:
https://github.com/version-1/tasq-todo/blob/main/docs/plan.md

Use `tq issue create` to turn the plan into Tasq issues for project key
`tasq-todo`. Create the issues in implementation order, record dependencies
with `--dependency` where needed, and include the relevant plan details in each
issue description. Do not implement the plan. Report every created issue ID
when you finish.
```

実際の作業では、まず Codex または Claude Code に自分のリポジトリを調査させ、計画を作成させてください。その後、承認した計画から課題を作成するよう、このプロンプトを調整します。

課題が作成されたことを確認します。

```sh
tq issue list --project tasq-todo
```

## 4. Web UI で課題を確認し、キューへ入れる

ローカルの Web UI を開きます。

```sh
tq web
```

プロジェクトと課題の一覧から `tasq-todo` を選び、作成した課題を開きます。エージェントに実行させる前に、タイトル、説明、計画、状態、コメント、アクティビティを確認してください。

準備ができたら UI で課題を `ready` に移動します。CLI からも同じ状態遷移を実行できます。

```sh
tq issue ready <issue-id>
```

## 5. エージェントの実行を追跡する

Tasq は、依存関係が解決された `ready` 状態の課題だけを実行します。`--dependency <issue-id>` を指定して作成した課題は、依存先が終端状態になるまで実行キューに入りません。大きな変更を安全な順序の小さな作業に分けられます。

課題詳細ページを開いたままにし、アクティビティとコメントで状態遷移、実行イベント、worktree 情報、エージェントからの引き継ぎを追跡します。

### blocked になった Codex 実行を復旧する

Codex の実行が blocked になった場合は、課題の Activity タブを開き、最新の実行に表示される thread ID をコピーします。リポジトリの checkout から、そのセッションを再開します。

```sh
codex resume <thread-id>
```

承認待ちを減らす自律的な設定は、[Codex 自律実行セットアップ](pathname:///guides/codex-autonomy-setup)を参照してください。`codex resume --last` などの復旧方法は、[blocked セッションの復旧](pathname:///guides/recover-blocked-session)で確認できます。

## 6. pull request を確認する

エージェントが課題を完了したら、課題のアクティビティまたはコメントから GitHub pull request へのリンクを確認します。マージ前に、変更内容とエージェントが報告した検証結果をレビューしてください。

課題が `review` に進まず、`blocked` など想定外の状態になった場合は、まず Web UI のコメントを確認して、エージェントが報告した原因を特定してください。多くの場合、Codex が必要なコマンドを実行する許可を得ていないか、権限が不足しています。再試行する前に、[Codex 自律実行セットアップ](pathname:///guides/codex-autonomy-setup)を参照して、コマンドと権限の設定を見直してください。それでも解決しない場合は、[blocked セッションの復旧](pathname:///guides/recover-blocked-session)が、対象の Codex 実行を続行する手掛かりになることがあります。

Tasq のすべての課題が終わるまで待つ必要はありません。まず 1 つの完了した pull request を自分のプロジェクトへ適用してみてください。複数のプロジェクトと独立したキューを同時に扱うとき、Tasq は特に強力になります。
