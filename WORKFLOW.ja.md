# Workflow

## Worktree 運用

作業はリポジトリ直下の `.worktrees/1` から `.worktrees/n` までの連番ディレクトリに worktree を作成して進めます。

同時に複数の作業を行う場合は、作業単位ごとに未使用の番号を割り当て、既存の worktree と混在させないでください。

例:

```sh
git worktree add .worktrees/1 <branch>
git worktree add .worktrees/2 <branch>
```

作業開始前に `.worktrees/` 配下の既存番号を確認し、次に空いている番号を使います。

## Task Flow

Task の開始から handoff までは、この flow を使います。

1. Task scope、expected output、影響しそうな files または components を確認します。
2. 新しい task branch を作成する作業では、`cmd-start-branch` で作業を開始します。
3. 編集前に current branch と working tree を確認します。

   ```sh
   git status --short --branch
   ```

4. Code または documentation を変更する前に、関連する design document と workflow document を読みます。

   - [docs/design.md](docs/design.md)
   - 変更対象 area の component-level workflow document

5. 既存の component boundary と ownership に合わせて、focused changes を行います。
6. Contract、setup、developer workflow に影響する場合は、関連 documentation と generated artifacts も更新します。
7. まず narrowest useful verification を実行し、shared behavior、contract、persistence、user-facing flow に影響する変更では verification を広げます。
8. Pull Request を作成する前に final diff を確認します。

   ```sh
   git diff
   git status --short
   ```

9. `cmd-create-pr` を使って、Task の Pull Request を作成または更新します。
10. Pull Request URL、changed files、実行した verification、残っている risks または skipped checks を簡潔にまとめて handoff します。

## GitHub 操作

Pull Request の確認、作成、状態確認などの GitHub 操作には GitHub CLI (`gh`) を使います。

## API 生成

API 生成には `generate:api` を使います。

## Documentation Updates

Documentation を更新するときは、英語版の `.md` と日本語版の `*.ja.md` を同期させます。

- 同じ content change に対して両方のファイルを更新します。
- 片方の language file しかない場合は、対応するもう片方を追加します。
- 英語版と日本語版の links を揃えます。
- 日本語版の `*.ja.md` は `AGENTS.md` から link しなくてかまいません。`AGENTS.md` では英語版の `.md` を link します。
- ADR は historical decision record として扱います。typo や broken link のような明らかな mechanical fix を除き、後続 decision に合わせて過去 ADR を書き換えません。新しい decision が過去 ADR を変更または制約する場合は、新しい ADR 側にその変更と関係を書きます。

## Component Workflows

特定の runtime area で作業するときは、component-level workflow documents を使います。

- [Issue Tracker](cmd/issue-tracker/WORKFLOW.ja.md)
- [Orchestrator](cmd/orchestrator/WORKFLOW.ja.md)
- [Web UI](web/WORKFLOW.ja.md)
