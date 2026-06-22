# Symphony サービス仕様

ステータス: Draft v1（言語非依存）

目的: プロジェクト作業を進めるために、コーディングエージェントを統制・実行するサービスを定義する。

この文書は [SPEC.md](SPEC.md) の日本語版です。規範的な判断で差異が出た場合は、取得元の英語仕様である `SPEC.md` を優先します。

## この文書の読み方

この仕様は、Symphony を実装するための契約を定義します。

初めて読む場合は、まず Section 1 から Section 3 で目的と全体像を確認してください。

次に Section 4 から Section 6 で、データモデルと `WORKFLOW.md` の契約を読んでください。実装時は Section 7 以降の状態遷移、ワークスペース、エージェント実行器、トラッカー連携、可観測性を参照します。

読みやすさのため、日本語版では次の表記を使います。

- 仕様上のキーワード（`MUST`, `SHOULD`, `MAY` など）は英語のまま残します。
- 設定名、イベント名、フィールド名、状態名は code style で表記します。
- 一般的な説明文では日本語を優先します。
- プロトコル名、設定キー、外部ツール名など、翻訳すると識別しづらくなる名前は英語のまま残します。
- `implementation-defined` は、実装が選択して文書化すべき振る舞いを指します。

## 規範用語

この文書における `MUST`, `MUST NOT`, `REQUIRED`, `SHOULD`, `SHOULD NOT`, `RECOMMENDED`, `MAY`, `OPTIONAL` は RFC 2119 に従って解釈します。

`Implementation-defined` は、その振る舞いが実装契約の一部である一方、この仕様では単一の普遍的ポリシーを規定しないことを意味します。実装は選択した振る舞いを文書化しなければなりません。

## 1. 問題定義

Symphony は、Issue Tracker から継続的に作業を読み取り、課題ごとに分離されたワークスペースを作成し、そのワークスペース内でコーディングエージェントのセッションを実行する長時間稼働の自動化サービスです。

この仕様版では、Issue Tracker として Linear を想定します。

このサービスは、次の運用上の問題を解決します。

- 課題の実行を手動スクリプトではなく、再現可能な常駐ワークフローにする。
- 課題ごとのワークスペースでエージェント実行を分離し、エージェントのコマンドがそのワークスペース内だけで実行されるようにする。
- ワークフロー方針をリポジトリ内の `WORKFLOW.md` に置き、エージェントのプロンプトと実行時設定をコードとともにバージョン管理する。
- 複数のエージェント実行を運用・デバッグするために十分な可観測性を提供する。

実装は、自身の trust and safety posture を明示的に文書化することが期待されます。

この仕様は、単一の承認、サンドボックス、運用者確認の方針を要求しません。信頼された環境向けに高信頼設定を採る実装もあれば、より厳格な承認やサンドボックスを要求する実装もあります。

重要な境界:

- Symphony はスケジューラー兼実行器であり、トラッカーの読み取り器です。
- チケットの書き込み、たとえば状態遷移、コメント、PR リンクの追加は、通常ワークフローや実行環境のツールを使ってコーディングエージェントが行います。
- 成功した実行は、ワークフローが定義する引き渡し状態、たとえば `Human Review` で終わることがあり、必ずしも `Done` で終わるとは限りません。

## 2. 目標と非目標

### 2.1 目標

- 固定間隔で Issue Tracker をポーリングし、上限付き並行数で作業を割り当てる。
- 割り当て、再試行、整合処理のために、唯一の正とするオーケストレーター状態を維持する。
- 課題ごとに決定的なワークスペースを作成し、実行間で保持する。
- 課題状態の変化により不適格になった実行中の処理を停止する。
- 一時的な失敗から指数バックオフで回復する。
- リポジトリが所有する `WORKFLOW.md` 契約から実行時の振る舞いを読み込む。
- 運用者が確認できる可観測性、少なくとも構造化ログを公開する。
- 永続データベースを必須にせず、トラッカーとファイルシステムをもとにした再起動時の回復をサポートする。ただし、正確なインメモリのスケジューラー状態は復元しない。

### 2.2 非目標

- 高機能な Web UI やマルチテナントの制御プレーン。
- 特定のダッシュボードや端末 UI 実装の規定。
- 汎用ワークフローエンジンや分散ジョブスケジューラー。
- チケット、PR、コメントの編集方法に関する組み込みの業務ロジック。そのロジックはワークフロープロンプトとエージェントのツール群に属します。
- コーディングエージェントとホスト OS が提供するものを超えた、強力なサンドボックス制御の義務化。
- すべての実装に対する単一の既定承認、サンドボックス、運用者確認姿勢の義務化。

## 3. システム概要

### 3.1 主要コンポーネント

1. `Workflow Loader`
   - `WORKFLOW.md` を読む。
   - YAML front matter とプロンプト本文を解析する。
   - `{config, prompt_template}` を返す。

2. `Config Layer`
   - ワークフロー設定値に対する型付き getter を提供する。
   - 既定値と環境変数参照を適用する。
   - 割り当て前にオーケストレーターが使う検証を行う。

3. `Issue Tracker Client`
   - アクティブ状態の割り当て候補課題を取得する。
   - 整合処理のため、特定の課題 ID の現在状態を取得する。
   - 起動時クリーンアップのため、終端状態の課題を取得する。
   - トラッカーのペイロードを安定した課題モデルに正規化する。

4. `Orchestrator`
   - ポーリング tick を所有する。
   - インメモリの実行時状態を所有する。
   - どの課題を割り当て、再試行し、停止し、解放するかを決定する。
   - セッション指標と再試行キューの状態を追跡する。

5. `Workspace Manager`
   - 課題識別子をワークスペースのパスへ対応付ける。
   - 課題ごとのワークスペースディレクトリが存在することを保証する。
   - ワークスペースのライフサイクル hook を実行する。
   - 終端状態の課題のワークスペースをクリーンアップする。

6. `Agent Runner`
   - ワークスペースを作成する。
   - 課題とワークフローテンプレートからプロンプトを構築する。
   - コーディングエージェントの app-server client を起動する。
   - エージェントの更新をオーケストレーターにストリームする。

7. `Status Surface` (OPTIONAL)
   - 端末出力やダッシュボードなど、運用者向けの表示として、人間が読める実行時状態を提示する。

8. `Logging`
   - 構造化された実行時ログを、設定された 1 つ以上の出力先へ出力する。

### 3.2 抽象レベル

Symphony は次の層に分けて保つと移植しやすくなります。

1. `Policy Layer` (repo-defined): `WORKFLOW.md` のプロンプト本文、チケット処理、検証、引き渡しに関するチーム固有の規則。
2. `Configuration Layer` (typed getters): front matter を型付き実行時設定として解析し、既定値、環境変数トークン、パス正規化を扱う。
3. `Coordination Layer` (orchestrator): ポーリングループ、課題の適格性、並行数、再試行、整合処理。
4. `Execution Layer` (workspace + agent subprocess): ファイルシステムのライフサイクル、ワークスペース準備、コーディングエージェントのプロトコル。
5. `Integration Layer` (Linear adapter): API 呼び出しとトラッカーデータの正規化。
6. `Observability Layer` (logs + OPTIONAL status surface): オーケストレーターとエージェントの振る舞いを運用者が確認できるようにする層。

