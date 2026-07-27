---
id: tq-output-stories
title: tq コマンド出力を確認する
sidebar_position: 2
---

# tq コマンド出力を確認する

`tqstory` を使うと、サービスの起動、データベースの作成、issue-tracker API の呼び出しをせずに、`tq` コマンドの出力を確認できます。各シナリオは固定データを使い、CLI と同じ出力レンダラーを呼び出します。

リポジトリのルートからシナリオを実行します。

```sh
go run ./cmd/tqstory <scenario>
```

Go module では、`go run` に `./` を付けてパッケージを指定する必要があります。

## シナリオ一覧

| シナリオ | 確認できる出力 |
| --- | --- |
| `tq_issue_list` | project と status に色を付けた issue 一覧。 |
| `tq_issue_detail` | issue の詳細フィールド。 |
| `tq_issue_action` | 成功した issue 操作と、その後の issue 詳細。 |
| `tq_project_list` | project の表と識別子の色。 |
| `tq_empty` | project 一覧が空の場合のメッセージ。 |
| `tq_project_check` | PASS と FAIL の project チェック。 |
| `tq_service_status` | running と stopped の service state。 |
| `tq_migration_status` | applied と pending の migration state。 |
| `tq_service_start_success` | service start 成功時の結果。 |
| `tq_service_stop_success` | service stop 成功時の結果。 |
| `tq_project_remove_success` | project 削除成功時の結果。 |
| `tq_workflow_remove_success` | workflow override 削除成功時の結果。 |
| `tq_project_remove_cancelled` | project 削除をキャンセルしたときの結果。 |
| `tq_project_remove_confirmation` | project 削除の警告と確認プロンプト。 |
| `tq_warning` | 破壊的操作の警告。 |
| `tq_service_start_fail` | text mode の service start 失敗。 |
| `tq_json_success` | ANSI エスケープシーケンスを含まない成功 JSON。 |
| `tq_json_error` | ANSI エスケープシーケンスを含まない JSON エラー。 |
| `all` | 各結果の前に見出しを付けた全シナリオ。 |

## コマンド結果を確認する

成功、キャンセル、確認表示を確認するには、次のシナリオを使います。

```sh
go run ./cmd/tqstory tq_service_start_success
go run ./cmd/tqstory tq_service_stop_success
go run ./cmd/tqstory tq_project_remove_success
go run ./cmd/tqstory tq_workflow_remove_success
go run ./cmd/tqstory tq_project_remove_cancelled
go run ./cmd/tqstory tq_project_remove_confirmation
```

確認シナリオはプロンプトを描画するだけで、入力の読み取りや project の削除は行いません。

## 一覧・状態・エラーを確認する

```sh
go run ./cmd/tqstory tq_issue_list
go run ./cmd/tqstory tq_service_status
go run ./cmd/tqstory tq_migration_status
go run ./cmd/tqstory tq_service_start_fail
go run ./cmd/tqstory tq_json_error
```

失敗シナリオは見た目を確認するためのプレビューで、終了ステータス `0` を返します。手動確認やスクリーンショットの取得に利用できます。JSON シナリオには ANSI エスケープシーケンスを含みません。

## すべてのシナリオを確認する

各結果の前に見出しを付けて、すべてのシナリオを実行します。

```sh
go run ./cmd/tqstory all
```

存在しないシナリオまたはシナリオなしで実行すると、利用可能なシナリオ名を表示し、終了ステータス `2` を返します。
