# セットアップガイド

このガイドは、`tq` の外側で行う設定のうち、Tasq プロジェクトで agent が作業しやすくなるものをまとめます。

ローカルマシン、Codex profile、agent runner を継続的な Tasq 作業向けに準備するときに使います。設定は信頼できるローカルプロジェクトに限定し、global に広い権限を与えないでください。

## 目的

- Codex が管理対象プロジェクトごとの Git metadata を更新できるようにする。
- 各ローカル Tasq project または worktree を trusted にして、project-local な Codex 設定を読み込めるようにする。
- 日常的なタスクコマンドを繰り返し確認なしで実行できるようにする。
- 対話的なローカル作業では、Codex に ChatGPT subscription access でログインする。

## Codex 認証

通常のローカル開発では、ChatGPT でログインします。これにより Codex は subscription access と、active な ChatGPT workspace の管理設定を使います。

```sh
codex login
```

リモート端末や headless 環境で browser login が使いづらい場合は、device authentication を使います。

```sh
codex login --device-auth
```

通常の対話的な Tasq 作業では、意図して OpenAI Platform account 側の従量課金を使いたい場合を除き、API key authentication は避けます。

## ローカルプロジェクトを trusted にする

Codex は trusted project に対してだけ、project-local な `.codex/` 設定、hooks、rules を読み込みます。agent が使うローカル Tasq checkout または worktree をすべて追加してください。

```toml
# ~/.codex/config.toml

[projects."/Users/YOU/src/tasq"]
trust_level = "trusted"

[projects."/Users/YOU/src/tasq/.worktrees/agents/issue-56"]
trust_level = "trusted"
```

path は絶対パスで書きます。agent runner が一時的な worktree を作る場合は、起動する具体的な worktree path を runner 側で追加します。

## Git metadata への書き込みを許可する

worktree checkout では、Git metadata が workspace directory の外に置かれることがあります。たとえば `git rebase` は、code workspace 自体が writable でも、親 repository の `.git/worktrees/<name>` 配下に書き込む必要があります。

管理対象プロジェクトごとに Git metadata の場所を解決します。

```sh
git rev-parse --path-format=absolute --git-common-dir
git rev-parse --path-format=absolute --git-dir
```

そのうえで、必要な Git metadata path を Codex 設定で許可します。通常の checkout では project の `.git` directory で足りることが多いです。linked worktree では、runner が計算できるなら親 repository の `.git` directory または特定の `.git/worktrees/<name>` directory を含めます。

直接 `workspace-write` で設定する例:

```toml
# ~/.codex/config.toml

sandbox_mode = "workspace-write"

[sandbox_workspace_write]
writable_roots = [
  "/Users/YOU/src/tasq/.git",
  "/Users/YOU/src/tasq/.git/worktrees/issue-56",
]
```

再利用可能な policy として扱いたい場合は、permission profile の方が適しています。

```toml
# ~/.codex/config.toml

default_permissions = "tasq-workspace"

[permissions.tasq-workspace.workspace_roots]
"/Users/YOU/src/tasq" = true
"/Users/YOU/src/tasq/.worktrees/agents/issue-56" = true
"/Users/YOU/src/tasq/.git" = true

[permissions.tasq-workspace.filesystem]
":minimal" = "read"

[permissions.tasq-workspace.filesystem.":workspace_roots"]
"." = "write"
"**/*.env" = "deny"
```

writable location には、広い glob よりも具体的な path を優先してください。複数 agent が 1 つの親 repository を共有する場合、親 `.git` 全体を許可する方が単純ですが範囲は広くなります。`.git/worktrees/<name>` だけを許可する方が狭い一方で、一部の Git 操作では親の ref や log への書き込みが必要になることがあります。

## コマンド rules

通常の Tasq タスクで想定されるコマンドは Codex rules にします。rules は sandbox 外実行に影響するため、`allow` は狭くし、remote state を変更するコマンドは `prompt` にしてください。

user-level の rules file を作ります。

```text
~/.codex/rules/default.rules
```

最初の設定例:

```python
prefix_rule(
    pattern = ["tq", "issue", ["get", "list"]],
    decision = "allow",
    justification = "Reading Tasq issues is part of normal task setup",
)

prefix_rule(
    pattern = ["tq", "comment", "list"],
    decision = "allow",
    justification = "Reading Tasq issue comments is part of normal task setup",
)

prefix_rule(
    pattern = ["make", ["help", "dev-docs-build"]],
    decision = "allow",
    justification = "Repository verification commands for docs-site tasks",
)

prefix_rule(
    pattern = ["git", ["status", "diff", "log", "show", "rev-parse", "merge-base", "branch"]],
    decision = "allow",
    justification = "Read-only Git inspection commands",
)

prefix_rule(
    pattern = ["gh", "pr", ["status", "view", "checks", "diff"]],
    decision = "allow",
    justification = "Read-only GitHub PR inspection commands",
)

prefix_rule(
    pattern = ["~/.codex/bin/safe-git-push"],
    decision = "prompt",
    justification = "Pushing branches changes remote state and should stay explicit",
)

prefix_rule(
    pattern = ["~/.codex/bin/safe-gh-edit"],
    decision = "prompt",
    justification = "Editing GitHub issues or PRs changes remote state and should stay explicit",
)
```

rules を変更したら Codex を再起動します。

使う前に rule の判定を確認します。

```sh
codex execpolicy check --pretty \
  --rules ~/.codex/rules/default.rules \
  -- tq issue get 56
```

## Runner チェックリスト

agent runner が worktree を作るときは、Codex 起動前に次の値を準備します。

1. Workspace path。例: `/Users/YOU/src/tasq/.worktrees/agents/issue-56`
2. `git rev-parse --path-format=absolute --git-common-dir` の結果。
3. `git rev-parse --path-format=absolute --git-dir` の結果。
4. Workspace path に対する trusted project entry。
5. Workspace と、Git command が必要とする Git metadata path への writable access。

これにより、file edit はできるのに Git metadata が workspace 外にあるため `git rebase`、`git checkout`、`git commit` が block される、という失敗を避けられます。

## Global に許可しないもの

次のような破壊的または広すぎるコマンドは global に allow しないでください。

- `rm`
- `git reset`
- `git checkout`
- `git push`
- `gh pr edit`
- `bash -lc` や `zsh -lc` のような shell wrapper

これらは狭い wrapper、明示的な prompt rule、または task ごとの approval で扱います。
