# Tasq Symphony 差分

このドキュメントは、Tasq と上流の Symphony サービス仕様
[SPEC.md](SPEC.md) の意図的な差分を記録します。

Tasq は、オーケストレーション、ワークスペース、ワークフロー、エージェント実行器、トラッカー、
可観測性の方向性として Symphony 仕様を使います。ただし Tasq には、課題状態を所有するローカルの
issue-tracker サービスが既にあるため、一部の実装方針は異なります。

## Tracker Adapter

Tasq は Symphony 仕様の Section 11 で説明されている Linear トラッカークライアントを実装しません。

代わりに、Tasq はローカルの issue-tracker API をトラッカーアダプター境界として扱います。

- issue-tracker は課題状態とプロジェクトデータを所有します。
- orchestrator はワークスペースレコード、メタデータ、ライフサイクルの振る舞いを所有します。
- issue-tracker はトラッカーアダプターからの読み取り用に、課題一覧と課題状態問い合わせのエンドポイントを公開します。
- orchestrator は過去の実行と runner event のデータを自身の SQLite ストアに保持します。

これにより、外部トラッカー連携を orchestrator から外し、リポジトリの既存サービス境界
を維持します。

## Workflow Front Matter Contract

Tasq の `WORKFLOW.md` front matter は、Symphony の front matter スキーマの上に重ねた、
小さな Tasq 固有の契約として意図的に定義されています。Tasq の正式なワークフロー契約は
[WORKFLOW_CONTRACT.md](WORKFLOW_CONTRACT.md) に記録されています。次の表は、Symphony `SPEC.md`
のフィールドと Tasq の front matter の振る舞いの対応を示します。

| トップレベルキー | 子フィールド | Symphony support | Tasq support | Tasq extension | Tasq の front matter / 振る舞い |
| --- | --- | --- | --- | --- | --- |
| `tracker` | `kind` | ✓ | ✓ | × | 現在のオーケストレーションは Linear クライアントではなくローカルの issue-tracker API から読み取ります。 |
|  | `endpoint` | ✓ | ✓ | × | Symphony のトラッカー形状との部分的な互換性のために解析されます。 |
|  | `api_key` | ✓ | ✓ | × | `$VAR` の解決をサポートします。 |
|  | `project_slug` | ✓ | ✓ | × | `tracker.kind` が設定された場合のみ必須です。 |
|  | `active_states` | ✓ | ✓ | × | Tasq の割り当て適格性は、ローカルの issue-tracker 状態で決まります。 |
|  | `terminal_states` | ✓ | ✓ | × | Tasq のクリーンアップと整合処理は、ローカルの issue-tracker 状態を使います。 |
| `polling` | `interval_ms` | ✓ | ✓ | × | Orchestrator のポーリング間隔です。 |
| `workspace` | `root` | ✓ | ✓ | × | 選択された `WORKFLOW.md` からの相対パスとして解決され、worktree 管理のため対象 Git リポジトリ内にある必要があります。 |
|  | `source` | × | × | × | Tasq は代わりに `workspace.root` 配下に Git worktree を作成します。 |
| `hooks` | `after_create` | ✓ | ✓ | × | 新規作成されたワークスペースに対してのみ、課題ワークスペース内で `bash -lc` 経由で実行されます。 |
|  | `before_run` | ✓ | ✓ | × | 各エージェント試行の前に実行されます。 |
|  | `after_run` | ✓ | ✓ | × | 各エージェント試行後に実行され、失敗はログに記録されて無視されます。 |
|  | `before_remove` | ✓ | ✓ | × | ワークスペースディレクトリが存在する場合にクリーンアップ前に実行され、失敗はログに記録されますがクリーンアップは継続します。 |
|  | `timeout_ms` | ✓ | ✓ | × | すべてのワークスペース hook に適用され、正の値である必要があります。 |
| `agent` | `max_concurrent_agents` | ✓ | ✓ | × | 全体の並行実行数の上限です。 |
|  | `max_turns` | ✓ | ✓ | × | 1 回の実行における Codex ターン数の上限です。継続ターンには `agent.continuation_turns_enabled` も必要です。 |
|  | `max_retry_backoff_ms` | ✓ | ✓ | × | 再試行バックオフの上限です。 |
|  | `max_concurrent_agents_by_state` | ✓ | ✓ | × | 正規化された map です。意味を持つかは、ローカルの issue-tracker 状態モデルに依存します。 |
|  | `continuation_turns_enabled` | × | ✓ | ✓ | `agent.max_turns` が 1 より大きい場合でも、継続ターンをこの値で制御します。 |
|  | `max_retry_attempts` | × | ✓ | ✓ | 実行を再試行する回数を設定します。 |
| `codex` | `command` | ✓ | ✓ | × | リポジトリルートの `WORKFLOW.md` は `codex --sandbox workspace-write app-server` を使います。 |
|  | `approval_policy` | ✓ | ✓ | × | リポジトリルートの `WORKFLOW.md` では設定されていません。 |
|  | `thread_sandbox` | ✓ | ✓ | × | リポジトリルートの `WORKFLOW.md` では設定されていません。 |
|  | `turn_sandbox_policy` | ✓ | ✓ | × | リポジトリルートの `WORKFLOW.md` では設定されていません。 |
|  | `turn_timeout_ms` | ✓ | ✓ | × | Codex ターンのタイムアウトです。 |
|  | `read_timeout_ms` | ✓ | ✓ | × | Codex app-server の読み取りタイムアウトです。 |
|  | `stall_timeout_ms` | ✓ | ✓ | × | Tasq は正の値として検証します。 |
| `server` | `port` | ✓ | ✓ | × | 任意の HTTP 拡張のポートです。CLI `--port` で上書きできます。 |
| `tasq` | `task_work_prompt` | × | ✓ | ✓ | orchestrator がレンダリング済みプロンプトの前に既定の `tq` issue-tracker 同期指示を付与するかを制御します。 |