### 3.3 外部依存

- Issue Tracker API。この仕様版では `tracker.kind: linear` に対して Linear。
- ワークスペースとログのためのローカルファイルシステム。
- OPTIONAL なワークスペース準備ツール。たとえば Git CLI。
- 対象 Codex app-server mode をサポートするコーディングエージェント実行ファイル。
- Issue Tracker とコーディングエージェントのためのホスト環境認証。

## 4. コアドメインモデル

### 4.1 エンティティ

#### 4.1.1 Issue

オーケストレーション、プロンプト描画、可観測性出力で使う、正規化された課題レコードです。

Fields:

- `id` (string)
  - トラッカー内部で安定している ID。
- `identifier` (string)
  - 人間が読めるチケットキー。
  - 例: `ABC-123`。
- `title` (string)
- `description` (string or null)
- `priority` (integer or null)
  - 小さい数値ほど、割り当て時の並び順で優先度が高い。
- `state` (string)
  - 現在のトラッカー状態名。
- `branch_name` (string or null)
  - トラッカーが提供するブランチメタデータ。
- `url` (string or null)
- `labels` (list of strings)
  - 小文字へ正規化する。
- `blocked_by` (list of blocker refs)
  - 各 blocker ref は次のフィールドを持つ。
    - `id` (string or null)
    - `identifier` (string or null)
    - `state` (string or null)
- `created_at` (timestamp or null)
- `updated_at` (timestamp or null)

#### 4.1.2 Workflow Definition

解析済みの `WORKFLOW.md` ペイロード:

- `config` (map)
  - YAML front matter の root object。
- `prompt_template` (string)
  - front matter 後の Markdown body を trim したもの。

#### 4.1.3 Service Config (Typed View)

`WorkflowDefinition.config` と環境解決から導かれる、型付きの実行時値です。

例として、次の値を含みます。

- ポーリング間隔
- ワークスペースルート
- アクティブ状態と終端状態の課題状態
- 並行数の上限
- コーディングエージェントの実行ファイル、引数、タイムアウト
- ワークスペース hook

#### 4.1.4 Workspace

1 つの課題識別子に割り当てられる、ファイルシステム上のワークスペースです。

論理フィールド:

- `path`: 絶対ワークスペースパス。
- `workspace_key`: サニタイズ済みの課題識別子。
- `created_now`: `after_create` hook の実行判定に使う boolean。

#### 4.1.5 Run Attempt

1 つの課題に対する 1 回の実行試行です。

論理フィールド:

- `issue_id`
- `issue_identifier`
- `attempt`: 初回は `null`、再試行または継続実行は `>=1`。
- `workspace_path`
- `started_at`
- `status`
- `error` (OPTIONAL)

#### 4.1.6 Live Session (Agent Session Metadata)

コーディングエージェントのサブプロセス実行中に追跡する状態です。

Fields:

- `session_id`: `<thread_id>-<turn_id>`。
- `thread_id`
- `turn_id`
- `codex_app_server_pid`
- `last_codex_event`
- `last_codex_timestamp`
- `last_codex_message`
- `codex_input_tokens`
- `codex_output_tokens`
- `codex_total_tokens`
- `last_reported_input_tokens`
- `last_reported_output_tokens`
- `last_reported_total_tokens`
- `turn_count`: 現在の worker の生存期間内で開始した、コーディングエージェントの turn 数。

#### 4.1.7 Retry Entry

課題に対する、予定済みの再試行状態です。

Fields:

- `issue_id`
- `identifier`: 状態表示やログで使う best-effort の人間向け ID。
- `attempt`: 再試行キューにおける 1 始まりの整数。
- `due_at_ms`: monotonic clock の timestamp。
- `timer_handle`: 実行時固有の timer reference。
- `error`: string or null。

#### 4.1.8 Orchestrator Runtime State

オーケストレーターが所有する、唯一の正とするインメモリ状態です。

Fields:

- `poll_interval_ms`
- `max_concurrent_agents`
- `running`: `issue_id -> running entry` の map。
- `claimed`: 予約済み、実行中、再試行中の課題 ID の set。
- `retry_attempts`: `issue_id -> RetryEntry` の map。
- `completed`: 記録用の set。割り当て判定には使わない。
- `codex_totals`: 集計済み token 数と実行秒数。
- `codex_rate_limits`: エージェントイベントから得た最新の rate-limit snapshot。

### 4.2 Stable Identifier と Normalization Rule

- `Issue ID`
  - トラッカー検索と内部 map key に使う。
- `Issue Identifier`
  - 人間が読めるログとワークスペース命名に使う。
- `Workspace Key`
  - `issue.identifier` のうち `[A-Za-z0-9._-]` 以外を `_` に置換して作る。
  - ワークスペースディレクトリ名に使う。
- `Normalized Issue State`
  - 小文字化した後に比較する。
- `Session ID`
  - コーディングエージェントの `thread_id` と `turn_id` から `<thread_id>-<turn_id>` として構成する。

## 5. Workflow Specification (Repository Contract)

### 5.1 File Discovery and Path Resolution

ワークフローファイルパスの優先順位:

1. 明示的なアプリケーション設定または実行時設定。CLI 起動時に渡すパスなど。
2. 既定値: 現在のプロセス作業ディレクトリにある `WORKFLOW.md`。

Loader の振る舞い:

- ファイルを読めない場合は `missing_workflow_file` error を返す。
- ワークフローファイルはリポジトリが所有し、バージョン管理されていることが期待される。

### 5.2 File Format

`WORKFLOW.md` は、OPTIONAL な YAML front matter を持つ Markdown ファイルです。

`WORKFLOW.md` は、プロンプト、実行時設定、hook、トラッカーの選択と設定を、別経路のサービス固有設定なしで記述・実行できる程度に自己完結しているべきです。

解析規則:

- ファイルが `---` で始まる場合、次の `---` までを YAML front matter として解析する。
- 残りの行はプロンプト本文になる。
- front matter がない場合、全体をプロンプト本文とし、空の config map を使う。
- YAML front matter は map/object に decode されなければならず、非 map YAML は error とする。
- プロンプト本文は使用前に trim する。

返されるワークフローオブジェクト:

- `config`: `config` key の下ではなく front matter の root object。
- `prompt_template`: trim 済み Markdown 本文。

### 5.3 Front Matter Schema

Top-level keys:

- `tracker`
- `polling`
- `workspace`
- `hooks`
- `agent`
- `codex`

未知の key は、将来互換性のために無視すべきです。拡張機能は追加の top-level key を定義できますが、フィールド schema、既定値、検証規則、動的適用か再起動必須かを文書化すべきです。

#### 5.3.1 `tracker` (object)

- `kind` (string)
  - 割り当てに REQUIRED。
  - 現在サポートされる値は `linear`。
