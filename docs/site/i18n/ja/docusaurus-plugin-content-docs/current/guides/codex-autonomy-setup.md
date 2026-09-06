---
id: codex-autonomy-setup
title: Codex 自律実行セットアップ
sidebar_position: 3
---

# Codex 自律実行セットアップ

Codex は、リポジトリに作業ルールが明記され、ローカルワークスペースが
trusted になっており、日常的な検証コマンドを中断なしで実行できると動きやすくなります。

このガイドは、Tasq の checkout で長めのローカル Codex 作業を進めるための
準備をまとめます。このセットアップを行ってもセッションが blocked になる場合は、
[blocked になったセッションを復旧する](./recover-blocked-session)を使って、
Tasq の activity から正確な Codex thread を resume してください。

Codex 製品の詳細は、公式 Codex documentation の
[custom instructions](https://developers.openai.com/codex/customization)、
[permissions](https://developers.openai.com/codex/enterprise/permissions)
を参照してください。

## 1. リポジトリ指示を最新に保つ

Codex はセッション開始時に `AGENTS.md` からリポジトリ指示を読みます。このファイルは
実務で使える内容に絞り、簡潔に保ってください。

- 望ましい言語と文体。
- ブランチ、commit、push、PR のルール。
- 必須の検証コマンド。
- ワークフローで使う場合は、許可済みコマンドを安全に実行するためのラッパースクリプト例。
- ドキュメント同期ルール。

ワークフローが `AGENTS.md` に置くには長くなった場合は、詳細を versioned docs に
移し、`AGENTS.md` からリンクしてください。

## 2. Codex Configuration をメンテナンスする

Codex の自律性は、リポジトリ指示だけでなくローカル設定にも依存します。
Tasq の checkout を追加したとき、長く使う worktree を作成したとき、または
リポジトリ外へ書き込む toolchain を導入したときは、`~/.codex/config.toml` を
更新してください。

まず、Codex に操作させたい checkout または worktree を trusted project として
登録します。

```toml
# ~/.codex/config.toml

[projects."/Users/YOU/src/tasq"]
trust_level = "trusted"
```

次に、workspace-write の権限を明示します。古い Codex configuration では
`sandbox_workspace_write.writable_roots` を使い、project、Tasq project の
`.git` directory、日常的な commands が必要とする tool cache directories への
書き込みを許可します。

```toml
# ~/.codex/config.toml

sandbox_mode = "workspace-write"
approval_policy = "on-request"

[sandbox_workspace_write]
writable_roots = [
  "/Users/YOU/src/tasq",
  "/Users/YOU/src/tasq/.git",
  "/Users/YOU/Library/Caches",
  "/Users/YOU/.cache",
  "/Users/YOU/.npm",
  "/Users/YOU/.pnpm-store",
  "/Users/YOU/Library/pnpm/store",
  "/Users/YOU/go/pkg/mod",
  "/Users/YOU/.cargo",
  "/Users/YOU/.gradle",
  "/Users/YOU/.m2",
]
```

現在の Codex permissions では、同じ filesystem rules を複数セッションで再利用
できるように named profile を使うことを推奨します。Setup Guide の
[最小限の Codex 権限設定](pathname:///getting-started/setup-guide#minimal-codex-permissions)
のプロファイルを起点にし、ワークフローが書き込む cache directory ごとに
`workspace_roots` の entry を追加してください。

```toml
# ~/.codex/config.toml

[permissions.tasq_workspace.workspace_roots]
"/Users/YOU/Library/Caches" = true
"/Users/YOU/.cache" = true
"/Users/YOU/.npm" = true
"/Users/YOU/.pnpm-store" = true
"/Users/YOU/Library/pnpm/store" = true
"/Users/YOU/go/pkg/mod" = true
"/Users/YOU/.cargo" = true
"/Users/YOU/.gradle" = true
"/Users/YOU/.m2" = true
```

cache の一覧は、ローカル workflow で使う language、CLI、SDK に合わせて調整します。
よくある例は npm、pnpm、Go、Cargo、Gradle、Maven、Python package caches、
SDK 固有の cache directories です。許可する path は狭く保ち、home directory
全体ではなく必要な cache directory だけを指定してください。

Tasq は agent work に Git worktrees を使います。Tasq に project を追加した場合は、
その project root の `.git` path への書き込みも許可してください。worktree checkout
では Git metadata が root repository を指す file として置かれることがあり、branch、
commit、stash、worktree operations では root の `.git` directory への書き込みが
必要になる場合があります。

## 3. 日常的なローカルコマンドを許可する

リスクの低い読み取りコマンドや検証コマンドが繰り返しの prompt なしで動くと、
Codex はより自律的に進められます。allowlist は狭く保ち、破壊的操作、
履歴の書き換え、認証情報の変更、直接のリモート書き込みは prompt または
明示的なラッパースクリプトの後ろに残してください。

繰り返し使うコマンド判断には Codex rules を使います。現在の `.rules` file format、
rule fields、`codex execpolicy check` については、公式の
[Codex rules documentation](https://developers.openai.com/codex/rules)を参照してください。

Tasq 作業で有用な低リスクのコマンド群は次のとおりです。

- `git status`、`git diff`、`git log`、`git show`。
- `rg`、`sed -n`、`cat`、`find`、`ls`。
- `go test ./...`、`npm run typecheck`、対象を絞った frontend test commands。
- ドキュメント変更用の `make dev-docs-build` または
  `cd docs/site && npm run build`。
- read-only の `gh pr view`、`gh pr diff`、`gh pr checks`。

ラッパースクリプトは任意の例であり、Tasq の必須コマンドではありません。引数を検証し、
副作用を限定するスクリプトをワークフロー側で用意している場合は、生のコマンドを直接
許可するのではなく、そのスクリプト経由で許可済みのリモート書き込みを実行できます。

```sh
~/.codex/bin/safe-git-push
~/.codex/bin/safe-gh-edit pr <number> --body-file <file>
```

ユーザー単位の既定値には `~/.codex/rules/default.rules` を使います。Tasq project
local の既定値として共有する場合は `.codex/rules/default.rules` に置き、Codex が
project `.codex/` layer を読み込めるように project を trusted に保ってください。

```python
# ~/.codex/rules/default.rules or .codex/rules/default.rules

prefix_rule(
    pattern = ["git", ["status", "diff", "log", "show"]],
    decision = "allow",
    justification = "Read-only Git inspection is safe for routine Tasq work.",
    match = [
        "git status --short",
        "git diff -- docs/site",
        "git log --oneline -5",
        "git show --stat",
    ],
)

prefix_rule(
    pattern = [["rg", "find", "ls", "cat"]],
    decision = "allow",
    justification = "Repository inspection commands are safe when scoped by the sandbox.",
    match = [
        "rg TODO docs",
        "find docs/site -maxdepth 2 -type f",
        "ls docs/site",
        "cat AGENTS.md",
    ],
)

prefix_rule(
    pattern = ["npm", "run", ["build", "typecheck", "test"]],
    decision = "allow",
    justification = "Tasq frontend and docs verification commands are expected.",
    match = [
        "npm run build",
        "npm run typecheck",
        "npm run test",
    ],
)

prefix_rule(
    pattern = ["go", "test"],
    decision = "allow",
    justification = "Go tests are routine local verification.",
    match = [
        "go test ./...",
        "go test ./internal/...",
    ],
)

prefix_rule(
    pattern = ["/Users/YOU/.codex/bin/safe-git-push"],
    decision = "prompt",
    justification = "Remote writes should stay explicit even when routed through a wrapper script.",
)
```

rules を編集した後は、Codex を再起動して更新後の file を読み込ませます。rule を
使い始める前に検証する場合は、次の command を実行します。

```sh
codex execpolicy check --pretty \
  --rules ~/.codex/rules/default.rules \
  -- git status --short
```

## 4. タスク目標と作業経路を整備する

大きめのタスクでは、Codex に明確なタスク目標と、作業の進め方が分かる経路の両方を
渡します。プロンプトではその実行で到達すべき成果を定義し、`WORKFLOW.md` では
リポジトリ内でエージェントが繰り返し使う作業ステップを説明します。

タスク目標には次を含めます。

- 対象 branch または PR。
- 対象範囲に含めるファイルやプロダクト領域。
- 変更してはいけないもの。
- 必須のテストやドキュメント build コマンド。
- Codex が commit、push、PR 作成まで行うべきか。

`WORKFLOW.md` では、エージェントの経路を明確にします。

- 編集前に現在の状態を確認する方法。
- 実装ステップとその順序。
- 作業を検証するコマンド。
- 人間の approval が必要な操作。
- 想定した経路が blocked になった場合の対応。

`WORKFLOW.md` は、その内容が適用される作業の近くに置きます。リポジトリ全体の
ワークフローは root に置き、プロダクト領域や package 固有のワークフローは対象
コードの近くに置けます。エージェントが常に参照すべき場合は、`AGENTS.md` から
関連する workflow をリンクしてください。

タスクが長くなる場合は、Codex に goal を保持し、進捗に応じて更新するよう依頼
します。明確なタスク目標と作業経路を組み合わせることで、行き止まりを減らし、
ターミナル接続が切れたときや、セッションの resume が必要になったときにも復旧
しやすくなります。復旧手順は
[blocked になったセッションを復旧する](./recover-blocked-session)を参照してください。

## 5. 自律性を制限範囲内に保つ

自律性は待ち時間を減らすためのもので、レビューを迂回するためのものではありません。
次の操作は明示的な確認対象に残します。

- ファイルやブランチの削除。
- Git history の書き換え。
- 認証情報、global configuration、managed policy の編集。
- prompt、review step、またはコマンドを制限するラッパーを通さない remote push。
- prompt、review step、またはコマンドを制限するラッパーを通さない GitHub issues / PRs の編集。
- trusted ワークスペースの外でのコマンド実行。

Codex が approval を求めた場合は、許可する前にコマンド、作業ディレクトリ、
想定される副作用を確認してください。
