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