- `endpoint` (string)
  - `tracker.kind == "linear"` の既定値は `https://api.linear.app/graphql`。
- `api_key` (string)
  - literal token または `$VAR_NAME`。
  - Linear の標準環境変数は `LINEAR_API_KEY`。
  - `$VAR_NAME` が空に解決される場合は missing と扱う。
- `project_slug` (string)
  - `tracker.kind == "linear"` の割り当てに REQUIRED。
- `active_states` (list of strings)
  - 既定値は `Todo`, `In Progress`。
- `terminal_states` (list of strings)
  - 既定値は `Closed`, `Cancelled`, `Canceled`, `Duplicate`, `Done`。

#### 5.3.2 `polling` (object)

- `interval_ms` (integer)
  - 既定値は `30000`。
  - 変更は実行時に再適用され、再起動なしで以後の tick scheduling に影響すべきです。

#### 5.3.3 `workspace` (object)

- `root` (path string or `$VAR`)
  - 既定値は `<system-temp>/symphony_workspaces`。
- `~` は展開する。
- relative path は `WORKFLOW.md` があるディレクトリから解決する。
- 実効ワークスペースルートは、使用前に絶対パスへ正規化する。

#### 5.3.4 `hooks` (object)

サポートするフィールド:

- `after_create`: 新しいワークスペースディレクトリが作成された場合だけ実行する。失敗はワークスペース作成の中断として扱う。
- `before_run`: ワークスペース準備後、コーディングエージェント起動前に各試行で実行する。失敗は現在の試行の中断として扱う。
- `after_run`: 成功、失敗、タイムアウト、キャンセルの後、ワークスペースが存在する場合に実行する。失敗はログに残して無視する。
- `before_remove`: ディレクトリが存在する場合、ワークスペース削除前に実行する。失敗はログに残して無視し、クリーンアップは継続する。
- `timeout_ms`: 既定値は `60000`。すべてのワークスペース hook に適用する。不正な値は設定検証失敗とする。変更は以後の hook 実行に再適用すべきです。

#### 5.3.5 `agent` (object)

- `max_concurrent_agents` (integer)
  - 既定値は `10`。
  - 変更は以後の割り当て判断に反映すべきです。
- `max_turns` (positive integer)
  - 既定値は `20`。
  - 1 worker session 内のコーディングエージェント turn 数を制限する。
  - 不正な値は検証失敗とする。
- `max_retry_backoff_ms` (integer)
  - 既定値は `300000`。
  - 以後の再試行スケジューリングに影響する。
- `max_concurrent_agents_by_state` (map `state_name -> positive integer`)
  - 既定値は `{}`。
  - 状態 key は小文字へ正規化する。
  - 正の数でない値や数値でない値は無視する。

#### 5.3.6 `codex` (object)

`approval_policy`, `thread_sandbox`, `turn_sandbox_policy` など、Codex が所有する設定値でサポートされる値は、対象 Codex app-server version が定義します。

実装者は、この仕様内の enum に頼らず、Codex 設定値をそのまま渡す値として扱うべきです。

インストール済みの schema は `codex app-server generate-json-schema --out <dir>` で確認できます。

- `command`
  - shell command string。
  - 既定値は `codex app-server`。
  - 実行時はワークスペースディレクトリで `bash -lc` により起動する。
  - 起動されたプロセスは、stdio 上で互換性のある app-server protocol を話さなければなりません。
- `approval_policy`
  - Codex `AskForApproval` value。
  - 既定値は implementation-defined。
- `thread_sandbox`
  - Codex `SandboxMode` value。
  - 既定値は implementation-defined。
- `turn_sandbox_policy`
  - Codex `SandboxPolicy` value。
  - 既定値は implementation-defined。
- `turn_timeout_ms`
  - 既定値は `3600000`。
- `read_timeout_ms`
  - 既定値は `5000`。
- `stall_timeout_ms`
  - 既定値は `300000`。
  - `<= 0` の場合、stall detection は無効。

### 5.4 Prompt Template Contract

`WORKFLOW.md` の Markdown 本文は、課題ごとのプロンプトテンプレートです。

描画要件:

- strict template engine を使う。Liquid 互換の意味論で十分です。
- 未知の変数は描画失敗。
- 未知の filter は描画失敗。

テンプレート入力変数:

- `issue`: 正規化された課題フィールド、labels、blockers を含む object。
- `attempt`: 初回試行では `null` または absent、再試行または継続実行では integer。

フォールバックプロンプト:

- ワークフローのプロンプト本文が空の場合、実行時は最小限の既定プロンプトを使ってもよい。
- ワークフローファイルの読み取りまたは解析失敗は設定または検証 error であり、黙って fallback してはなりません。

### 5.5 Workflow Validation and Error Surface

Error classes:

- `missing_workflow_file`
- `workflow_parse_error`
- `workflow_front_matter_not_a_map`
- `template_parse_error`
- `template_render_error`

ワークフローファイルの読み取り error や YAML error は、修正されるまで新しい割り当てを block します。テンプレート error は、影響を受けた実行試行だけを fail させます。

## 6. Configuration Specification

### 6.1 Configuration Resolution Pipeline

設定は次の順に解決します。

1. ワークフローファイルパスを選択する。
2. YAML front matter を raw config map として解析する。
3. missing OPTIONAL fields に組み込み既定値を適用する。
4. 設定値が明示的に `$VAR_NAME` を含む場合だけ間接参照を解決する。
5. 型付き値へ変換し、検証する。

環境変数は YAML value を全体的に上書きしません。設定値が明示的に参照した場合だけ使います。

パスとコマンドの変換:

- path field は `~` と、環境変数由来の path value に含まれる `$VAR` expansion をサポートする。
- expansion はローカルファイルシステムパスとして意図された値にだけ適用し、URI や任意の shell command string は rewrite しない。
- relative `workspace.root` は、選択された `WORKFLOW.md` のディレクトリから解決する。

### 6.2 Dynamic Reload Semantics

dynamic reload は REQUIRED です。

- software は `WORKFLOW.md` changes を detect しなければなりません。
- change 時には workflow config と prompt template を restart なしで re-read/re-apply しなければなりません。
- polling cadence、concurrency limits、active/terminal states、codex settings、workspace paths/hooks、future run の prompt content などの live behavior を調整しようとしなければなりません。
- reloaded config は future dispatch、retry scheduling、reconciliation decision、hook execution、agent launch に適用されます。
- in-flight agent session を自動 restart することは REQUIRED ではありません。
- listener/resource を持つ extension は、implementation が live rebind を support しない限り restart を要求してもよい。
- filesystem watch event missed に備え、dispatch 前など runtime operation 中にも defensive reload/revalidation すべきです。
- invalid reload は service を crash させてはならず、last known good effective configuration で運用を続け、operator-visible error を出します。

### 6.3 Dispatch Preflight Validation

dispatch preflight validation は new work を dispatch する前の scheduler preflight です。poll と worker launch に必要な workflow/config を validate し、すべての workflow behavior を audit するものではありません。

Startup:

- scheduling loop 開始前に configuration を validate する。
- startup validation failure は startup failure とし、operator-visible error を emit する。