表の補足:

- `Symphony Support` は、そのフィールドが `SPEC.md` のコアフィールドまたは文書化済み拡張フィールドとして定義されていることを表します。
- `Tasq Support` は、Tasq がそのフィールドを `WORKFLOW.md` の front matter 契約で解析または実装していることを表します。
- `Tasq Extension` は、そのフィールドが Tasq 固有であり、Symphony コアスキーマの一部ではないことを表します。
- `tracker.kind` は Symphony の割り当てでは必須です。Tasq は解析しますが、ローカルの割り当てでは使いません。
- Symphony では `codex.stall_timeout_ms <= 0` によって停止検知を無効化できます。Tasq はこのフィールドを正の値として検証します。

Symphony との差分:

- 有効な `WORKFLOW.md` は、課題のキュー投入または割り当て時にプロジェクトごとに解決されます。
  すでに実行中の作業に対する動的な監視と再読み込みは延期されています。
- 未知の front matter フィールドは、前方互換性のために無視されます。
- `workspace.source` はサポートされません。Tasq は `workspace.root` 配下に Git worktree で課題ごとの
  ワークスペースを作成します。
- 大きなトランスクリプト artifact のパスと可観測性の出力先は、workflow front matter では設定できません。
  Tasq は runner の進捗、ワークスペースメタデータ、ワークスペース準備の失敗、クリーンアップ状態を
  orchestrator の SQLite データベースに記録します。
- ローカルの `tq project check` コマンドは、Symphony スキーマ全体ではなく、Tasq の既定プロジェクト
  テンプレートが要求する front matter フィールドを検証します。
- ワークフローパスの選択では、プロセスレベルの明示的なワークフローパスや cwd の既定値を使いません。
  Tasq は "Workflow Path Selection" に記載するように、プロジェクトごとに有効なワークフローを解決します。
