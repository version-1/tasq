# 課題状態とキュー状態

このドキュメントでは、`issue.status`、システム全体のキューから見た `queueStatus`、想定される課題ワークフローを定義します。フィールド単位のスキーマ詳細は [schema.ja.md](schema.ja.md) を参照してください。API レスポンスの形は [api.ja.md](api.ja.md) を参照してください。

## 所有者

issue-tracker は `issue.status`、依存関係の辺、システム全体のキュー投入可否を所有します。`queueStatus` は、そのキューから見た課題の状態です。データベースには保存せず、`GET /api/v1/summary` は各課題の現在の `queueStatus` を返します。

API は `issue.status` が許可された enum 値のいずれかであることを検証します。一方で、すべての更新に厳密な状態機械を強制するわけではありません。CLI のショートカットや orchestrator のフローは下の標準遷移に従いますが、`PATCH /api/v1/issues/{id}` は任意の有効な status に更新できます。

## `issue.status`

| Status | 意味 | キュー / 依存関係での役割 |
| --- | --- | --- |
| `backlog` | まだ割り当て対象ではない下書きまたは計画中の作業。 | アクティブな依存状態。`queueStatus` は `backlog`。 |
| `ready` | 割り当て候補にできる作業。 | アクティブな依存状態。`queueStatus` は依存関係に応じて `pending` または `queued`。 |
| `in_progress` | すでに着手され、処理中の作業。 | アクティブな依存状態。`queueStatus` は `processing`。 |
| `review` | 人によるレビューまたは最終確認を待っている作業。 | アクティブな依存状態。後続作業はレビュー完了を待つべきです。`queueStatus` は `inactive`。 |
| `blocked` | 外部入力またはブロッカーの解消が必要で、進められない作業。 | 依存先としては後続をブロックする状態。`queueStatus` は `inactive`。 |
| `failed` | 失敗として終了した作業。 | 依存先としては後続をブロックする状態。`queueStatus` は `inactive`。 |
| `cancelled` | 意図的に停止され、続行しない作業。 | 満たされた依存状態。`queueStatus` は `inactive`。 |
| `duplicate` | 別の課題で表現されている重複作業。 | 満たされた依存状態。`queueStatus` は `inactive`。 |
| `done` | 完了した作業。 | 満たされた依存状態。`queueStatus` は `completed`。 |

アクティブな依存状態は `backlog`、`ready`、`in_progress`、`review` です。

満たされた依存状態は `done`、`cancelled`、`duplicate` です。`blocked` と `failed` を含むそれ以外の依存先 status は、依存している `ready` の課題を `pending` に残します。

## `queueStatus`

`queueStatus` は、システム全体のキューから見た課題単位の状態です。`GET /api/v1/summary` は各 `IssueSummary` に現在の値を含めますが、summary がキューモデルを所有しているわけではありません。

| `queueStatus` | 導出条件 |
| --- | --- |
| `backlog` | `issue.status` が `backlog`。 |
| `pending` | `issue.status` が `ready` で、依存先にアクティブな課題が 1 件以上ある。 |
| `queued` | `issue.status` が `ready` で、アクティブな依存先がない。 |
| `processing` | `issue.status` が `in_progress`。 |
| `completed` | `issue.status` が `done`。 |
| `inactive` | `issue.status` がキュー処理の対象外。対象は `review`、`blocked`、`failed`、`cancelled`、`duplicate`。 |

`GET /api/v1/queue` は割り当て用のキュー view で、`ready` の課題だけを `queued` と `pending` の配列に分けて返します。`queueStatus` は同じキューモデルを使い、summary の利用者がキュールールを再実装せずにボード上のすべての課題を表示できるよう分類します。

## 標準遷移

これらは想定されるワークフロー上の遷移です。API が許可する更新の完全な一覧ではありません。