Per tick:

- 各 dispatch cycle 前に re-validate する。
- validation failure の場合、その tick の dispatch は skip し、reconciliation は active に保ち、operator-visible error を emit する。

Validation checks:

- workflow file が load/parse できる。
- `tracker.kind` が存在し supported。
- `$` resolution 後の `tracker.api_key` が存在する。
- selected tracker kind に必要な `tracker.project_slug` が存在する。
- `codex.command` が存在し non-empty。

### 6.4 Core Config Fields Summary

core conformance では、実装した extension を除き extension fields の認識や validation は不要です。主要 field は次のとおりです。

- `tracker.kind`: required string。現在は `linear`。
- `tracker.endpoint`: `tracker.kind=linear` の default は `https://api.linear.app/graphql`。
- `tracker.api_key`: string or `$VAR`。canonical env は `LINEAR_API_KEY`。
- `tracker.project_slug`: Linear で required。
- `tracker.active_states`: default `["Todo", "In Progress"]`。
- `tracker.terminal_states`: default `["Closed", "Cancelled", "Canceled", "Duplicate", "Done"]`。
- `polling.interval_ms`: default `30000`。
- `workspace.root`: absolute に resolve される path。default `<system-temp>/symphony_workspaces`。
- hooks: `after_create`, `before_run`, `after_run`, `before_remove`, `timeout_ms`。
- agent: `max_concurrent_agents`, `max_turns`, `max_retry_backoff_ms`, `max_concurrent_agents_by_state`。
- codex: `command`, `approval_policy`, `thread_sandbox`, `turn_sandbox_policy`, `turn_timeout_ms`, `read_timeout_ms`, `stall_timeout_ms`。

## 7. Orchestration State Machine

orchestrator は scheduling state を mutate する唯一の component です。すべての worker outcome は orchestrator に報告され、明示的な state transition に変換されます。

### 7.1 Issue Orchestration States

これは tracker state ではなく、service internal claim state です。

1. `Unclaimed`: issue は running ではなく retry scheduled もない。
2. `Claimed`: duplicate dispatch を避けるため orchestrator が issue を reserve 済み。実際には `Running` または `RetryQueued`。
3. `Running`: worker task が存在し、issue が `running` map にある。
4. `RetryQueued`: worker は running ではないが retry timer が `retry_attempts` にある。
5. `Released`: terminal、non-active、missing、または retry path 完了により claim が removed。

successful worker exit は、issue が永久に完了したことを意味しません。

worker は exit 前に、同じ live thread/workspace で複数の coding-agent turn を続けてもよいです。その場合、各 normal turn completion 後に tracker issue state を re-check します。

issue が active state のままであれば、`agent.max_turns` まで同じ live thread で continuation turn を始めるべきです。

- first turn は full rendered task prompt を使う。
- continuation turn では original prompt を再送しない。
- continuation turn には continuation guidance のみを送るべきです。

worker が normally exit した後も、orchestrator は約 1 秒の continuation retry を schedule します。これは、issue がまだ active で別 worker session が必要か確認するためです。

### 7.2 Run Attempt Lifecycle

run attempt は次の state を遷移します。

- `PreparingWorkspace`
- `BuildingPrompt`
- `LaunchingAgentProcess`
- `InitializingSession`
- `StreamingTurn`
- `Finishing`
- `Succeeded`
- `Failed`
- `TimedOut`
- `Stalled`
- `CanceledByReconciliation`

terminal reason は retry logic と log が異なるため重要です。

### 7.3 Transition Triggers

- `Poll Tick`: active run の reconcile、config validation、candidate issue fetch、slot が尽きるまで dispatch。
- `Worker Exit (normal)`: running entry を remove、runtime totals を update、continuation retry を schedule。
- `Worker Exit (abnormal)`: running entry を remove、runtime totals を update、exponential-backoff retry を schedule。
- `Codex Update Event`: live session fields、token counters、rate limits を update。
- `Retry Timer Fired`: active candidates を re-fetch し re-dispatch または release claim。
- `Reconciliation State Refresh`: terminal または non-active issue state の run を stop。
- `Stall Timeout`: worker を kill して retry を schedule。

### 7.4 Idempotency and Recovery Rules

- orchestrator は duplicate dispatch を避けるため state mutation を 1 つの authority で serialize する。
- worker launch 前に `claimed` と `running` checks が REQUIRED。
- reconciliation は各 tick の dispatch 前に走る。
- restart recovery は durable orchestrator DB なしで tracker-driven かつ filesystem-driven。
- startup terminal cleanup はすでに terminal state の issue の stale workspace を remove する。

## 8. Polling, Scheduling, and Reconciliation

### 8.1 Poll Loop

startup では config validation、startup cleanup、immediate tick scheduling を行い、その後 `polling.interval_ms` ごとに repeat します。effective poll interval は workflow config の再適用時に update すべきです。

Tick sequence:

1. running issue を reconcile する。
2. dispatch preflight validation を行う。
3. active states で tracker から candidate issue を fetch する。
4. dispatch priority で issue を sort する。
5. slot が残る間 eligible issue を dispatch する。
6. observability/status consumer に state change を notify する。

per-tick validation failure の場合、その tick の dispatch は skip しますが reconciliation は先に実行します。

### 8.2 Candidate Selection Rules

issue は次をすべて満たす場合だけ dispatch-eligible です。

- `id`, `identifier`, `title`, `state` がある。
- state が `active_states` にあり `terminal_states` ではない。
- まだ `running` ではない。
- まだ `claimed` ではない。
- global concurrency slot がある。
- per-state concurrency slot がある。
- `Todo` state の blocker rule を満たす。`Todo` の場合、non-terminal blocker が 1 つでもあれば dispatch しない。

Sorting order:

1. `priority` ascending。null/unknown は last。
2. `created_at` oldest first。
3. `identifier` lexicographic tie-breaker。

### 8.3 Concurrency Control

global limit は `available_slots = max(max_concurrent_agents - running_count, 0)` です。

per-state limit は、normalized state key に `max_concurrent_agents_by_state[state]` があればそれを使います。なければ global limit に fallback します。

runtime は `running` map の current tracked state で issue を count します。

### 8.4 Retry and Backoff

retry entry 作成時は同じ issue の既存 retry timer を cancel し、`attempt`, `identifier`, `error`, `due_at_ms`, new timer handle を store します。

normal continuation retry は fixed delay `1000` ms。failure-driven retry は `delay = min(10000 * 2^(attempt - 1), agent.max_retry_backoff_ms)` です。

retry handling:

1. active candidate issues を fetch する。
2. `issue_id` で specific issue を探す。
3. 見つからなければ release claim。
4. 見つかり candidate-eligible なら、slot があれば dispatch、なければ `no available orchestrator slots` で requeue。
5. 見つかったが active でなければ release claim。

### 8.5 Active Run Reconciliation

reconciliation は毎 tick 実行されます。

処理は次の 2 部構成です。

1. stall detection
2. tracker state refresh

