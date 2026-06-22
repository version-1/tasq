# Codex App-Server 契約

Tasq は `bash -lc <codex.command>` で Codex app-server を local subprocess として起動し、process
working directory には issue ごとの workspace を使います。Repository workflow は `codex.command`
を `codex --sandbox workspace-write app-server` に設定し、development container 内で app-server が
workspace-write sandbox posture で起動するようにします。

Installed Codex CLI schema が protocol の source of truth です。Runner の message shape は次の
command で生成される schema に基づきます。

```sh
codex app-server generate-json-schema --out <dir>
```

Runner は Codex orchestration と transport framing を分離します。Orchestration layer は JSON-RPC
request ID、response correlation、thread/turn sequencing、approval denial policy、runner event
emission を所有します。Transport implementation は connection lifecycle と byte-frame send/receive
operation だけを提供します。

現在の production wiring は引き続き stdio subprocess transport を起動します。Runtime transport
selection は最初の websocket implementation の scope 外です。

Stdio transport は v2 app-server JSON-RPC line protocol を使います。

- `clientInfo` と `experimentalApi` を含む `initialize`。
- initialize 成功後の `initialized` notification。
- Absolute workspace path を `cwd` として渡し、後続 run が thread ID で resume できるよう
  runner task に resume thread ID が無い場合に persistent thread (`ephemeral: false`) を作る
  `thread/start`。
- Runner task に previous run から取得した resume thread ID がある場合、stored `threadId` と
  absolute workspace path を `cwd` として渡す `thread/resume`。
- 同じ `cwd`、返された `threadId`、1 つの text input を渡す `turn/start`。New thread には
  rendered workflow prompt を渡します。Resumed thread には original issue prompt ではなく
  continuation guidance を渡します。
- Success として扱う `turn/completed` notification。
- Failure として扱う `error`、subprocess exit、context cancellation、response timeout、turn
  timeout。

Process lifetime は引き続き run-scoped です。Tasq は successful run、failed run、timed-out run
を含め、各 runner run の終了時に app-server subprocess を close します。したがって retry をまたぐ
resume は、issue lifetime のために long-lived worker process を保持するのではなく、stored thread ID
に別 process から reconnect する動作です。

Persistent thread と rollout state は workspace-scoped runtime state として扱います。Terminal issue
cleanup は per-issue workspace を remove し、その issue の runstore resume pointer を invalidate
しなければなりません。これにより、後続 dispatch が workspace-local artifact 削除済みの thread に
reconnect できないようにします。Non-terminal retry は issue の latest stored thread ID を再利用してよいです。

この lifecycle の verification coverage は次の通りです。

- `internal/orchestrator/runner`: resumed run が stored `threadId` で `thread/resume` を送信し、
  workspace `cwd` を使い、`session_started` を emit し、continuation guidance で turn を開始する。
- `internal/orchestrator/coordinator`: dispatch が latest persisted thread ID を retry task に渡し、
  新しく emit された `thread_id` を永続化する。
- Terminal cleanup tests は cleanup contract の両面を assert しなければなりません。
  workspace-local thread/rollout artifact が remove され、同じ issue の後続 dispatch が stale
  resume thread ID を受け取らないことを確認します。

Websocket transport は OpenAI Codex app-server 契約
<https://developers.openai.com/codex/app-server/> に従います。

- `--listen ws://IP:PORT` endpoint で app-server を起動し、その URL に接続します。
- Upstream では websocket transport は experimental かつ unsupported です。
- この phase で Tasq が対象にするのは `ws://127.0.0.1:PORT` のような local listener です。
- この phase では authentication header、subprotocol、Origin header は追加しません。
- 各 WebSocket text frame は 1 つの JSON-RPC message を含みます。
- Connection close、frame read failure、context deadline は JSON-RPC session layer に transport
  error として渡します。
- 同じ listener は `/readyz` と `/healthz` も公開しますが、初期実装の Tasq client は health probe
  に依存しません。

Tasq は現在、approval や dynamic tool request のような server-to-client request を unsupported
JSON-RPC error として扱います。これにより、unattended run が operator input を無期限に待つことを防ぎます。

Command-execution と file-change approval request は既知の approval request です。Tasq は即座に
`{"decision":"cancel"}` で応答し、approval method と raw request payload を含む
`approval_required` error で run を failed にします。その後 dispatcher は、まだ ready または
in-progress の issue を `blocked` にし、operator が将来の retry で requested action を許可すべきか
判断できるように blocker comment に詳細を書き込みます。Approval、sandbox、user-input behavior は、
より広い production usage を有効にする前に見直すべきです。

Runner progress は SQLite の `runner_events` に保存されます。現在の implementation は filesystem に
separate large transcript artifact を書きません。
