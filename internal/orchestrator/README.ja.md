# Orchestrator パッケージ

orchestrator は、agent run の状態、runner イベント、workspace セットアップのメタデータ、および任意のローカル runtime inspection を管理します。

`cmd/orchestrator` が composition root です。このディレクトリ配下のパッケージを組み立てる役割を持ち、flags、signal handling、起動時のログといった process レベルの関心事はここに置きます。

## パッケージ境界

### `run`

orchestrator run のドメインモデルを定義します。

責務:

- run status の値を定義する。
- 永続化する run record を定義する。

このパッケージは storage、HTTP クライアント、runner の実装に依存してはいけません。

### `runstore`

orchestrator が所有する状態を SQLite に永続化します。

責務:

- orchestrator の database を開き、migrate する。
- run record を作成する。
- run status を更新する。
- runner イベントと workspace のメタデータを記録する。
- HTTP API 向けに、実行中の run と issue 単位の run 詳細を照会する。

このパッケージは domain type について `run` に依存します。issue-tracker API を呼び出したり、agent を実行したりしてはいけません。

### `tracker`

ローカル issue-tracker HTTP API を orchestrator 向けに適合させます。

責務:

- issue と issue の状態を取得する。
- issue-tracker が使う標準の API envelope を decode する。
- issue-tracker のエラー応答を、error code と message とともに表面化する。

このパッケージは orchestrator にとっての tracker 境界です。Tasq は現状ここでローカル issue-tracker API のみを使用し、Linear client は実装していません。

### `workflow`

repository の workflow contract を読み込みます。

責務:

- `WORKFLOW.md` を読み込む。
- サポートする front matter フィールドを parse する。
- orchestration の設定既定値を提供する。
- workflow ファイルからの相対パスで workspace root path を解決する。

このパッケージは configuration の parsing のみを所有します。worker を起動したり、workspace を作成したり、外部 API を呼び出したりしてはいけません。

### `workspace`

orchestrator の workspace を作成し、検証します。

責務:

- 設定済みの workspace root を管理する。
- サニタイズ済みの issue 単位 workspace directory を作成する。
- 設定済みの source から新規作成した workspace に内容を投入する。
- cleanup が要求されたとき、terminal、failed、cancelled の workspace を削除する。
- workspace path が設定済みの root から外れないようにする。

workspace への投入と cleanup の状態は、`runstore` のメタデータを通じて永続化されます。

### `runner`

agent runner の境界を定義します。

責務:

- runner interface を定義する。
- runner の task input と result output を定義する。
- テスト向けの simulated runner と、実際の実行向けの Codex app-server subprocess runner を提供する。

Codex runner は [../../docs/symphony/CODEX_APP_SERVER.md](../../docs/symphony/CODEX_APP_SERVER.md) に記載された contract に従います。

## 依存の方向

依存は小さな contract へ向かって内側に流し、外側への依存は coordinator からのみ許可します。

```text
cmd/orchestrator
  ├─ httpserver ── runstore ── run
  └─ workflow
```

ルール:

- `run` は依存を極力持たないままにします。
- `runstore`、`tracker`、`workflow`、`workspace`、`runner`、`httpserver` は、それぞれ自身の contract に集中させます。
- `cmd/orchestrator` は、幅広い root package の裏に construction を隠すのではなく、パッケージ同士を組み立てる役割に徹します。