```mermaid
stateDiagram-v2
  [*] --> backlog
  backlog --> ready
  ready --> in_progress
  in_progress --> review
  review --> done
  in_progress --> blocked
  ready --> blocked
  blocked --> ready
  blocked --> in_progress
  in_progress --> failed
  failed --> backlog
  ready --> cancelled
  in_progress --> cancelled
  backlog --> cancelled
  ready --> duplicate
  backlog --> duplicate
```

主な CLI ショートカットは、次の status 更新に対応します。

| Command | 結果 |
| --- | --- |
| `tq issue draft <id>` | `issue.status` を `backlog` にします。 |
| `tq issue ready <id>` | `issue.status` を `ready` にします。 |
| `tq issue cancel <id>` | `issue.status` を `cancelled` にします。 |
| `tq issue close <id>` | `issue.status` を `done` にします。 |
| `tq issue update <id> --status <status>` | 任意の有効な `issue.status` にします。 |

## 課題ごとの Auto OK

各課題カードの Auto OK スイッチは、初期状態が OFF です。設定と送信成功回数は
issue-tracker に保存し、ブラウザの再読み込みやサービス再起動後も維持します。
OFF から ON に切り替えると回数をリセットします。ON の再設定ではリセットしません。

ON の課題が `blocked` になると、理由を問わず、自動返信の予約と `ready` への変更を
同じトランザクションで行います。すでに `blocked` の課題で ON にした場合も即座に
予約します。通常のキューと同様、依存関係の解消、実行枠の確保、実行中の試行の終了を
待ってから再開します。ブラウザを閉じても処理は続きます。

直前の保存済みスレッドを再開し、1 ターンの本文として `ok` だけを送ります。
ワークフロー、継続指示、変更要求の指示は追加しません。未処理の変更要求は、後の
通常の再開で扱います。`turn/start` が受理された時点で回数を加算し、まだ `ready` なら
`in_progress` に変更します。成功ごとに `tasq-auto-ok` の進捗コメントと
`auto_ok_reply` の実行イベントを記録します。5 回目の受理時点で、停止理由を `limit`
として自動 OFF にします。5 回目のターンは実行し、その後の `blocked` は停止したまま
にします。作業の進展や再起動では回数をリセットしません。

ワークスペース準備、再開スレッドの欠如、起動、返信の受理に失敗した場合は、停止理由を
`error` として自動 OFF にし、`blocked` に戻して理由をブロッカーコメントに残します。
失敗は回数に含めません。受理後のエラーは通常の実行失敗として扱い、次の自動返信の
対象になります。割り当て前に OFF にすると未送信の予約を取り消して `blocked` に
戻します。OFF の後に ON にすると回数をリセットして新しい世代を作るため、旧世代の
受理通知は新しい cycle に影響しません。

予約には永続化した世代番号を付けます。オーケストレーターは `turn/start` の直前に
`POST /api/v1/issues/{id}/auto-ok-result` の `phase: "claim"` で予約を claim します。
その後の受理通知で履歴・回数・停止状態を同じトランザクションに保存します。claim 後は
ライフサイクル更新で受理通知を破棄しません。先に再度 `blocked` になっても受理を回数に
含め、5 回目までは次世代を予約します。重複通知や古い世代の通知では状態を変更しません。
ただし app-server の受理直後、結果の保存前にプロセスが異常終了する場合まで含めた、
分散処理全体での厳密な 1 回だけの送信は保証しません。

Auto OK が送るのは会話の入力です。コマンド実行・ファイル変更の承認要求には、引き続き
[app-server の承認方針](../symphony/CODEX_APP_SERVER.ja.md)を適用します。

## Orchestrator に関する注意

orchestrator は `GET /api/v1/queue` を読み、そのエンドポイントの `queued` 配列に含まれる課題だけを割り当てます。依存関係の解決は issue-tracker 側に残し、orchestrator worker がキュー投入可否のロジックを重複実装しないようにします。

runner や承認の失敗によって人の対応が必要になった場合、orchestrator のフローは実行可能な課題を `blocked` に移し、blocker comment を追加することがあります。ブロッカーが解消されたら、運用者は課題を `ready` に戻せます。
