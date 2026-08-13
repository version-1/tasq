# Tasq Symphony Workflow Contract

Tasq は run の割り当て時に、プロジェクトごとの有効なワークフローを解決します。既に割り当て済みの
run に対する実行時の再読み込みは意図的に無効化されています。ワークフローの実行時設定を変更した場合は、
run を再起動または再割り当てしてください。

front matter パーサーは YAML を使用し、以下の Tasq 固有フィールドをサポートします。未知のフィールドは
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

`tasq.task_work_prompt` は、orchestrator がレンダリング済みのエージェントプロンプトの先頭に追加する
Tasq の既定の作業指示を制御します。省略時の既定値は `true` です。ワークフローテンプレートが
独自の issue-tracker 同期指示を提供する場合にのみ `false` に設定してください。

`workspace.root` の相対パスは、選択されたワークフローファイルからの相対パスとして解決されます。
path フィールドでは環境変数の間接参照と `~` 展開がサポートされています。

`agent.max_turns` が 1 より大きい場合でも、継続ターンはデフォルトで無効です。同じ Codex
thread 上で複数ターンを実行する準備ができている場合に `agent.continuation_turns_enabled` を有効化
してください。

runner の進捗、ワークスペースメタデータ、ワークスペース準備の失敗、クリーンアップ状態は orchestrator の
SQLite データベースに保存されます。大きなトランスクリプトは現在の実装ではファイルシステムには書き出されません。

workspace hook は `bash -lc` で実行され、課題の workspace directory が `cwd` になります。
`hooks.timeout_ms` はすべての hook に適用され、デフォルトは `60000` ミリ秒です。非正の timeout 値は
ワークフロー設定の検証で失敗します。

hook 失敗時の挙動:

- `after_create` は新規作成されたワークスペースでのみ実行されます。失敗するとワークスペース作成が中断され、
  部分的な workspace directory が削除されます。
- `before_run` は各エージェント試行の前に実行されます。失敗すると現在の試行が中断されます。
- `after_run` は各エージェント試行の後に実行されます。失敗はログに記録されますが無視されます。
- `before_remove` は directory が存在する場合にワークスペース削除の前に実行されます。失敗はログに記録され、
  クリーンアップは続行されます。

front matter は YAML なので、複数行の hook script には literal block scalar（`|`）が使えます。

## Prompt Template

front matter の閉じ `---` 以降がすべてプロンプトテンプレートです。orchestrator はエージェント試行ごとに
1 回レンダリングし、結果をコーディングエージェントへの最初のメッセージとして送信します。
`tasq.task_work_prompt` が省略されているか `true` の場合、orchestrator は template variable expansion
の前に既定の `tq` issue-tracker 同期指示を先頭へ追加します。そのため
`{{ issue.id }}` などの変数は既定の指示とワークフローテンプレートの両方で展開されます。

### 利用可能な変数

| 変数                      | 型     | 説明                                            |
|--------------------------|--------|------------------------------------------------|
| `{{ issue.id }}`         | int    | issue-tracker の数値課題 ID                    |
| `{{ issue.title }}`      | string | 課題タイトル                                   |
| `{{ issue.description }}`| string | 課題説明本文                                   |
| `{{ attempt }}`          | int    | attempt 番号（初回は 0、リトライは 1 以上）         |
| `{{ tq.command }}`       | string | service 起動元から継承した CLI command         |

変数は単純な文字列置換で展開されます。認識されない `{{ ... }}` トークンはそのまま残ります。

### 作成ガイドライン

プロンプトテンプレートには、エージェントに **何をすべきか** と、issue-tracker とやり取りするための
**利用可能なツール** を記述します。

#### Issue ステータス更新

デフォルトでは、Tasq はエージェントに `{{ tq.command }}` が表す CLI command で progress comment と
issue status update を行うよう指示を注入します。`tq service start` で起動した service は、起動元の
正規化済み executable path を `TQ_EXECUTABLE` で継承します。そのため `tqdev` から起動した場合は、
`PATH` 上で別の `tq` が先に見つかっても同じ `tqdev` executable を使い続けます。この環境契約を持たずに
orchestrator を直接起動した場合は、後方互換性のため `tq` に fallback します。

managed agent run は `TQ_MANAGED_RUN=1` も継承します。この文脈では、run を所有する orchestrator を
終了させる可能性があるため、`tq update` と `tq service stop` は service state を変更する前に失敗します。
service lifecycle command と update command は user shell から実行してください。

ワークフロー作成者は通常、これらの `tq` 指示をプロンプトテンプレートで繰り返す必要はありません。
`tasq.task_work_prompt` を `false` に設定した場合、ワークフローテンプレート側で同等の issue-tracker
同期ガイダンスを提供する責任があります。

#### 成果物

エージェントがタスク完了を判断できるように、期待する成果物を定義します。一般的なパターン:

- 変更をコミットし、pull request を作成する。
- 課題ステータスをハンドオフ状態（例: `review`）に更新する。
- 実装メモを課題の description やコメントに残す。

成功時と失敗時にエージェントが課題をどのステータスに遷移させるべきか、明示的に記述してください。

#### 含めるべきでないもの

- 実行時設定（poll interval、concurrency、timeout）は front matter に属します。プロンプトには
  含めないでください。
- secret や credential は環境変数から取得すべきです。テンプレートテキストには含めないでください。
