# デッドコード監査、2026-11-15

## 範囲と手法

本報告は Issue #45〜#56 と、#57 の領域横断再検査を統合したものです。`origin/main`
のコミット `81ad6df` で `git ls-files` が返した 628 ファイルを起点に、全追跡ファイルを
担当領域または理由付きの非実行資料へ分類し、公開入口、生成物、embed、設定、動的参照を
追跡しました。

### 再現可能な分類

実行可能な記録は [`scripts/audit/classify-dead-code-audit.sh`](../../scripts/audit/classify-dead-code-audit.sh)
です。`scripts/audit/classify-dead-code-audit.sh 81ad6df` を実行すると baseline TSV
（`path`、`owner`、`runtime/exclusion`）を出力し、未割当があれば失敗します。runtime file は path prefix により次のように
#45〜#56へ割り当てました。`internal/issue/domain`、`internal/issue/api`、
`cmd/issue-tracker` は #45、`internal/issue/store`、issue-tracker schema/migration は #46、
orchestrator の coordinator/startup と runstore/HTTP path は #47/#50、runner は #48、
workflow/workspace/tracker は #49、`internal/cli/tq` と `cmd/tq` は #51/#52、Issue frontend
feature は #53、残りの `cmd/web` と OpenAPI は #54、development/CI/release root は #55、
docs site/HTML は #56 です。`internal/config`、`internal/buildinfo`、`internal/migration`、
`cmd/tasq-tui`、`cmd/tqstory`、migrations embed は #52 に属します。

baseline output は runtime/configuration/assets が553件、exclusion が75件です。内訳は
documentation record が66件、plugin-metadata record が9件です。runtime owner 別の件数は
#45=18、#46=26、#47=7、#48=10、#49=14、#50=16、#51=20、#52=28、#53=135、#54=157、
#55=55、#56=67 です。exclusion は棚卸しから省いたものではありません。各項目について
reader、tool discovery、link、static publication の consumer を確認しました。first-match
script は input ごとにちょうど1行を出力し、baseline output には未割当と重複行がありません。

| 領域 | Issue | 検査結果 |
| --- | --- | --- |
| Issue Tracker の domain・HTTP API | #45 | 未使用の `IsActiveDependencyStatus` を削除。 |
| Issue Tracker の store・schema | #46 | 到達不能な nil slice fallback を2件削除。migration と schema の履歴は保持。 |
| Orchestrator の coordinator・起動 | #47 | 領域監査で検査。統合ベースに追加の確実な削除なし。 |
| runner・transport | #48 | 文書化済みの experimental WebSocket transport を保持。 |
| workflow・workspace・tracker 連携 | #49 | 冗長な workspace wrapper を削除。 |
| run model・永続化・HTTP API | #50 | 領域監査で検査。統合ベースに追加の確実な削除なし。 |
| CLI のリソース操作・API・出力 | #51 | 確実な削除なし。raw API と CLI 契約を保持。 |
| CLI 運用・TUI・設定・共通基盤 | #52 | 確実な削除なし。process、OS、update 経路を保持。 |
| Web UI の Issue 機能 | #53 | 廃止済み Issue UI コードを削除。 |
| Web UI 共通部・他画面・生成境界・Go server | #54 | 未使用 logo variant と生成 client alias を削除。 |
| 開発・CI・release・install | #55 | 最終的な削除なし。停止 script の候補はレビュー後に復元。 |
| docs site・実行 HTML | #56 | 未使用の site 依存と静的 HTML CSS を削除。 |

横断検査では残りの追跡ファイルも分類しました。リポジトリ/コンポーネントの指示、ADR・
設計/参照文書、plugin manifest と skill、静的ドキュメント媒体、license は非実行資料です。
利用者向け、ツール探索対象、またはリンクされる文書なので保持しています。SQL migration は
埋め込まれた後方互換の履歴入力であり、現在の経路が古い列名を参照しないことだけでは削除
できません。生成 API client と公開 OpenAPI 文書も、生成元と外部利用者の契約を維持します。

## 保持した境界ケース

- 公開 HTTP/CLI endpoint、JSON/YAML field、migration 履歴、embed、platform 固有の
  release/install 経路は互換性契約として保持しました。
- runner の WebSocket package は transport-neutral の意図的なサポートです。runtime 選択と
  real server integration は `docs/symphony/DEVIATIONS.md` で保留と明記されています。
- frontend の動的参照には、Vite 環境変数で制御する MSW、Docusaurus の自動探索、MDX、
  Storybook、static URL、生成 client が含まれます。
- failure、cancel、cleanup、timeout、process identity の経路は、到達不能が証明できない限り
  保持しました。

## 統合検証

- Host Go `go1.27.3 darwin/arm64`: `go test ./...`、`go vet ./...`、
  `go build ./cmd/...` が成功しました。初回 test は `cmd/web` が `frontend/dist` を embed
  するため frontend build 前だけ停止し、`dist` 生成後に全て成功しました。
- Dev-container Node `22.23.4`: `cmd/web/frontend` で `npm test`（29 files、164 tests）、
  `npm run typecheck`、`npm run build` が成功しました。
- Dev-container Node `22.23.4`: `docs/site` で全 locale の `npm run build`、個別の
  `npm run build -- --locale ja` が成功しました。
- `git diff --check` が成功しました。サービス停止、release、installer、live update は検証として
  実行していません。

API generator も実行しました。OpenAPI 入力に差分がないにもかかわらず、導入済み generator は
`orchestrator.ts` に空白だけの差分を出しました。この差分は commit せず、生成物は追跡済みの
canonical form を維持しています。これは contract 変更ではなく generator の formatting
non-idempotence として記録します。

## 限界

リポジトリ全体のテキスト参照は候補を見つける根拠であり、runtime 到達性の証明ではありません。
削除は未使用または到達不能を確認できたコード・asset に限定しました。外部からの直接 URL と、
#54 では repository 内の利用がない logo variant を削除しました。一方 #56 の
`agent-task-queue.svg` は外部利用を否定できない public static URL のため保持しました。将来
文書化された protocol extension は意図的に保持しています。