stall detection では、event があれば `last_codex_timestamp`、なければ `started_at` から `elapsed_ms` を計算します。

`elapsed_ms > codex.stall_timeout_ms` なら worker を terminate し retry を queue します。`stall_timeout_ms <= 0` の場合は skip します。

tracker state refresh では、running issue ID の current issue state を fetch します。

- terminal の場合:
  - worker を terminate する。
  - workspace cleanup を行う。
  - workspace-scoped coding-agent artifact を破棄する。
  - persisted thread/rollout state を破棄する。
  - 後続 run がその state に reconnect できる resume pointer も破棄する。
- active の場合:
  - in-memory issue snapshot を update する。
- active でも terminal でもない場合:
  - cleanup なしで terminate する。
- refresh failure の場合:
  - worker は running のままにする。
  - 次 tick で再試行する。

### 8.6 Startup Terminal Workspace Cleanup

service startup 時、terminal states の issue を query し、返された issue identifier ごとに workspace directory を remove します。

さらに、その issue の persisted coding-agent thread/rollout artifact と orchestrator-owned resume pointer を delete または invalidate します。

terminal issue fetch が失敗した場合は warning を log して startup を続行します。

これにより、restart 後に stale terminal workspace と resumable terminal-issue conversation が蓄積しないようにします。

## 9. Workspace Management and Safety

### 9.1 Workspace Layout

workspace root は normalized absolute `workspace.root` です。

issue ごとの path は `<workspace.root>/<sanitized_issue_identifier>` です。workspace は同じ issue の run 間で reuse し、successful run で自動 delete しません。

### 9.2 Workspace Creation and Reuse

`issue.identifier` を sanitize して `workspace_key` を作り、workspace root 配下に path を計算し、directory が存在することを保証します。

この call で directory を作った場合だけ `created_now=true` とし、configured なら `after_create` hook を実行します。

workspace preparation の directory creation 以外、たとえば checkout/sync/code generation は implementation-defined であり、通常 hooks で扱います。

### 9.3 OPTIONAL Workspace Population

仕様は built-in VCS/repository bootstrap behavior を要求しません。

実装は implementation-defined logic や hooks で workspace を populate/synchronize してよいです。failure は current attempt の error として返します。

new workspace の準備中 failure では partial directory を remove してもよいですが、reused workspace は明示的に選択・文書化した policy なしに destructively reset すべきではありません。

### 9.4 Workspace Hooks

hooks は workspace directory を `cwd` として、host OS に適した local shell context で実行します。

POSIX では `sh -lc <script>` または `bash -lc <script>` が conforming default です。

hook timeout は `hooks.timeout_ms`、default `60000 ms` です。hook start、failure、timeout を log します。

failure semantics:

- `after_create` failure/timeout は workspace creation に fatal。
- `before_run` failure/timeout は current run attempt に fatal。
- `after_run` failure/timeout は log して ignore。
- `before_remove` failure/timeout は log して ignore。

### 9.5 Safety Invariants

最重要の portability constraint です。

1. coding agent は issue ごとの workspace path 内でのみ実行する。launch 前に `cwd == workspace_path` を validate する。
2. workspace path は workspace root 内に留まらなければならない。両方を absolute normalize し、workspace root を prefix directory として持つことを require する。
3. workspace key は sanitize する。directory name は `[A-Za-z0-9._-]` のみを許可し、それ以外は `_` に置換する。

## 10. Agent Runner Protocol (Coding Agent Integration)

この章は、Codex app-server を統合する際の Symphony の language-neutral responsibility を定義します。

対象 Codex version の app-server protocol が、protocol schema、message payload、transport framing、method name の source of truth です。

実装は対象 Codex app-server version に有効な message を送らなければなりません。

この仕様と対象 protocol が衝突する場合、protocol shape と transport behavior は Codex protocol が優先します。一方、orchestration behavior、workspace selection、prompt construction、continuation handling、observability extraction はこの章の Symphony-specific requirements が制御します。

### 10.1 Launch Contract

- command: `codex.command`
- invocation: `bash -lc <codex.command>`
- working directory: workspace path
- transport/framing: 対象 Codex app-server version が要求する protocol transport

default command は `codex app-server` です。

approval policy、sandbox policy、cwd、prompt input、optional tool declarations は、対象 protocol が support する field で supplied します。

max line size は 10 MB を推奨します。

### 10.2 Session Startup Responsibilities

startup は対象 Codex app-server contract に従います。

Symphony client は、さらに次の責務を持ちます。

- per-issue workspace で app-server subprocess を起動する。
- protocol に従って session/thread/turn を開始する。
- cwd を受け取る protocol field では、absolute per-issue workspace path を渡す。
- first turn では rendered issue prompt を使う。
- 実装が文書化した approval/sandbox policy を、対象 protocol の supported field で渡す。
- 可能なら issue-identifying metadata を title/session metadata に含める。

previous worker run から resume する場合は、fresh app-server subprocess を起動して persisted thread identity に reconnect します。previous subprocess を worker run をまたいで alive にしてはなりません。

continuation turn では、同じ live thread に continuation guidance を送ります。

`thread_id` と `turn_id` は対象 protocol から抽出し、`session_id = "<thread_id>-<turn_id>"` を emit します。

1 worker run 内の continuation turn は同じ `thread_id` を reuse します。persisted `thread_id` は issue が non-terminal の間だけ再利用できます。

Terminal issue cleanup は、future dispatch が同じ issue から開始される前に persisted thread/rollout state を remove し、resume pointer を clear しなければなりません。

### 10.3 Streaming Turn Processing

client は、active turn が terminate するまで対象 Codex app-server protocol に従って update を処理します。

turn outcome は次のように扱います。

- protocol completion は success。
- protocol failure/cancellation、turn timeout、subprocess exit は failure。

continuation が必要な場合は、同じ live thread で別 turn を開始します。app-server subprocess は worker run の終了まで alive に保つべきです。

app-server subprocess を issue-lifetime process と扱ってはなりません。worker exit では常に subprocess を close します。後続 retry や continuation retry は、別 subprocess を起動して persisted protocol state を使って resume します。

stdio transport では、protocol stream と diagnostic stderr handling を分離します。

### 10.4 Emitted Runtime Events

app-server client は orchestrator callback に structured event を emit します。

event は `event`, `timestamp`, `codex_app_server_pid`, optional `usage`, payload fields を含むべきです。

重要な event 例:

- `session_started`
- `startup_failed`
- `turn_completed`
- `turn_failed`
- `turn_cancelled`
- `turn_ended_with_error`
- `turn_input_required`
- `approval_auto_approved`
- `unsupported_tool_call`
- `notification`
- `other_message`
- `malformed`

### 10.5 Approval, Tool Calls, and User Input Policy

approval、sandbox、user-input behavior は implementation-defined です。

各実装は、選択した approval/sandbox/operator-confirmation posture を文書化しなければなりません。

approval request や user-input-required event は、run を indefinitely stalled にしてはなりません。

high-trust behavior の例:

- command execution approval を auto-approve する。
- file-change approval を auto-approve する。
- user-input-required turn を hard failure とする。