- Codex app-server のオーケストレーションは内部的に transport に依存せず、Tasq は stdio と websocket の
  transport パッケージを含みます。本番のワークフロー実行では、まだ stdio subprocess transport を
  起動します。実行時の transport 選択と実 Codex websocket server との結合検証は延期されています。

## Workspace Key

Symphony は `MT-649` のような `issue.identifier` からワークスペースキーを導出します。
Tasq は正式な課題識別子として `issue-<ID>` を意図的に使い、ワークスペース名にも同じ値を
使います。例: `issue-42`。

理由:

- Tasq のローカル issue-tracker は、安定した数値 ID を持つ課題を所有します。
- Tasq は現在、人間が読める外部トラッカー識別子を別フィールドとしてモデル化していません。
- `issue-<ID>` を使うことで、Linear 固有のフィールドをローカルの課題契約に追加せずに、ワークスペースパスを
  決定的に保てます。

同じワークスペースキー規約は、ワークスペースメタデータ、起動時の終端状態クリーンアップ、
アクティブ実行の整合処理でも使われます。

## Current Implementation Gap

現在の orchestrator は Symphony 準拠に向けて段階的に移行しています。まだ完全な Symphony
実装ではありません。

実装済み、または進行中:

- Symphony front matter の小さなサポート対象サブセットを使ったワークフローファイルの読み込み。
- 有効なワークフローの解決は、プロジェクトごとにキュー投入または割り当て時点で行われます。
  すでに実行中の作業に対する実行時の再読み込みは意図的に延期されています。
- ワークスペースルートの解決と、サニタイズ済みの課題ごとのワークスペースディレクトリ。
- `hooks.timeout_ms` を含むワークスペースライフサイクル hook。
- シミュレーション実装と Codex app-server subprocess 実装を持つ runner interface。
- stdio と websocket transport 実装を持つ Codex app-server transport 契約。
- SQLite への runner event ログ記録とワークスペースメタデータレコード。
- live Codex app-server thread に対する、設定で制御される継続ターン。
- 新しい app-server subprocess を起動し、永続化済み thread state に再接続する worker run 間の
  再試行と再開。
- 上限付き指数バックオフを使う in-process の再試行スケジューリング。
- 終端または非アクティブな課題状態と停止処理のための、アクティブ実行の整合処理。
- 初回ワークスペース作成時の Git worktree ワークスペース作成。
- クリーンアップメタデータを伴う、終端および failed/cancelled ワークスペースのクリーンアップ。
  古い thread/rollout artifact のクリーンアップも含みます。
- ワークスペース準備失敗に関する、運用者向けログ。

解消済みの実装差分:

- 以前の Codex runner 実装は `ephemeral: true` で app-server thread を作成していたため、
  thread が disk に materialize されず、後続の `threadId` による再開ができませんでした。Tasq は現在、
  永続 Codex thread を `ephemeral: false` で作成し、返された `thread_id` を永続化し、
  適格な非終端の再試行で `thread/resume` を使います。

未実装:

- 動的な `WORKFLOW.md` 再読み込み。
- すべての変数と filter の検査を伴う厳密なプロンプトレンダリング。
- トークンと rate limit の集計。
- stdio と websocket の Codex app-server transport の実行時選択。
- 実 Codex websocket app-server との結合検証。
- 任意の Symphony HTTP status/API surface 全体。

Tasq は [WORKFLOW_CONTRACT.md](WORKFLOW_CONTRACT.md) に記録されている workflow front matter フィールド
をサポートします。未知のフィールドは前方互換性のために無視されます。

## Workspace Creation Strategy

Tasq はリポジトリのソースディレクトリからファイルをコピーするのではなく、`git worktree add` で
課題ごとの orchestrator ワークスペースを作成します。

理由:

- コーディングエージェントには、課題ごとのワークスペースが Git リポジトリルートである必要があります。
- `.git` なしでファイルをコピーすると、リポジトリ調査コマンドが別の Git ルートを観測し、編集が
  課題ワークスペースではなく親リポジトリを対象にする可能性があります。
