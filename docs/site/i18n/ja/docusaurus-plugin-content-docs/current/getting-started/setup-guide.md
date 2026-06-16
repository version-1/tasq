---
id: setup-guide
title: セットアップガイド
sidebar_position: 3
---

# セットアップガイド

この guide は、繰り返しの local agent work を予測しやすくするために `tq` の外側で行う setup を扱います。

これらの設定は trusted local project に限定してください。Git history、remote state、credential、system configuration を変更できる command に広い global permission を与えることは避けてください。

## Codex 認証

通常の interactive work では ChatGPT で sign in し、Codex が subscription access を使い、active ChatGPT workspace が制御されるようにします。

```sh
codex login
```

remote terminal では device authentication を使います。

```sh
codex login --device-auth
```

ChatGPT subscription access ではなく OpenAI Platform account 経由で usage を請求したい場合にのみ、API key authentication を使ってください。

## ローカルプロジェクトを信頼する

Codex は trusted project に対してのみ project-local `.codex/` settings を読み込みます。エージェントが使う checkout または worktree をすべて追加してください。

```toml
# ~/.codex/config.toml

[projects."/Users/YOU/src/tasq"]
trust_level = "trusted"

[projects."/Users/YOU/src/tasq/.worktrees/agents/issue-57"]
trust_level = "trusted"
```

absolute path を使ってください。agent runner が使い捨て worktree を作成する場合は、Codex を開始する前に具体的な worktree path を追加します。

## Git メタデータの書き込みを許可する

Linked worktree は Git metadata を workspace directory の外側に保持することがよくあります。`git rebase`、`git commit`、`git checkout` などの command は parent repository の `.git` directory への access を必要とする場合があります。

各 worktree から必要な path を解決します。

```sh
git rev-parse --path-format=absolute --git-common-dir
git rev-parse --path-format=absolute --git-dir
```

その後、workspace と必要な Git metadata path を Codex permission profile で許可します。

```toml
[permissions.tasq-workspace.workspace_roots]
"/Users/YOU/src/tasq" = true
"/Users/YOU/src/tasq/.worktrees/agents/issue-57" = true
"/Users/YOU/src/tasq/.git" = true
```

exact path を優先してください。parent `.git` directory 全体を許可する方が多くの worktree では簡単ですが、特定の `.git/worktrees/<name>` path だけを許可する方が狭い権限になります。

## コマンドルール

通常の read command と verification command は狭く許可してください。remote write と destructive operation は prompt または safe wrapper の背後に置きます。

低リスクで有用な rule は通常、次を対象にします。

- `tq issue get`、`tq issue list`、`tq comment list`。
- `git status`、`git diff`、`git log`、`git show`。
- documentation change 向けの `make dev-docs-build`。
- read-only の `gh pr view`、`gh pr diff`、`gh pr checks`。

`rm`、`git reset`、`git push`、直接の `gh pr edit`、任意引数を受け取る shell wrapper のような広い command を global に許可しないでください。
