# ADR-0001: Project Path をホストローカルパスとして保存する

## Context

Tasq は issue-tracker API を通じて project と workspace を保存します。ローカル開発では、issue-tracker が Docker Compose で動き、`tq` はホスト上または Compose のヘルパーコンテナ内で実行される場合があります。

`tq` と issue-tracker の両方を Compose 内で実行すると、リポジトリは `/workspace` として見えます。このパスはコンテナにとっては有効ですが、ユーザーの実際のローカルプロジェクトパスではありません。project record に `/workspace` を永続化すると、開発実行時の詳細がユーザーデータに漏れ、すべての project を同じコンテナファイルシステムにマウントしない限り、複数のローカル project の登録にも対応できません。

製品モデルでは、project はユーザーのマシン上にあるリポジトリを表します。そのため、永続化する project location は `/Users/admin/Projects/Private/tasq` のように、ユーザーが選択したホストローカルの絶対パスであるべきです。

## Decision

`Project.Location` と `Workspace.Path` はホストローカルの絶対パスとして永続化します。

`tq project add <path>` コマンドはクライアントホスト上でパスを解決し、ローカルに存在することを確認してから、ホストローカルの絶対パスを issue-tracker API に送信します。

issue-tracker API はパスの形だけを検証します。絶対パスであることと長さ制限は要求しますが、サーバーファイルシステムを使ったディレクトリ存在確認は行いません。API サーバーはコンテナ内や、クライアントのホストパスを参照できない別の実行時で動く可能性があります。

Runner やコンテナ連携は、分離された環境で作業を実行する必要がある場合に限って、ホストローカルパスを実行時パスへ変換します。この対応付けは project record ではなく、runner/container adapter 境界の責務です。

## Alternatives

### コンテナパスを保存する

`/workspace` を保存する方法は、`go-tools` と issue-tracker が同じ bind mount を共有する単一 project の Docker Compose 開発環境では動作します。

しかし、これは環境固有の実行時パスを保存するため、永続モデルとしては採用しません。任意のホスト project を登録するには、その project も issue-tracker コンテナにマウントする必要があります。

### ホストパスと同じ絶対パスにマウントする

Docker Compose でホストディレクトリをコンテナ内の同じ絶対パスにマウントする方法もあります。たとえば `/Users/admin/Projects:/Users/admin/Projects` です。

この方法ならサーバー側の存在確認を通せる場合がありますが、マシン固有で、OS や CI をまたぐと壊れやすく、issue-tracker の検証を特定の実行時レイアウトに結合してしまいます。

### すべてをホスト上で実行する

issue-tracker と `tq` の両方をホスト上で実行すれば、サーバー側のパス存在確認は自然に機能します。

ただし、これはサポート対象の Compose 開発環境をカバーできず、API の挙動がデプロイ構成に依存してしまいます。

## Consequences

Project record はコンテナから見たパスではなく、ユーザーの project location を表すため、実行バックエンドをまたいでも扱いやすくなります。

API はホストファイルシステムに直接アクセスできない場合でも、有効なホストパスを受け付けられます。

対象ファイルシステムを参照できるクライアントは、record 作成前の存在確認を担当します。`tq project add` はこの確認を行います。

コンテナ内で実行される実行時コンポーネントには、明示的なパス対応付け手順が必要です。たとえば `/Users/admin/Projects/Private/tasq` を `/workspace` に対応付ける、または project path ごとに runner コンテナへマウントする必要があります。

`tq` を経由しない API consumer は、構文上有効な絶対パスであれば存在しないパスの record も作成できます。issue-tracker はクライアントホストのファイルシステム状態を確実には検証できないため、これは許容します。

## Notes

`make tq` はホスト上で `tq` を実行し、Compose issue-tracker には割り当てられた localhost port 経由で接続します。これにより、パス解決はユーザーのファイルシステムと揃いながら、Compose API service も利用できます。

この検証境界は `docs/design/schema.md` にも記載します。
