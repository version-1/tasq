# リリースバイナリ起動メモ

English counterpart: [release-binary-startup.md](release-binary-startup.md).

このメモは、README の Getting Started に載せるバイナリ単体起動フローの前提を記録する。

リリースアーカイブに `tq`、`issue-tracker`、`orchestrator`、`web` が含まれ、インストーラーがサービス実行ファイルを `TQ_HOME` 配下へ配置し、ローカルデータベースへ migration が適用済みであれば、Docker なしでフル体験を再現できる。

## リリース成果物

GoReleaser は各リリースアーカイブに次のバイナリを含める。

- `tq`
- `issue-tracker`
- `orchestrator`
- `web`

Release workflow は GoReleaser の前に `cmd/web/frontend/dist` をビルドする必要がある。`web` バイナリは Go の `embed` でそのディレクトリを埋め込むため、ダウンロード済みのリリースアーカイブを実行するだけなら Node.js や frontend files は不要。

## ランタイム状態

Tasq はマシンローカルの runtime data を `TQ_HOME` 配下に保存する。`TQ_HOME` が未設定の場合は `~/.tasq` が使われる。

関係するファイルは次のとおり。

- `$TQ_HOME/WORKFLOW.md`: 初回利用時に default workflow template で作成される。
- `$TQ_HOME/system/state.json`: 起動中サービスが書き込む service discovery state。
- `$TQ_HOME/system/data/issues.sqlite`: issue-tracker database。
- `$TQ_HOME/system/data/orchestrator.sqlite`: orchestrator database。
- `$TQ_HOME/system/log/*.log`: `tq service start` が書き込む logs。
- `$TQ_HOME/system/bin/{issue-tracker,orchestrator,web}`: `tq service start` が起動する private binaries。

新規 database では、サービス起動前に migration が必要。

```sh
tq migrate
```

この手順がない場合、`issue-tracker` は pending migration error で終了し、`tq migrate` の実行を促す。

## 管理起動

バイナリだけでフル体験する最短手順は次のとおり。

```sh
export TQ_HOME="${HOME}/.tasq"
tq migrate
tq service start
tq project add --key my-project /path/to/project
tq issue create --project my-project --title "Try Tasq from binaries"
```

Web UI は、次のコマンドで表示される Web port を使って `http://127.0.0.1:<web-port>` を開く。

```sh
tq service status
```

または次のコマンドを使う。

```sh
tq web
```

`tq service start` は 3 つの managed service を固定の loopback port で起動する。port 表は [インストール](../site/i18n/ja/docusaurus-plugin-content-docs/current/getting-started/install.md)を参照。各 service の data:

| Service | Data |
| --- | --- |
| `issue-tracker` | `$TQ_HOME/system/data/issues.sqlite` |
| `orchestrator` | `$TQ_HOME/system/data/orchestrator.sqlite` |
| `web` | `web` バイナリに埋め込まれた static assets |

`tq service start` は `$TQ_HOME/system/bin` の `issue-tracker`、`orchestrator`、`web` だけを起動する。実行中の `tq` の隣、`PATH`、source tree は探索しない。管理対象の binary が欠落または実行不可の場合は、サービスを起動する前にすべての不正な path を報告し、同じ `TQ_HOME` を指定した再インストールを案内する。

`tq service start` には custom port 用の flags はない。default port が使用中の場合、README では競合プロセスの停止、または手動起動を案内する。

## 手動起動

手動起動は custom port が必要な場合や、各バイナリの flags を説明する場合に有用。developer workflow であり、サービス binary の直接実行は配布上サポートしない。

```sh
export TQ_HOME="$(pwd)/.tasq"
tq migrate

issue-tracker -addr 127.0.0.1:37651
orchestrator -issue-tracker http://127.0.0.1:37651 -port 37652
web -addr 127.0.0.1:37653 \
  -tracker-url http://127.0.0.1:37651 \
  -orchestrator-url http://127.0.0.1:37652
```

各 flags は次のとおり。

| Binary | Flags | Defaults and behavior |
| --- | --- | --- |
| `issue-tracker` | `-addr`, `-db` | `-addr` は default で `:37651`。`-db` は default で `$TQ_HOME/system/data/issues.sqlite`。`state.json` に `issue_tracker` を書く。 |
| `orchestrator` | `-db`, `-issue-tracker`, `-port` | `-db` は default で `$TQ_HOME/system/data/orchestrator.sqlite`。`state.json` に issue-tracker address がある場合、`-issue-tracker` は省略できる。`-port -1` では workflow config が有効化しない限り HTTP を無効にする。`state.json` に `orchestrator` を書く。 |
| `web` | `-addr`, `-tracker-url`, `-orchestrator-url` | `-addr` は default で `:37653`。backend URLs は default で `http://127.0.0.1:37651` と `http://127.0.0.1:37652`。`state.json` に `web` を書く。 |
| `tq` | `--api-url`, `--output` | API URL の解決順は `--api-url` または `-api-url`、`TQ_API_URL`、`$TQ_HOME/system/state.json`、`http://localhost:37651`。 |

default 以外の port を使う場合は、`web` に `-tracker-url` と `-orchestrator-url` を渡す。`tq` や `orchestrator` と違い、`web` は backend URLs を `state.json` から discovery しない。

## 実施した検証

`.tmp/issue-161` に独立した `TQ_HOME` を作り、random loopback ports で binary-only flow を確認した。

1. `npm run build` で frontend をビルドし、その後 `tq`、`issue-tracker`、`orchestrator`、`web` を local binaries としてビルドした。
2. migration 前に `issue-tracker` を起動すると、pending migration 名と `tq migrate` の案内を出して失敗することを確認した。
3. `tq migrate` を実行し、`$TQ_HOME/system/data` に 2 つの SQLite databases が作成されることを確認した。
4. `issue-tracker -addr 127.0.0.1:0` を起動し、`state.json` に `issue_tracker.addr` と database path が書かれることを確認した。
5. `--api-url` なしで `tq project add --key issue-161-check .tmp/issue-161/project` と `tq issue create --project issue-161-check --title "Binary startup check"` を実行し、`tq` が `state.json` から API URL を解決することを確認した。
6. `-issue-tracker` なしで `orchestrator -port 0` を起動し、`state.json` から issue-tracker URL を解決して自身の state を書くことを確認した。
7. 明示的な backend URLs を渡して `web -addr 127.0.0.1:0` を起動した。
8. `/` で埋め込み SPA が返り、`/tracker/api/v1/issues` で Web proxy 経由で作成済み issue が返ることを確認した。

## README の判断ポイント

README では、次の制約付きで binary full experience をサポート済みとして案内できる。

- `tq service start` の前に `tq migrate` を入れる。
- インストーラーが `tq` を公開 install directory に、3サービスを `$TQ_HOME/system/bin` に配置すると書く。
- default ports を明記する（port 表は [インストール](../site/i18n/ja/docusaurus-plugin-content-docs/current/getting-started/install.md)を参照）。
- custom ports が必要な場合は手動 service startup が必要だと説明する。
- ダウンロード済み release の runtime requirements には Node.js を含めない。`web` は built frontend を埋め込んでいるため。
