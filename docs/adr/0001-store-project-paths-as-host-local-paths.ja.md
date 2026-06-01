# ADR-0001: Project Path を Host-Local Path として保存する

## Context

Tasq は issue-tracker API を通じて project と workspace を保存します。ローカル開発では、issue-tracker が Docker Compose で動き、`tq` はホスト上または Compose の helper container 内で実行される場合があります。

`tq` と issue-tracker の両方を Compose 内で実行すると、repository は `/workspace` として見えます。この path は container にとっては有効ですが、ユーザーの実際の local project path ではありません。project record に `/workspace` を永続化すると、開発 runtime の詳細が user data に漏れ、すべての project を同じ container filesystem に mount しない限り複数 local project の登録にも対応できません。

製品モデルでは、project はユーザーのマシン上にある repository を表します。そのため、永続化する project location は `/Users/admin/Projects/Private/tasq` のように、ユーザーが選択した host-local absolute path であるべきです。

## Decision

`Project.Location` と `Workspace.Path` は host-local absolute path として永続化します。

`tq project add <path>` command は client host 上で path を解決し、local に存在することを確認してから、host-local absolute path を issue-tracker API に送信します。

issue-tracker API は path の形だけを検証します。absolute path であることと length limit は要求しますが、server filesystem を使った directory existence check は行いません。API server は container や、client の host path を参照できない別 runtime で動く可能性があります。

Runner や container integration は、isolated environment で作業を実行する必要がある場合に限って host-local path を runtime path へ変換します。この mapping は project record ではなく、runner/container adapter boundary の責務です。

## Alternatives

### Container Path を保存する

`/workspace` を保存する方法は、`go-tools` と issue-tracker が同じ bind mount を共有する単一 project の Docker Compose 開発環境では動作します。

しかし、これは environment-specific な runtime path を保存するため、durable model としては採用しません。任意の host project を登録するには、その project も issue-tracker container に mount する必要があります。

### Host Path と同じ Absolute Path に Mount する

Docker Compose で host directory を container 内の同じ absolute path に mount する方法もあります。たとえば `/Users/admin/Projects:/Users/admin/Projects` です。

この方法なら server-side existence check を通せる場合がありますが、machine-specific で、OS や CI をまたぐと壊れやすく、issue-tracker validation を特定の runtime layout に結合してしまいます。

### すべてを Host 上で実行する

issue-tracker と `tq` の両方を host 上で実行すれば、server-side path existence check は自然に機能します。

ただし、これは supported な Compose development environment をカバーできず、API behavior が deployment topology に依存してしまいます。

## Consequences

Project record は container から見た path ではなく、ユーザーの project location を表すため、execution backend をまたいでも扱いやすくなります。

API は host filesystem に直接アクセスできない場合でも、有効な host path を受け付けられます。

target filesystem を参照できる client は、record 作成前の existence check を担当します。`tq project add` はこの check を行います。

container 内で実行される runtime component には、明示的な path mapping step が必要です。たとえば `/Users/admin/Projects/Private/tasq` を `/workspace` に mapping する、または project path ごとに runner container へ mount する必要があります。

`tq` を経由しない API consumer は、syntactically valid な absolute path であれば存在しない path の record も作成できます。issue-tracker は client-host filesystem state を確実には検証できないため、これは許容します。

## Notes

`make tq` は host 上で `tq` を実行し、Compose issue-tracker には割り当てられた localhost port 経由で接続します。これにより、path resolution はユーザーの filesystem と揃いながら、Compose API service も利用できます。

この validation boundary は `docs/schema.md` にも記載します。