unsupported dynamic tool call は、対象 protocol に従って tool failure response を返し session を続けます。

Optional client-side tool extension として `linear_graphql` を定義します。

これは configured tracker auth を使って、Linear に raw GraphQL query/mutation を実行する tool です。

Input contract:

- `query` は non-empty string かつ exactly one GraphQL operation でなければならない。
- `variables` は optional JSON object。

runtime は configured Linear endpoint/auth を reuse し、agent が raw token を disk から読む必要がないようにします。

Result mapping:

- transport success かつ top-level GraphQL `errors` なしなら `success=true`。
- GraphQL errors がある場合は `success=false` で body を保持する。
- invalid input、missing auth、transport failure は error payload を返す。

### 10.6 Timeouts and Error Mapping

timeouts:

- `codex.read_timeout_ms`: startup と sync request の request/response timeout。
- `codex.turn_timeout_ms`: turn stream total timeout。
- `codex.stall_timeout_ms`: event inactivity に基づき orchestrator が enforce。

recommended normalized categories:

- `codex_not_found`
- `invalid_workspace_cwd`
- `response_timeout`
- `turn_timeout`
- `port_exit`
- `response_error`
- `turn_failed`
- `turn_cancelled`
- `turn_input_required`

### 10.7 Agent Runner Contract

`Agent Runner` は workspace、prompt、app-server client を wrap します。

責務:

- workspace を create/reuse する。
- workflow template から prompt を build する。
- app-server session を start する。
- event を orchestrator に forward する。

error は worker attempt failure とし、retry は orchestrator に任せます。successful run 後も workspace は保持します。

## 11. Issue Tracker Integration Contract

### 11.1 REQUIRED Operations

実装は次の tracker adapter operations を support しなければなりません。

1. `fetch_candidate_issues()`: configured project の configured active states の issue を返す。
2. `fetch_issues_by_states(state_names)`: startup terminal cleanup に使う。
3. `fetch_issue_states_by_ids(issue_ids)`: active-run reconciliation に使う。

### 11.2 Query Semantics (Linear)

`tracker.kind == "linear"` の requirements:

- GraphQL endpoint default は `https://api.linear.app/graphql`。
- auth token は `Authorization` header で送る。
- `tracker.project_slug` は Linear project `slugId` に map する。
- candidate issue query は `project: { slugId: { eq: $projectSlug } }` で project filter する。
- issue-state refresh query は GraphQL issue ID と variable type `[ID!]` を使う。
- candidate issues には pagination REQUIRED。
- default page size は `50`。
- network timeout は `30000 ms`。

Linear GraphQL schema は drift し得るため、query construction は isolate し、この仕様が要求する exact query fields/types を test します。

non-Linear implementation は transport detail を変えてよいですが、normalized output は Section 4 の domain model に一致しなければなりません。

### 11.3 Normalization Rules

candidate issue normalization は Section 4.1.1 の field を生成すべきです。

- `labels` は lowercase。
- `blocked_by` は relation type `blocks` の inverse relations から derived。
- `priority` は integer のみ。
- `created_at`/`updated_at` は ISO-8601 timestamp として parse する。

### 11.4 Error Handling Contract

recommended error categories:

- `unsupported_tracker_kind`
- `missing_tracker_api_key`
- `missing_tracker_project_slug`
- `linear_api_request`
- `linear_api_status`
- `linear_graphql_errors`
- `linear_unknown_payload`
- `linear_missing_end_cursor`

- candidate fetch failure は log して、その tick の dispatch を skip します。
- running-state refresh failure は log して、active worker を running のままにします。
- startup terminal cleanup failure は warning を log し、startup を続行します。

### 11.5 Tracker Writes

Symphony は orchestrator に first-class tracker write API を要求しません。

ticket mutation は通常、workflow prompt で定義された tool を使って coding agent が行います。service は scheduler/runner と tracker reader のままです。

workflow-specific success は、`Done` ではなく `Human Review` のような next handoff state に到達することを意味する場合があります。

## 12. Prompt Construction and Context Assembly

### 12.1 Inputs

prompt rendering の input は `workflow.prompt_template`、normalized `issue` object、optional `attempt` integer です。

### 12.2 Rendering Rules

strict variable/filter checking で render します。template compatibility のため issue object key を string に変換し、labels/blockers など nested arrays/maps は template が iterate できるよう保持します。

### 12.3 Retry/Continuation Semantics

workflow prompt が first run、successful prior session 後の continuation、error/timeout/stall 後の retry で異なる instruction を出せるように、`attempt` を template に渡すべきです。

### 12.4 Failure Semantics

prompt rendering failure は run attempt を即時 fail させ、orchestrator が他の worker failure と同様に retry behavior を決定します。

## 13. Logging, Status, and Observability

### 13.1 Logging Conventions

issue-related log には `issue_id` と `issue_identifier` が REQUIRED です。

coding-agent session lifecycle log には `session_id` が REQUIRED です。

stable `key=value` phrasing を使い、outcome と concise failure reason を含めます。必要がない限り large raw payload を log しません。

### 13.2 Logging Outputs and Sinks

log の出力先は規定しません。operator は debugger なしで startup/validation/dispatch failure を見られなければなりません。configured log sink が失敗しても可能なら service は running を続け、残りの sink で warning を emit すべきです。

### 13.3 Runtime Snapshot / Monitoring Interface

synchronous runtime snapshot を dashboard/monitoring 向けに expose する場合、次を返すべきです。

- `running`
- `retrying`
- `codex_totals`
- `rate_limits`

running row は `turn_count` を含むべきです。recommended error modes は `timeout` と `unavailable` です。

### 13.4 Human-Readable Status Surface

terminal output や dashboard などの human-readable status surface は OPTIONAL かつ implementation-defined です。存在する場合は orchestrator state/metrics からのみ描画し、correctness に REQUIRED であってはなりません。

### 13.5 Session Metrics and Token Accounting

token accounting は absolute thread totals を優先します。

dashboard/API totals には、delta-style payload を直接使わないようにします。absolute totals では前回報告値との差分を追跡し、double-counting を避けます。

generic `usage` map は、event type が累積であると定義していない限り cumulative total と扱いません。aggregate totals は orchestrator state に accumulate します。

runtime は snapshot/render 時点の live aggregate として報告すべきです。ended session の cumulative counter に active-session elapsed time を足して snapshot/status view を生成してよいです。

rate-limit tracking は、任意の agent update で見た最新 payload を保持します。

### 13.6 Humanized Agent Event Summaries

raw agent protocol event の humanized summary は OPTIONAL です。実装する場合も observability-only output とし、orchestrator logic は humanized string に依存してはなりません。

### 13.7 OPTIONAL HTTP Server Extension

HTTP server extension は conformance に REQUIRED ではありません。dashboard/API は observability/control surface のみであり、orchestrator correctness に REQUIRED になってはなりません。

extension config:

- `server.port`: optional integer。HTTP server extension を enable する。`0` は local development/test 用 ephemeral port を要求する。CLI `--port` は `server.port` を override する。

