# tasq

AI コーディングエージェント向けタスクマネージャー。

tasq は、実装作業を見えるキューにし、そのキューのためのローカルサービスを起動し、`tq` CLI と Web UI の両方から進捗を確認できるようにするツールです。

[![Tasq の紹介動画を見る](docs/site/static/img/tasq-thumbnail.png)](https://github.com/user-attachments/assets/8c4fdc9c-c70b-4f86-8e0a-323f8880ffb7)

English counterpart: [README.md](README.md).

## Tasq を使う理由

AI エージェントによる並列作業では、コードを書くことよりも、タスク・ワークスペース・レビューの調整がボトルネックになります。Tasq は、課題キュー、分離されたワークスペース、ローカルサービス、CLI と Web UI によって、その作業を見える状態に保ちます。

課題とワークフロー、製品全体の説明は [Tasq ドキュメントサイト](https://version-1.github.io/tasq/ja/)を参照してください。

## Features

- エージェント向けサイズのタスク、優先度、依存関係、コメントを扱う課題キュー。
- タスク作成、状態更新、進捗コメント追加、ワークフローのスクリプト化に使う `tq` CLI。
- SQLite を使うローカルの issue-tracker、orchestrator、Web UI サービス。
- プロジェクト、課題、コメント、キューの状態、サービスの状態を確認する Web UI。
- 課題を実際のローカルリポジトリのパスに結び付けるプロジェクト登録。
- バイナリのみのローカル構成に必要な実行時バイナリを含むリリースアーカイブ。

## Install

確認したインストーラーで、最新の正式リリースアーカイブをインストールします。

```sh
curl -fsSLO https://raw.githubusercontent.com/version-1/tasq/main/scripts/install.sh
less install.sh
sh install.sh
```

インストール先、`TQ_HOME`、サービスの起動、更新については、[インストール](https://version-1.github.io/tasq/ja/getting-started/install)を参照してください。

## Documentation

チュートリアル、ガイド、概念、リファレンスは、[Tasq ドキュメントサイト](https://version-1.github.io/tasq/ja/)を参照してください。
