# Tasq Symphony Workflow Contract

Tasq は run dispatch 時に project ごとの effective workflow を解決します。既に dispatch 済みの run に
対する runtime reload は意図的に無効化されています。workflow の runtime 設定を変更した場合は run を
再起動または再 dispatch してください。

front matter パーサは YAML を使用し、以下の Tasq 固有フィールドをサポートします。未知のフィールドは
前方互換性のために無視されます。

## サポートされるフィールド

```yaml
tasq:
  task_work_prompt: true
polling:
  interval_ms: 30000
workspace:
  root: .workspaces
agent:
  max_concurrent_agents: 10
  max_turns: 20
  continuation_turns_enabled: false
  max_retry_attempts: 3
  max_retry_backoff_ms: 300000
codex:
  command: codex app-server
  read_timeout_ms: 5000
  turn_timeout_ms: 3600000
  stall_timeout_ms: 300000
server:
  port: 8080
hooks:
  after_create: |
    echo "created workspace"
  before_run: |
    echo "before agent run"
  after_run: |
    echo "after agent run"
  before_remove: |
    echo "before workspace cleanup"
  timeout_ms: 60000
```

`tasq.task_work_prompt` は、orchestrator が rendered agent prompt の先頭に追加する Tasq の
default task-work instructions を制御します。省略時の default は `true` です。workflow template が
独自の issue-tracker synchronization instructions を提供する場合にのみ `false` に設定してください。

`workspace.root` の相対パスは、選択された workflow ファイルからの相対パスとして解決されます。
path フィールドでは環境変数の間接参照と `~` 展開がサポートされています。

`agent.max_turns` が 1 より大きい場合でも、continuation turns はデフォルトで無効です。同じ Codex
thread 上で複数ターンを実行する準備ができている場合に `agent.continuation_turns_enabled` を有効化
してください。

runner progress、workspace metadata、workspace setup failure、cleanup state は orchestrator の SQLite
データベースに保存されます。大きなトランスクリプトは現在の実装ではファイルシステムには書き出されません。

workspace hooks は `bash -lc` で実行され、issue の workspace directory が `cwd` になります。
`hooks.timeout_ms` はすべての hooks に適用され、デフォルトは `60000` ミリ秒です。非正の timeout 値は
workflow configuration validation で失敗します。

hook failure の挙動:

- `after_create` は新規作成された workspace でのみ実行されます。失敗すると workspace 作成が中断され、
  部分的な workspace directory が削除されます。
- `before_run` は各 agent attempt の前に実行されます。失敗すると current attempt が中断されます。
- `after_run` は各 agent attempt の後に実行されます。失敗はログに記録されますが無視されます。
- `before_remove` は directory が存在する場合に workspace 削除の前に実行されます。失敗はログに記録され、
  cleanup は続行されます。

front matter は YAML なので、multiline hook scripts には literal block scalar（`|`）が使えます。

## Prompt Template

front matter の閉じ `---` 以降がすべてプロンプトテンプレートです。orchestrator は agent attempt ごとに
1 回レンダリングし、結果を coding agent への最初のメッセージとして送信します。
`tasq.task_work_prompt` が省略されているか `true` の場合、orchestrator は template variable expansion
の前に default `tq` issue-tracker synchronization instructions を先頭へ追加します。そのため
`{{ issue.id }}` などの変数は default instructions と workflow template の両方で展開されます。

### 利用可能な変数

| 変数                      | 型     | 説明                                            |
|--------------------------|--------|------------------------------------------------|
| `{{ issue.id }}`         | int    | issue-tracker の数値 issue ID                    |
| `{{ issue.title }}`      | string | issue タイトル                                   |
| `{{ issue.description }}`| string | issue 説明本文                                   |
| `{{ attempt }}`          | int    | attempt 番号（初回は 0、リトライは 1 以上）         |

変数は単純な文字列置換で展開されます。認識されない `{{ ... }}` トークンはそのまま残ります。

### 作成ガイドライン

プロンプトテンプレートには、agent に **何をすべきか** と、issue-tracker とやり取りするための
**利用可能なツール** を記述します。

#### Issue ステータス更新

デフォルトでは、Tasq は agent に `tq` CLI で progress comment と issue status update を行うように
指示する instructions を注入します。runtime 環境に `tq` が `PATH` 上にあり、`TQ_API_URL` が
issue-tracker のエンドポイントに設定されている必要があります。

workflow authors は通常、これらの `tq` instructions を prompt template で繰り返す必要はありません。
`tasq.task_work_prompt` を `false` に設定した場合、workflow template 側で同等の issue-tracker
synchronization guidance を提供する責任があります。

#### 成果物

agent がタスク完了を判断できるように、期待する成果物を定義します。一般的なパターン:

- 変更をコミットし、pull request を作成する。
- issue ステータスをハンドオフ状態（例: `review`）に更新する。
- 実装メモを issue の description やコメントに残す。

成功時と失敗時に agent が issue をどのステータスに遷移させるべきか、明示的に記述してください。

#### 含めるべきでないもの

- runtime 設定（poll interval、concurrency、timeout）は front matter に属します。プロンプトには
  含めないでください。
- secret や credential は環境変数から取得すべきです。テンプレートテキストには含めないでください。