server は CLI `--port` または `WORKFLOW.md` front matter の `server.port` がある場合に start します。

positive `server.port` はその port に bind し、明示設定がない限り loopback default bind を使うべきです。

listener setting の change は hot-rebind しなくても conformant で、restart-required でよいです。

#### 13.7.1 Human-Readable Dashboard (`/`)

`/` に human-readable dashboard を host します。active sessions、retry delays、token consumption、runtime totals、recent events、health/error indicators など current state を描画すべきです。

#### 13.7.2 JSON REST API (`/api/v1/*`)

runtime state と operational debugging のため `/api/v1/*` に JSON REST API を提供します。minimum endpoints:

- `GET /api/v1/state`: running sessions、retry queue/delays、aggregate token/runtime totals、latest rate limits など summary view を返す。
- `GET /api/v1/<issue_identifier>`: identified issue の runtime/debug details を返す。unknown issue は `404` と JSON error envelope。
- `POST /api/v1/refresh`: immediate tracker poll + reconciliation cycle を best-effort で queue する。

API は recommended baseline shape を保ち、field 追加は許容しますが、version 内の既存 field を壊すべきではありません。

defined route の unsupported method は `405 Method Not Allowed` を返すべきです。

API errors は `{"error":{"code":"...","message":"..."}}` のような JSON envelope を使うべきです。

## 14. Failure Model and Recovery Strategy

### 14.1 Failure Classes

1. `Workflow/Config Failures`: missing `WORKFLOW.md`、invalid YAML front matter、unsupported tracker kind、missing credentials/project slug、missing coding-agent executable。
2. `Workspace Failures`: directory creation failure、population/synchronization failure、invalid workspace path config、hook timeout/failure。
3. `Agent Session Failures`: startup handshake failure、turn failed/cancelled、turn timeout、user input requested handled as failure、subprocess exit、stalled session。
4. `Tracker Failures`: API transport errors、non-200 status、GraphQL errors、malformed payloads。
5. `Observability Failures`: snapshot timeout、dashboard render errors、log sink configuration failure。

### 14.2 Recovery Behavior

- dispatch validation failure は new dispatch を skip し、service を alive に保ち、可能な限り reconciliation を続けます。
- worker failure は exponential backoff retry に変換します。
- tracker candidate-fetch failure はその tick を skip し、次 tick で retry します。
- reconciliation state-refresh failure は current workers を維持します。
- dashboard/log failure は orchestrator を crash させません。

### 14.3 Partial State Recovery

current design は、scheduler state を意図的に in-memory とします。

restart recovery とは、tracker state を poll し、preserved workspace を reuse して有用な運用を再開できることを意味します。

次の state が process restart を越えて survive することは意味しません。

- retry timer
- running session
- live worker state

restart 後は retry timer も running session も復元しません。代わりに、startup terminal workspace cleanup、fresh polling of active issues、eligible work の re-dispatch により recover します。

Active issue は、対象 protocol が resume を support する場合、previous worker run の persisted coding-agent thread state を resume してもよいです。

Terminal issue は、cleanup 後に previous coding-agent state を resume してはなりません。後で reopened または active work として再作成された場合は、thread state なしで開始します。

### 14.4 Operator Intervention Points

operator は次の操作により behavior を制御できます。

- `WORKFLOW.md` の編集
- tracker 上の issue state change
- service restart

`WORKFLOW.md` change は Section 6.2 に従って自動 detect/re-apply されます。

terminal state は reconciled 時に running session stop と workspace cleanup、non-active state は cleanup なしの stop を引き起こします。

## 15. Security and Operational Safety

### 15.1 Trust Boundary Assumption

各実装は自身の trust boundary を定義します。

trusted environment 向けか、restrictive environment 向けか、また auto-approved actions、operator approvals、stricter sandboxing、あるいはその組み合わせに依存するかを明確に述べるべきです。

workspace isolation と path validation は重要な baseline control ですが、実装が選ぶ approval/sandbox policy の代替ではありません。

### 15.2 Filesystem Safety Requirements

mandatory:

- workspace path は configured workspace root 配下に留まらなければならない。
- coding-agent cwd は current run の per-issue workspace path でなければならない。
- workspace directory name は sanitized identifier を使わなければならない。

additional hardening:

- dedicated OS user で実行する。
- workspace root permission を制限する。
- 可能なら workspace root を dedicated volume に mount する。

### 15.3 Secret Handling

workflow config で `$VAR` indirection を support します。API token や secret env value を log してはなりません。secret の presence は値を print せず validate します。

### 15.4 Hook Script Safety

workspace hooks は `WORKFLOW.md` 由来の arbitrary shell scripts です。

fully trusted configuration として扱い、workspace directory 内で実行します。

hook output は log で truncate すべきであり、orchestrator hang を避けるため hook timeout は REQUIRED です。

### 15.5 Harness Hardening Guidance

repository、Issue Tracker、外部制御され得る input に対して Codex agent を実行することは危険を伴います。permissive deployment は data leak、destructive mutation、machine compromise につながり得ます。

実装は risk profile を明示的に評価し、必要に応じて execution harness を harden すべきです。

具体策には、次が含まれます。

- Codex approval/sandbox setting の tightening
- OS/container/VM sandboxing
- network restriction
- separate credentials
- eligible tracker source の filtering
- `linear_graphql` tool の scope narrowing
- agent に渡す tool/credential/filesystem/network destination の最小化

正しい control は deployment-specific ですが、実装はそれを明確に文書化し、harness hardening を optional afterthought ではなく core safety model の一部として扱うべきです。

## 16. Reference Algorithms

### 16.1 Service Startup

startup は logging と observability output を構成し、workflow watch を開始し、in-memory state を初期化します。dispatch config を validate し、失敗時は startup を fail します。startup terminal workspace cleanup を実行し、delay 0 の tick を schedule して event loop に入ります。

### 16.2 Poll-and-Dispatch Tick

tick では running issues を reconcile し、dispatch config を validate します。validation failure や tracker fetch failure は log と notify を行い、次 tick を schedule して戻ります。candidate issues を sort し、slot がある限り eligible issue を dispatch します。

### 16.3 Reconcile Active Runs

stall run を reconcile し、running issue ID がなければ戻ります。tracker state refresh が失敗した場合は worker を running のままにします。terminal state は cleanup ありで terminate、active state は running entry の issue を update、それ以外は cleanup なしで terminate します。

### 16.4 Dispatch One Issue

worker spawn に成功したら `running` entry を作成し、session/token/retry/runtime fields を初期化し、issue ID を `claimed` に追加して retry attempt を remove します。spawn failure は retry schedule に変換します。

### 16.5 Worker Attempt

worker attempt は workspace create/reuse、`before_run` hook、app-server session startup、prompt build、turn execution、issue state refresh、continuation decision、session stop、`after_run` hook を行います。各 failure は worker failure として扱います。

### 16.6 Worker Exit and Retry Handling

