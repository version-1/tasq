# tasq

AI coding agent task manager.

tasq は、Claude Code、Codex などの AI coding agent で複数の実装タスクを並列実行するための CLI tool です。

タスクごとに独立した workspace を作成し、task registration、agent execution、state management、review、integration までの workflow を支援します。

## Problem

AI coding agents により、複数の実装タスクを同時に進められるようになりました。ボトルネックは code generation そのものから、並列作業の管理へ移ります。

### 人間側の context switching

Agent は並列で実行できますが、人間は各タスクの状態を追い続ける必要があります。

- どのタスクを依頼したか。
- どの agent が実行中か。
- 各タスクがどこまで進んでいるか。
- 次に何を review すべきか。

これにより coordination cost が増え、active work の全体像を安定して把握しにくくなります。

### Workspace conflicts

複数 agent を 1 つの repository checkout で動かすと、競合が起きやすくなります。

- Branch switching が他の作業を中断する。
- 未完了の変更が重なる。
- 複数 agent が同じ workspace から同じ file を編集する。

### 繰り返しの setup work

Agent task ごとに、同じ準備作業が必要になりがちです。

- Branch を作成する。
- Worktree を作成する。
- Dependencies を install または verify する。
- 適切な setup / verification commands を実行する。

この setup をタスクごとに繰り返すと、parallel execution の速度が落ちます。

## Solution

tasq は executable tasks を queue として管理し、ready になったタスクに対して agent-ready な workspace を作成します。

![Tasq task queue to parallel agent workspaces](docs/site/static/img/agent-task-queue.svg)

目的は code generation を速くすることだけではありません。parallel agent work によって増える management cost を下げることです。

## Features

### Task queue

実装タスクが workflow を進む状態を追跡します。

```sh
tasq add "implement user login"
```

```text
TODO
READY
RUNNING
DONE
```

### Isolated workspace

タスクごとに Git worktree を作成し、agent が独立して作業できるようにします。

```text
project/
├── main
└── .worktrees/
    ├── task-a
    ├── task-b
    └── task-c
```

### Parallel agent execution

1 つの mutable checkout を共有せず、複数 agent を同時に実行できます。

```text
task-a -> Codex A
task-b -> Codex B
task-c -> Codex C
```

### Review workflow

Agent の成果物をタスク単位で review し、統合できます。

```text
RUNNING
   |
   v
REVIEW
   |
   v
MERGED
```

tasq は AI coding-agent workflow のために、task management、workspace isolation、agent execution support を提供します。

## Components

- Issue Tracker: SQLite backed の Go REST API です。issues、comments、projects、workspaces、UI summaries を所有します。
- Orchestrator: SQLite backed の Go service です。runtime inspection 用に run state と runner events を記録します。
- `tq`: agent と workflow tool が issue-tracker API 経由で issue を作成、取得、一覧表示、更新するための Go CLI です。
- Web UI: issue-tracker API 用の Go-served Vite + React client です。
- TUI: 同じ issue-tracker API を使う Go terminal client です。

Architecture 全体は [docs/design.md](docs/design.md) を参照してください。
Local configuration は [docs/design/configuration.ja.md](docs/design/configuration.ja.md) を参照してください。

## Developer Documentation

Local development、verification、command references、repository workflow は [docs/development.ja.md](docs/development.ja.md) を参照してください。

English counterpart: [README.md](README.md).
