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
