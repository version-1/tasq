# Codex App-Server 契約

Tasq は `bash -lc <codex.command>` で Codex app-server をローカルサブプロセスとして起動し、プロセスの
作業ディレクトリには課題ごとのワークスペースを使います。リポジトリのワークフローは `codex.command`
を `codex --sandbox workspace-write app-server` に設定し、開発コンテナ内で app-server が
workspace-write sandbox posture で起動するようにします。

インストール済み Codex CLI schema がプロトコルの正とする情報源です。Runner のメッセージ形式は、
次のコマンドで生成される schema に基づきます。

```sh
codex app-server generate-json-schema --out <dir>
```

再開サポートは Codex CLI 0.141.0 から生成した v2 schema に基づきます。`thread/resume` は
必須の `threadId` 文字列を受け取ります。Schema は in-memory history と rollout path による
再開入力もサポートしますが、schema は可能な限り thread ID を使うことを推奨しており、runstore
もその識別子を永続化するため、Tasq は `threadId` を使います。`ThreadResumeResponse` は
`ThreadStartResponse` と同じ `thread` response shape を持つため、runner は start/resume のどちらでも
返された `thread.id` を後続の turn path に使います。

Runner は Codex のオーケストレーションと transport framing を分離します。オーケストレーション層は
JSON-RPC request ID、response correlation、thread/turn sequencing、approval denial policy、runner event
の出力を所有します。Transport 実装は connection lifecycle と byte-frame の送受信操作だけを提供します。

現在の本番向け wiring は引き続き stdio subprocess transport を起動します。実行時の transport
選択は最初の websocket 実装のスコープ外です。

Stdio transport は v2 app-server JSON-RPC line protocol を使います。

- `clientInfo` と `experimentalApi` を含む `initialize`。
- initialize 成功後の `initialized` notification。
- Absolute workspace path を `cwd` として渡し、後続 run が thread ID で resume できるよう
  runner task に resume thread ID が無い場合に永続 thread (`ephemeral: false`) を作る
  `thread/start`。
- Runner task に previous run から取得した resume thread ID がある場合、保存済みの `threadId` と
  absolute workspace path を `cwd` として渡す `thread/resume`。
- 同じ `cwd`、返された `threadId`、1 つの text input を渡す `turn/start`。New thread には
  レンダリング済みワークフロープロンプトを渡します。Resumed thread には元の課題プロンプトではなく
  継続用の指示を渡します。
- 成功として扱う `turn/completed` notification。
- 失敗として扱う `error`、subprocess exit、context cancellation、response timeout、turn timeout。

プロセスの存続期間は引き続き run-scoped です。Tasq は successful run、failed run、timed-out run
を含め、各 runner run の終了時に app-server subprocess を閉じます。したがって retry をまたぐ
再開は、課題の存続期間のために long-lived worker process を保持するのではなく、保存済み thread ID
に別プロセスから再接続する動作です。

`ephemeral: false` は Tasq の再開契約に必須です。Codex は `ephemeral` を、thread を disk に
materialize するかどうかとして定義しており、ephemeral thread は後続 subprocess から `threadId` で
load できません。そのため new thread は永続 thread として作成し、retry は課題が非終端の間だけ
`thread/resume` を使います。

Persistent thread と rollout state は、ワークスペーススコープの実行時状態として扱います。終端課題の
クリーンアップは、課題ごとのワークスペースを削除し、その課題の runstore resume pointer を無効化
しなければなりません。これにより、後続の割り当てがワークスペースローカル artifact 削除済みの thread に
再接続できないようにします。非終端の retry は、課題の最新の保存済み thread ID を再利用してよいです。

このライフサイクルの検証範囲は次の通りです。

- `internal/orchestrator/runner`: resumed run が stored `threadId` で `thread/resume` を送信し、
  workspace `cwd` を使い、`session_started` を発行し、continuation guidance で turn を開始する。
- `internal/orchestrator/coordinator`: dispatch が最新の永続化済み thread ID を retry task に渡し、
  新しく発行された `thread_id` を永続化する。
- Terminal cleanup tests は cleanup contract の両面を assert しなければなりません。
  workspace-local thread/rollout artifact が削除され、同じ課題の後続 dispatch が古い
  resume thread ID を受け取らないことを確認します。

Websocket transport は OpenAI Codex app-server 契約
<https://developers.openai.com/codex/app-server/> に従います。

- `--listen ws://IP:PORT` endpoint で app-server を起動し、その URL に接続します。
- upstream では websocket transport は experimental かつ unsupported です。
- この phase で Tasq が対象にするのは `ws://127.0.0.1:PORT` のようなローカル listener です。
- この phase では authentication header、subprotocol、Origin header は追加しません。
- 各 WebSocket text frame は 1 つの JSON-RPC message を含みます。
- Connection close、frame read failure、context deadline は JSON-RPC session layer に transport
  error として渡します。
- 同じ listener は `/readyz` と `/healthz` も公開しますが、初期実装の Tasq client は health probe
  に依存しません。

Tasq は現在、approval や dynamic tool request のような server-to-client request を unsupported
JSON-RPC error として扱います。これにより、無人実行が運用者入力を無期限に待つことを防ぎます。

Command-execution と file-change approval request は既知の approval request です。Tasq は即座に
`{"decision":"cancel"}` で応答し、approval method と raw request payload を含む
`approval_required` error で run を failed にします。その後 dispatcher は、まだ ready または
in-progress の課題を `blocked` にし、運用者が将来の retry で requested action を許可すべきか
判断できるように blocker comment に詳細を書き込みます。Approval、sandbox、user-input の振る舞いは、
より広い本番利用を有効にする前に見直すべきです。

Runner progress は SQLite の `runner_events` に保存されます。現在の実装は filesystem に
別個の大きなトランスクリプト artifact を書きません。