- worktree は Git メタデータを維持しつつ、課題プロジェクトの `Project.Location` と設定された相対
  `workspace.root` 配下に決定的なワークスペースパスを保ちます。

`workspace.source` workflow field は意図的にサポートされません。Workspace manager が
相対 workspace-root suffix を解決できるように、`workspace.root` は orchestrator project の Git
リポジトリ内になければなりません。Tasq は割り当てられた課題ごとに `Issue.ProjectID` をローカル
issue-tracker API で解決し、同じ相対 suffix を使って参照先プロジェクトの location 配下に worktree
を作成します。例: `<Project.Location>/.worktrees/agents/issue-42`。

ワークスペースブランチは `agent/<workspace-key>` を使います。例: `agent/issue-42`。クリーンアップは
`git worktree remove --force` を使い、対応するローカルブランチを best-effort で削除します。

## Codex Thread Resume Lifecycle

Symphony は、1 回の worker run 内にある live Codex app-server thread に対する継続ターンを説明します。
Tasq はこのライフサイクルを拡張し、別々の worker run をまたいで適格な非終端作業を再開できるだけの
thread state を永続化します。

Tasq が以前の worker run を再開するときは、新しい app-server subprocess を起動し、永続化済み
`thread_id` を使って再接続し、同じワークスペース `cwd` を維持し、元の課題プロンプトではなく
継続用の指示を送ります。以前の app-server subprocess は課題の存続期間に紐づくプロセスとして
扱われません。Worker exit は常にそれを閉じます。

永続化された `thread_id` 値は、課題が非終端の間だけ再利用できます。終端課題のクリーンアップでは、
永続化済み thread/rollout state と orchestrator が所有する再開ポインターを含む、ワークスペーススコープの
コーディングエージェント artifact を削除します。同じ課題が後で reopen される、またはアクティブな作業として
再作成される場合、割り当ては古い thread state なしで開始します。

SPEC.md の以下のセクションと乖離します:

| Section | Symphony の前提 | Tasq の振る舞い |
| --- | --- | --- |
| §7.7, §8.1 | 終端クリーンアップは古いワークスペースディレクトリを中心に扱う | 終端クリーンアップは thread/rollout artifact も削除し、再開ポインターを無効化する |
| §10.2–10.3 | 継続状態は 1 回の worker run 内にある live app-server subprocess にスコープされる | Worker run 間の再試行と再開は新しい subprocess を起動し、永続化済み thread state を通じて再接続する |
| §14.3 | 再起動時の回復は、ワークスペースクリーンアップ後に適格な作業を再割り当てする | アクティブ課題は永続化済み Codex thread state を再開できる。終端課題は古い状態を再開しないようクリーンアップされる |
| §17.1, §17.6, §18 | 準拠確認はワークスペースクリーンアップと live-thread continuation を対象にする | Tasq は cross-worker resume と終端 thread/rollout cleanup も検証する |

## Workflow Path Selection

Symphony はプロセスレベルのワークフローパス選択モデルを定義しています。実行時は明示的な
ワークフローパスを受け取ることができ、指定がなければ現在のプロセス作業ディレクトリの
`WORKFLOW.md` を既定値とします。Tasq は Symphony の workflow-path flag を公開しません。これには
以前の `--workflow` flag 形式も含みます。

Tasq はワークフロー設定を orchestrator process 単位ではなくプロジェクト単位で解決します。課題を
割り当てるとき、orchestrator は次の順序で有効なワークフローを選択します。

1. 課題の `Project.Location` 配下にある物理的なプロジェクト `WORKFLOW.md` ファイル。
2. ローカルの issue-tracker データベースのプロジェクトレコードに保存されたワークフロー内容。
3. Orchestrator が使うグローバルな fallback ワークフロー。