worker exit では running entry を remove し runtime seconds を totals に加えます。normal exit は bookkeeping の `completed` に加え continuation retry attempt 1 を schedule します。abnormal exit は next attempt で retry を schedule します。retry timer は candidate fetch、issue lookup、slot check を行い、dispatch または requeue/release します。

## 17. Test and Validation Matrix

conforming implementation はこの仕様の behavior を cover する tests を含むべきです。

Validation profiles:

- `Core Conformance`: すべての conforming implementation に REQUIRED な deterministic tests。
- `Extension Conformance`: 実装が ship する OPTIONAL feature にのみ REQUIRED。
- `Real Integration Profile`: production use 前に RECOMMENDED な environment-dependent smoke/integration checks。

### 17.1 Workflow and Config Parsing

workflow path precedence、workflow reload、invalid reload の last-known-good 維持、missing/invalid workflow error、front matter non-map error、defaults、`tracker.kind` validation、`tracker.api_key` と `$VAR` resolution、path expansion、Codex command preservation、per-state concurrency normalization、prompt rendering strictness を test します。

### 17.2 Workspace Manager and Safety

deterministic workspace path、directory creation/reuse、non-directory collision handling、workspace population error、hooks の execution/failure semantics、workspace sanitization/root containment、agent cwd validation を test します。

### 17.3 Issue Tracker Client

candidate fetch active states/project slug、Linear `slugId` filter、empty state fetch shortcut、pagination、blocker normalization、label lowercase、state refresh by ID、GraphQL ID typing、error mapping を test します。

### 17.4 Orchestrator Dispatch, Reconciliation, and Retry

dispatch sort、Todo blocker rule、active/non-active/terminal reconciliation、normal/abnormal worker exit retry、backoff cap、retry queue entry contents、stall detection、slot exhaustion requeue、snapshot API behavior を test します。Terminal state の test は running agent stop、workspace cleanup、thread/rollout artifact removal、resume pointer invalidation を確認します。

### 17.5 Coding-Agent App-Server Client

次の項目を test します。

- workspace cwd launch
- targeted Codex protocol startup
- policy payload
- thread/turn identity extraction
- timeouts
- transport framing
- stderr separation
- approval/user-input policy
- unsupported tool call handling
- usage/rate-limit extraction
- client-side tools と `linear_graphql` extension

Worker run 間の retry/resume では、次のことを test しなければなりません。

- 後続 process が previous `thread_id` を targeted resume request で受け取る。
- 同じ workspace `cwd` を使う。
- original issue prompt ではなく continuation guidance で next turn を開始する。

Terminal cleanup では、次のことを test しなければなりません。

- workspace-local thread/rollout file を削除する。
- resume pointer を invalidate する。
- cleanup 済み terminal issue の後続 dispatch が stale `thread_id` を受け取らないことを確認する。

### 17.6 Observability

operator-visible validation failure、structured log context、log sink failure resilience、token/rate-limit aggregation、status surface と humanized event summaries が correctness に影響しないことを test します。

### 17.7 CLI and Host Lifecycle

positional workflow path、default `./WORKFLOW.md`、explicit/default path error、startup failure surface、normal/abnormal exit code を test します。

### 17.8 Real Integration Profile

valid credentials と network access がある場合、real tracker smoke test を run することが RECOMMENDED です。isolated test identifiers/workspaces を使い、可能なら tracker artifacts を cleanup します。skipped real-integration test は skipped として報告し、明示的に CI/release validation で有効化した場合の failure は job failure とすべきです。

## 18. Implementation Checklist

### 18.1 REQUIRED for Conformance

- explicit runtime path と cwd default の workflow path selection。
- YAML front matter + prompt body split を持つ `WORKFLOW.md` loader。
- defaults と `$` resolution を持つ typed config layer。
- config と prompt の dynamic watch/reload/re-apply。
- single-authority mutable state を持つ polling orchestrator。
- candidate fetch、state refresh、terminal fetch を持つ Issue Tracker client。
- sanitized per-issue workspace を持つ workspace manager。
- workspace lifecycle hooks と timeout config。
- JSON line protocol の coding-agent app-server subprocess client。
- `codex.command` config、strict prompt rendering、retry queue、reconciliation、terminal workspace cleanup。
- terminal issue の thread/rollout cleanup と resume pointer invalidation。
- `issue_id`, `issue_identifier`, `session_id` を含む structured logs。
- operator-visible observability。

### 18.2 RECOMMENDED Extensions

- HTTP server extension は CLI `--port` を `server.port` より優先し、安全な default bind host を使い、Section 13.7 の baseline endpoints/error semantics を expose する。
- `linear_graphql` client-side tool extension は configured Symphony auth で Linear GraphQL access を app-server session に expose する。
- TODO: process restart を越えた retry queue と session metadata persistence。
- TODO: UI 実装を規定せず workflow front matter で observability settings を configurable にする。
- TODO: agent tools だけでなく orchestrator に first-class tracker write APIs を追加する。
- TODO: Linear 以外の pluggable Issue Tracker adapters。

### 18.3 Operational Validation Before Production

- valid credentials と network access で Section 17.8 の Real Integration Profile を run する。
- target host OS/shell environment で hook execution と workflow path resolution を verify する。
- OPTIONAL HTTP server を ship する場合、configured port behavior と loopback/default bind expectation を target environment で verify する。

## Appendix A. SSH Worker Extension

この appendix は、central orchestrator を 1 つ保ちながら worker run を SSH 経由で remote host 上で実行する extension profile を説明します。

extension config:

- `worker.ssh_hosts`: SSH host string list。省略時は local execution。
- `worker.max_concurrent_agents_per_host`: configured SSH hosts 全体に適用される optional per-host cap。

### A.1 Execution Model

orchestrator は polling、claims、retries、reconciliation の single source of truth のままです。`worker.ssh_hosts` は remote execution の candidate destinations を提供します。各 worker run は 1 host に割り当てられ、その host は issue workspace とともに execution identity の一部になります。`workspace.root` は orchestrator host ではなく remote host 上で解釈されます。coding-agent app-server は local subprocess ではなく SSH stdio 経由で起動されますが、session lifecycle は orchestrator が所有します。

### A.2 Scheduling Notes

SSH hosts は dispatch pool として扱ってよいです。retry では可能なら以前使った host を prefer してよいです。全 host が capacity 上限の場合、別 execution mode に silently fallback するのではなく dispatch を待つべきです。作業が meaningful side effect を既に生んでいる場合、別 host での rerun は invisible failover ではなく new attempt として扱うべきです。

### A.3 Problems to Consider

- remote environment drift: host ごとに shell、coding-agent executable、auth、repository prerequisite が必要です。
- workspace locality: workspace は通常 host-local なので、別 host への移動は shared storage がない限り cold restart です。
- path and command safety: remote path resolution、shell quoting、workspace boundary checks がより重要になります。
- startup and failover semantics: host connectivity/startup failure と in-workspace agent failure を区別すべきです。
- host health and saturation: dead/overloaded host は capacity を減らすべきで、duplicate execution や accidental fallback を起こしてはなりません。
- cleanup and observability: operator は run の host、workspace location、cleanup の成否を把握できる必要があります。
