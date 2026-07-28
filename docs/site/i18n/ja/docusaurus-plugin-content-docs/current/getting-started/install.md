---
id: install
title: インストール
sidebar_position: 2
---

# インストール

Tasq をインストールしてローカルサービスを起動します。完了したら、[エージェントチュートリアル](pathname:///getting-started/agent-tutorial)へ進んでください。

## CLI をインストールする

インストール前に `TQ_HOME` を決めます。インストーラーは公開コマンドの `tq` を `PATH` 上へ配置し、非公開のサービスバイナリは選択した home 配下へ配置します。

```sh
curl -fsSLO https://raw.githubusercontent.com/version-1/tasq/main/scripts/install.sh
less install.sh
export TQ_HOME="${HOME}/.tasq"
TQ_HOME="$TQ_HOME" sh install.sh
export PATH="${HOME}/.local/bin:${PATH}"
tq version
```

実行前にインストーラーの内容を確認してください。デフォルトでは `tq` を `~/.local/bin` に、`issue-tracker`、`orchestrator`、`web` を `$TQ_HOME/system/bin` に配置します。サービスバイナリは `tq service start` が管理するため、直接実行は配布上サポートしません。

## ローカルサービスを起動する

Tasq はマシンローカルの実行時データとサービスバイナリを `TQ_HOME` 配下に保存します。インストール時に `TQ_HOME` を指定しなかった場合は `~/.tasq` を使います。サービス起動時にも同じ home を使用してください。

```sh
tq migrate
tq service start
tq service status
```

`tq service start` は issue-tracker、orchestrator、Web サーバーを固定のローカル loopback ポートで起動し、検出情報を `$TQ_HOME/system/state.json` に書き込みます。

| サービス | ポート |
| --- | ---: |
| issue-tracker | `37651` |
| orchestrator | `37652` |
| web | `37653` |

## Tasq を更新する

現在のバージョンを確認し、最新の正式リリースへ更新してから、新しいバージョンを確認します。

```sh
tq version
tq update
tq version
```

:::warning サービスの一時停止と自動マイグレーション

`tq update` はローカルサービスを一時停止し、リリースをインストールしてデータベースマイグレーションを適用した後、サービスを再起動します。ローカルサービスを中断できない間は実行しないでください。

:::

確認を省略する方法や、正式リリースまたはプレリリースのバージョンを指定する方法は、[Tasq を更新する](pathname:///guides/update-tasq)を参照してください。

## チュートリアルへ進む

CLI とローカルサービスの準備ができました。[エージェントチュートリアル](pathname:///getting-started/agent-tutorial)で、プロジェクトの登録、エージェント用ワークフローの準備、課題から pull request までの流れを体験してください。
