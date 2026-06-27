# 課題状態とキュー状態

このドキュメントでは、`issue.status`、派生値の `queueStatus`、想定される課題ワークフローを定義します。フィールド単位のスキーマ詳細は [schema.ja.md](schema.ja.md) を参照してください。API レスポンスの形は [api.ja.md](api.ja.md) を参照してください。

## 所有者

issue-tracker は `issue.status`、依存関係の辺、派生的なキュー投入可否を所有します。`queueStatus` はデータベースに保存しません。`GET /api/v1/summary` で各課題サマリーを返すときに導出します。

API は `issue.status` が許可された enum 値のいずれかであることを検証します。一方で、すべての更新に厳密な状態機械を強制するわけではありません。CLI のショートカットや orchestrator のフローは下の標準遷移に従いますが、`PATCH /api/v1/issues/{id}` は任意の有効な status に更新できます。

## `issue.status`

| Status | 意味 | キュー / 依存関係での役割 |
| --- | --- | --- |
| `backlog` | まだ割り当て対象ではない下書きまたは計画中の作業。 | アクティブな依存状態。summary の `queueStatus` は `backlog`。 |
| `ready` | 割り当て候補にできる作業。 | アクティブな依存状態。summary の `queueStatus` は依存関係に応じて `pending` または `queued`。 |
| `in_progress` | すでに着手され、処理中の作業。 | アクティブな依存状態。summary の `queueStatus` は `processing`。 |
| `review` | 人によるレビューまたは最終確認を待っている作業。 | アクティブな依存状態。後続作業はレビュー完了を待つべきです。summary の `queueStatus` は `done`。 |
| `blocked` | 外部入力またはブロッカーの解消が必要で、進められない作業。 | 満たされた依存状態。summary の `queueStatus` は `done`。 |
| `failed` | 失敗として終了した作業。 | 満たされた依存状態。summary の `queueStatus` は `done`。 |
| `cancelled` | 意図的に停止され、続行しない作業。 | 満たされた依存状態。summary の `queueStatus` は `done`。 |
| `duplicate` | 別の課題で表現されている重複作業。 | 満たされた依存状態。summary の `queueStatus` は `done`。 |
| `done` | 完了した作業。 | 満たされた依存状態。summary の `queueStatus` は `done`。 |

アクティブな依存状態は `backlog`、`ready`、`in_progress`、`review` です。

満たされた依存状態は `done`、`cancelled`、`duplicate`、`failed`、`blocked` です。

## `queueStatus`

`queueStatus` は `GET /api/v1/summary` の `IssueSummary` に含まれる、課題単位の派生状態です。UI と TUI で表示するための値であり、永続化された真実の状態として扱ってはいけません。

| `queueStatus` | 導出条件 |
| --- | --- |
| `backlog` | `issue.status` が `backlog`。 |
| `pending` | `issue.status` が `ready` で、依存先にアクティブな課題が 1 件以上ある。 |
| `queued` | `issue.status` が `ready` で、アクティブな依存先がない。 |
| `processing` | `issue.status` が `in_progress`。 |
| `done` | `issue.status` がキュー処理の対象外。対象は `review`、`done`、`blocked`、`failed`、`cancelled`、`duplicate`。 |

`GET /api/v1/queue` の意味はより狭く、`ready` の課題だけを `queued` と `pending` の配列に分けて返します。一方で、summary の `queueStatus` はボード上のすべての課題を対象にします。

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

## Orchestrator に関する注意

orchestrator は `GET /api/v1/queue` を読み、そのエンドポイントの `queued` 配列に含まれる課題だけを割り当てます。依存関係の解決は issue-tracker 側に残し、orchestrator worker がキュー投入可否のロジックを重複実装しないようにします。

runner や承認の失敗によって人の対応が必要になった場合、orchestrator のフローは実行可能な課題を `blocked` に移し、blocker comment を追加することがあります。ブロッカーが解消されたら、運用者は課題を `ready` に戻せます。