つまり orchestrator process の cwd は、割り当て済み課題のワークフローの振る舞いに対する既定のソース
ではありません。cwd は運用者コマンドとプロセス起動には引き続き関係しますが、課題の割り当ては課題に
紐づくプロジェクトを使います。

SPEC.md の以下のセクションと乖離します:

| Section | Symphony の前提 | Tasq の振る舞い |
| --- | --- | --- |
| §5.2, §6.1 | 明示的な実行時ワークフローパス。未指定時は cwd の `WORKFLOW.md` | `--workflow` によるパス選択はない。プロジェクトごとにワークフローを解決する |
| §17.7 | CLI は位置引数のワークフローパスを受け取り、`./WORKFLOW.md` に fallback する | Orchestrator CLI は全プロジェクト用の単一ワークフローファイルを選択しない |
| §18 | ワークフローパス選択は、明示的な実行時パスと cwd の既定値をサポートする | 有効なワークフローは課題の割り当て時に解決される |

## Multi-Project Orchestration

Symphony は 1 プロセスが 1 プロジェクトを担当する前提です（1 つの `WORKFLOW.md`、1 つの
`tracker.project_slug`、1 つの `workspace.root`）。Tasq は単一の orchestrator process で複数の
プロジェクトを扱います。

そのため Symphony モデルは、ワークフロー設定を orchestrator instance に対して選択された単一
ファイルとして扱います。Tasq はワークフロー設定をプロジェクトデータとして扱います。複数のプロジェクト
が独立したワークフローを持つことができ、orchestrator は特定の課題を割り当てるときにだけ
該当ワークフローを解決します。

Orchestrator 自体はプロジェクトを意識しません。単一の呼び出しでローカル issue-tracker から全プロジェクトの
適格な課題をポーリングし、同じ並行実行プールで割り当て、実行時状態（`claimed`、`running`、
`retry_attempts`）を課題 ID をキーとする flat map で管理します。

プロジェクト固有の振る舞いは、割り当て時に課題単位で解決されます。

- **ワークスペースパス**: `Issue.ProjectID → Project.Location` と相対 `workspace.root` suffix
  で解決されます（上記 "Workspace Creation Strategy" に記載）。
- **プロンプトと hook**: 各プロジェクトが自身の `WORKFLOW.md` を所有します。Orchestrator は該当課題の
  プロンプト構築と hook 実行時に、プロジェクトローカルのワークフローファイルを読み込みます。
- **ポーリング**: 単一の poll tick で全プロジェクトの候補を取得します。プロジェクトごとの
  `polling.interval_ms` はサポートされません。Orchestrator は 1 つのグローバル間隔を使います。
- **並行数**: `agent.max_concurrent_agents` は全プロジェクト横断のグローバル上限です。プロジェクトごとの
  並行数上限は現在サポートされていません。

SPEC.md の以下のセクションと乖離します:

| Section | Symphony の前提 | Tasq の振る舞い |
| --- | --- | --- |
| §5.1–5.2 | 1 プロセスに 1 つの `WORKFLOW.md` | プロジェクトごとに 1 つの `WORKFLOW.md`、課題単位で解決 |
| §5.3.1 | `tracker.project_slug`（単数） | 使用しない。issue-tracker が全プロジェクトの課題を返す |
| §4.1.8, §8.3 | 実行時状態と並行数は暗黙的に単一プロジェクトにスコープされる | flat なグローバル状態。プロジェクトごとの分割なし |
| §8.1 | 1 つのワークフローからの 1 つの poll interval | 1 つのグローバルな poll interval。プロジェクトごとの interval は無視 |

## Compatibility Notes

Symphony がスケジューリングに関して "tracker" と言う箇所は、将来の設計で外部トラッカーアダプター
を明示的に追加しない限り、Tasq の orchestrator ではローカルの issue-tracker API と読み替えます。

Symphony が Linear 固有のクエリ意味論を説明している箇所は、ローカルの issue-tracker 境界が選択された
トラッカーアダプターである間、orchestrator には適用されない要件として扱います。
