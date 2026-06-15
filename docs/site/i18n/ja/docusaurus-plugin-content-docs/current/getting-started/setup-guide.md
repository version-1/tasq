---
id: setup-guide
title: セットアップガイド
sidebar_position: 3
---

# セットアップガイド

この guide では、繰り返しの local agent work を予測可能にするための `tq` 外の setup を扱います。

これらの設定は trusted local projects に scope してください。Git history、remote state、credentials、system configuration を変更できる command に広い global permission を与えることは避けてください。

## Codex 認証

通常の interactive work では ChatGPT で sign in し、Codex が subscription access と active ChatGPT workspace control を使うようにします。

```sh
codex login
```

remote terminals では device authentication を使います。

```sh
codex login --device-auth
```

API key authentication は、ChatGPT subscription access ではなく OpenAI Platform account に usage を請求したい場合にだけ使ってください。

## ローカルプロジェクトを信頼する

Codex は trusted projects でのみ project-local `.codex/` settings を読み込みます。agents が使う checkout または worktree をすべて追加してください。

```toml
# ~/.codex/config.toml

[projects."/Users/YOU/src/tasq"]
trust_level = "trusted"

[projects."/Users/YOU/src/tasq/.worktrees/agents/issue-57"]
trust_level = "trusted"
```

absolute paths を使ってください。agent runner が disposable worktrees を作る場合は、Codex を開始する前に具体的な worktree path を追加します。

## Git メタデータの書き込みを許可する

linked worktrees は Git metadata を workspace directory の外に置くことがよくあります。`git rebase`、`git commit`、`git checkout` などの commands は parent repository の `.git` directory への access が必要になる場合があります。

各 worktree から必要な paths を解決します。

```sh
git rev-parse --path-format=absolute --git-common-dir
git rev-parse --path-format=absolute --git-dir
```

次に、workspace と必要な Git metadata path を Codex permission profile で許可します。

```toml
[permissions.tasq-workspace.workspace_roots]
"/Users/YOU/src/tasq" = true
"/Users/YOU/src/tasq/.worktrees/agents/issue-57" = true
"/Users/YOU/src/tasq/.git" = true
```

exact paths を優先してください。parent `.git` directory 全体を許可すると多くの worktrees で単純になりますが、特定の `.git/worktrees/<name>` path だけを許可すると範囲は狭くなります。

## コマンドルール

routine read と verification commands は狭く許可してください。remote writes と destructive operations は prompts または safe wrappers の背後に置きます。

有用で low-risk な rules は通常、次を対象にします。

- `tq issue get`、`tq issue list`、`tq comment list`。
- `git status`、`git diff`、`git log`、`git show`。
- documentation changes 向けの `make dev-docs-build`。
- read-only の `gh pr view`、`gh pr diff`、`gh pr checks`。

`rm`、`git reset`、`git push`、direct `gh pr edit`、arbitrary arguments を持つ shell wrappers のような広い commands は global に許可しないでください。
